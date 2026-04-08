package main

import (
	"net"
	"strings"
)

func handleCommand(cmd_arr []any, cmd_channel chan storage_cmd, con net.Conn) error {
	if len(cmd_arr) <= 0 {
		return NewError("command too short")
	}
	cmd, ok := cmd_arr[0].(string)
	cmd = strings.ToUpper(cmd)
	if ok {
		switch cmd {
		case "PING":
			con.Write([]byte(encodeSimpleString("PONG")))
		case "ECHO":
			if len(cmd_arr) != 2 {
				return NewError("ECHO accepts exactly 1 argument")
			}
			val, err := encodeObj(cmd_arr[1])
			if err != nil {
				return err
			}
			con.Write([]byte(val))
		case "SET":
			if len(cmd_arr) != 3 {
				return NewError("SET accepts exactly 2 arguments")
			}
			key, ok := cmd_arr[1].(string)
			if !ok {
				return NewError("couldn't resolve key")
			}
			cmd_channel <- storage_cmd{
				cmd:   SET,
				key:   key,
				value: cmd_arr[2],
				to:    con,
			}
		case "GET":
			if len(cmd_arr) != 2 {
				return NewError("GET accepts exactly 1 argument")
			}
			key, ok := cmd_arr[1].(string)
			if !ok {
				return NewError("couldn't resolve key")
			}
			cmd_channel <- storage_cmd{
				cmd:   GET,
				key:   key,
				value: nil,
				to:    con,
			}
		default:
			return NewError("unrecognised cmd")
		}
	} else {
		return NewError("cmd not of type string")
	}
	return nil
}
