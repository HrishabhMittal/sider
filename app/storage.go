package main

import "net"

var storage map[string]any = make(map[string]any)

const (
	GET = iota
	SET
)

type storage_cmd struct {
	cmd   int
	key   string
	value any
	to    net.Conn
}

func handleStorage(cmds chan storage_cmd) {
	for v := range cmds {
		switch v.cmd {
		case GET:
			val, ok := storage[v.key]
			if !ok {
				v.to.Write([]byte(NULL_BULK_STRING))
				continue
			}
			str, _ := encodeObj(val)
			v.to.Write([]byte(str))
		case SET:
			storage[v.key] = v.value
			v.to.Write([]byte(encodeSimpleString("OK")))
		default:
			v.to.Write([]byte(NULL_BULK_STRING))
		}
	}
}
