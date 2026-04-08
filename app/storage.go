package main

import (
	"net"
	"time"
)

var storage map[string]storage_obj = make(map[string]storage_obj)

const (
	GET = iota
	SET
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
		default:
			v.to.Write([]byte(NULL_BULK_STRING))
		}
	}
}
