package main

import (
	"net"
	"slices"
	"strconv"
	"time"
)

var storage map[string]storage_obj = make(map[string]storage_obj)

const (
	GET = iota
	SET
	RPUSH
	LPUSH
	LRANGE
	LLEN
	LPOP
	BLPOP
)

type storage_cmd struct {
	cmd       int
	key       string
	value     any
	to        net.Conn
	timestamp time.Time
	expiry    time.Duration
}
type storage_obj struct {
	value     any
	timestamp time.Time
	expiry    time.Duration
}

func handleStorage(cmds chan storage_cmd) {

	blocked_pops := make(map[string][]storage_cmd)

	for cmd := range cmds {
		switch cmd.cmd {
		case GET:
			val, ok := storage[cmd.key]
			if !ok {
				cmd.to.Write([]byte(NULL_BULK_STRING))
			} else if val.expiry > 0 && val.timestamp.Add(val.expiry).Compare(cmd.timestamp) != 1 {
				delete(storage, cmd.key)
				cmd.to.Write([]byte(NULL_BULK_STRING))
			} else {
				str, _ := encodeObj(val.value)
				cmd.to.Write([]byte(str))
			}
		case SET:
			storage[cmd.key] = storage_obj{
				value:     cmd.value,
				timestamp: cmd.timestamp,
				expiry:    cmd.expiry,
			}
			cmd.to.Write([]byte(encodeSimpleString("OK")))
		case LPUSH:
			val, ok := storage[cmd.key]
			if !ok {
				obj := []any{}
				storage[cmd.key] = storage_obj{
					value:     obj,
					timestamp: cmd.timestamp,
				}
			}
			val, ok = storage[cmd.key]
			if arr, ok := val.value.([]any); ok {
				arg_arr := cmd.value.([]any)
				slices.Reverse(arg_arr)
				arr = append(arg_arr, arr...)

				obj, _ := encodeObj(len(arr))
				val.value = arr
				storage[cmd.key] = val
				cmd.to.Write([]byte(obj))

				for len(blocked_pops[cmd.key]) > 0 && len(storage[cmd.key].value.([]any)) > 0 {
					b_cmd := blocked_pops[cmd.key][0]
					blocked_pops[cmd.key] = blocked_pops[cmd.key][1:]

					s_val := storage[cmd.key]
					s_arr := s_val.value.([]any)
					poppedValue := s_arr[0]
					s_val.value = s_arr[1:]

					if len(s_val.value.([]any)) == 0 {
						delete(storage, cmd.key)
					} else {
						storage[cmd.key] = s_val
					}

					resp, _ := encodeObj([]any{b_cmd.key, poppedValue})
					b_cmd.to.Write([]byte(resp))
				}
			} else {
				cmd.to.Write([]byte(encodeSimpleString("TYPE ERROR")))
			}
		case RPUSH:
			val, ok := storage[cmd.key]
			if !ok {
				obj := []any{}
				storage[cmd.key] = storage_obj{
					value:     obj,
					timestamp: cmd.timestamp,
				}
			}
			val, ok = storage[cmd.key]
			if arr, ok := val.value.([]any); ok {
				arr = append(arr, cmd.value.([]any)...)
				obj, _ := encodeObj(len(arr))
				val.value = arr
				storage[cmd.key] = val
				cmd.to.Write([]byte(obj))

				for len(blocked_pops[cmd.key]) > 0 && len(storage[cmd.key].value.([]any)) > 0 {
					b_cmd := blocked_pops[cmd.key][0]
					blocked_pops[cmd.key] = blocked_pops[cmd.key][1:]

					s_val := storage[cmd.key]
					s_arr := s_val.value.([]any)
					poppedValue := s_arr[0]
					s_val.value = s_arr[1:]

					if len(s_val.value.([]any)) == 0 {
						delete(storage, cmd.key)
					} else {
						storage[cmd.key] = s_val
					}

					resp, _ := encodeObj([]any{b_cmd.key, poppedValue})
					b_cmd.to.Write([]byte(resp))
				}
			} else {
				cmd.to.Write([]byte(encodeSimpleString("TYPE ERROR")))
			}
		case LPOP:
			val, ok := storage[cmd.key]
			if !ok {
				cmd.to.Write([]byte(NULL_BULK_STRING))
			} else {
				optional := false
				popped := 1
				if arr, ok := cmd.value.([]any); ok {
					if len(arr) == 1 {
						optional = true
						popped_str, _ := arr[0].(string)
						var err error
						popped, err = strconv.Atoi(popped_str)
						if err != nil {
							cmd.to.Write([]byte(encodeSimpleError("OPTIONAL ARG COULDNT BE RESOLVED")))
							continue
						}
					}
				}
				if arr, ok := val.value.([]any); ok {
					if popped >= len(arr) {
						popped = len(arr)
					}
					if !optional {
						obj, _ := encodeObj(arr[0])
						cmd.to.Write([]byte(obj))
						val.value = arr[1:]
					} else {
						obj, _ := encodeObj(arr[0:popped])
						cmd.to.Write([]byte(obj))
						val.value = arr[popped:]
					}

					if len(val.value.([]any)) == 0 {
						delete(storage, cmd.key)
					} else {
						storage[cmd.key] = val
					}
				} else {
					cmd.to.Write([]byte(NULL_BULK_STRING))
				}
			}
		case BLPOP:
			timeout := 0
			if arr, ok := cmd.value.([]any); ok {
				if len(arr) == 1 {
					timeout_str, _ := arr[0].(string)
					var err error
					timeout, err = strconv.Atoi(timeout_str)
					if err != nil {
						cmd.to.Write([]byte(encodeSimpleError("OPTIONAL ARG COULDNT BE RESOLVED")))
						continue
					}
				}
			}
			_ = timeout

			val, ok := storage[cmd.key]
			if ok {

				if arr, isArr := val.value.([]any); isArr && len(arr) > 0 {
					poppedValue := arr[0]
					val.value = arr[1:]
					if len(val.value.([]any)) == 0 {
						delete(storage, cmd.key)
					} else {
						storage[cmd.key] = val
					}

					obj, _ := encodeObj([]any{cmd.key, poppedValue})
					cmd.to.Write([]byte(obj))
					continue
				}
			}

			blocked_pops[cmd.key] = append(blocked_pops[cmd.key], cmd)
		case LLEN:
			val, ok := storage[cmd.key]
			var l int
			if !ok {
				l = 0
			} else {
				if val_arr, ok := val.value.([]any); ok {
					l = len(val_arr)
				} else {
					l = 0
				}
			}
			obj, err := encodeObj(l)
			if err != nil {
				cmd.to.Write([]byte(encodeSimpleError("ERROR WHILE HANDLING LLEN\n")))
				continue
			}
			cmd.to.Write([]byte(obj))
		case LRANGE:
			val, ok := storage[cmd.key]
			if !ok {
				cmd.to.Write([]byte(EMPTY_ARR))
				continue
			}
			low, err := strconv.Atoi(cmd.value.([]any)[0].(string))
			if err != nil {
				cmd.to.Write([]byte(encodeSimpleError("COULDNT CONVERT LOW TO INT")))
				continue
			}
			high, err := strconv.Atoi(cmd.value.([]any)[1].(string))
			if err != nil {
				cmd.to.Write([]byte(encodeSimpleError("COULDNT CONVERT HIGH TO INT")))
				continue
			}
			if arr, ok := val.value.([]any); ok {
				if low < 0 {
					low += len(arr)
				}
				if low < 0 {
					low = 0
				}
				if high < 0 {
					high += len(arr)
				}
				if high < 0 {
					high = 0
				}
				if low >= len(arr) || low > high {
					cmd.to.Write([]byte(EMPTY_ARR))
					continue
				}
				high += 1
				if high > len(arr) {
					high = len(arr)
				}
				obj, err := encodeObj(arr[low:high])
				if err != nil {
					cmd.to.Write([]byte(encodeSimpleError("COULDNT ENCODE OBJECT")))
					continue
				}
				cmd.to.Write([]byte(obj))
			} else {
				cmd.to.Write([]byte("COULDNT DECODE ARRAY"))
			}
		default:
			cmd.to.Write([]byte(NULL_BULK_STRING))
		}
	}
}
