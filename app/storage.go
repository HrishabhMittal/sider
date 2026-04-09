package main

import (
	"net"
	"strconv"
	"time"
)

var storage map[string]storage_obj = make(map[string]storage_obj)

const (
	GET = iota
	SET
	RPUSH
	LRANGE
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

//	func clamp(num int,low int,high int) int {
//		if num < low {
//			num = low
//		} else if num > high {
//			num = high
//		}
//		return num
//	}
func handleStorage(cmds chan storage_cmd) {
	for v := range cmds {
		switch v.cmd {
		case GET:
			val, ok := storage[v.key]
			if !ok {
				v.to.Write([]byte(NULL_BULK_STRING))
			} else if val.expiry > 0 && val.timestamp.Add(val.expiry).Compare(v.timestamp) != 1 {
				delete(storage, v.key)
				v.to.Write([]byte(NULL_BULK_STRING))
			} else {
				str, _ := encodeObj(val.value)
				v.to.Write([]byte(str))
			}
		case SET:
			storage[v.key] = storage_obj{
				value:     v.value,
				timestamp: v.timestamp,
				expiry:    v.expiry,
			}
			v.to.Write([]byte(encodeSimpleString("OK")))
		case RPUSH:
			val, ok := storage[v.key]
			if !ok {
				obj := []any{}
				storage[v.key] = storage_obj{
					value:     obj,
					timestamp: v.timestamp,
				}
			}
			val, ok = storage[v.key]
			if arr, ok := val.value.([]any); ok {
				arr = append(arr, v.value.([]any)...)
				obj, err := encodeObj(len(arr))
				val.value = arr
				storage[v.key] = val
				if err != nil {
					v.to.Write([]byte(encodeSimpleError("INTERNAL ERROR")))
				} else {
					v.to.Write([]byte(obj))
				}
			} else {
				v.to.Write([]byte(encodeSimpleString("TYPE ERROR")))
			}
		case LRANGE:
			val, ok := storage[v.key]
			if !ok {
				v.to.Write([]byte(encodeSimpleError("KEY DOESNT EXIST")))
			}
			low, err := strconv.Atoi(v.value.([]any)[0].(string))
			if err != nil {
				v.to.Write([]byte(encodeSimpleError("COULDNT CONVERT LOW TO INT")))
				return
			}
			high, err := strconv.Atoi(v.value.([]any)[1].(string))
			if err != nil {
				v.to.Write([]byte(encodeSimpleError("COULDNT CONVERT HIGH TO INT")))
				return
			}
			if arr, ok := val.value.([]any); ok {
				if low < 0 || low >= len(arr) || high < 0 || low > high {
					v.to.Write([]byte(EMPTY_ARR))
					return
				}
				high += 1
				if high > len(arr) {
					high = len(arr)
				}
				obj, err := encodeObj(arr[low:high])
				if err != nil {
					v.to.Write([]byte(encodeSimpleError("COULDNT ENCODE OBJECT")))
					return
				}
				v.to.Write([]byte(obj))
			}
		default:
			v.to.Write([]byte(NULL_BULK_STRING))
		}
	}
}
