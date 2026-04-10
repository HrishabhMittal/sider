package main

import (
	"net"
	"strconv"
	"strings"
	"time"
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
			if len(cmd_arr) < 3 {
				return NewError("SET accepts atleast 2 arguments")
			}
			key, ok := cmd_arr[1].(string)
			duration := -1 * time.Second
			if len(cmd_arr) == 5 {
				str, ok := cmd_arr[3].(string)
				if ok {
					str_num, ok := cmd_arr[4].(string)
					if ok {
						num, err := strconv.Atoi(str_num)
						if err != nil {
							return err
						}
						str = strings.ToUpper(str)
						switch str {
						case "PX":
							duration = time.Duration(num) * time.Millisecond
						case "EX":
							duration = time.Duration(num) * time.Second
						default:
						}
					}
				}
			}
			if !ok {
				return NewError("couldn't resolve key")
			}
			cmd_channel <- storage_cmd{
				cmd:       SET,
				key:       key,
				value:     cmd_arr[2],
				to:        con,
				timestamp: time.Now(),
				expiry:    duration,
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
				cmd:       GET,
				key:       key,
				value:     nil,
				to:        con,
				timestamp: time.Now(),
			}
		case "RPUSH":
			if len(cmd_arr) < 3 {
				return NewError("RPUSH accepts more than 2 argument")
			}
			key, ok := cmd_arr[1].(string)
			if !ok {
				return NewError("couldn't resolve key")
			}
			cmd_channel <- storage_cmd{
				cmd:       RPUSH,
				key:       key,
				value:     cmd_arr[2:],
				to:        con,
				timestamp: time.Now(),
			}
		case "LPUSH":
			if len(cmd_arr) < 3 {
				return NewError("RPUSH accepts more than 2 argument")
			}
			key, ok := cmd_arr[1].(string)
			if !ok {
				return NewError("couldn't resolve key")
			}
			cmd_channel <- storage_cmd{
				cmd:       LPUSH,
				key:       key,
				value:     cmd_arr[2:],
				to:        con,
				timestamp: time.Now(),
			}

		case "LPOP":
			if len(cmd_arr) != 2 {
				return NewError("LPOP accepts exactly 1 arguments")
			}
			key, ok := cmd_arr[1].(string)
			if !ok {
				return NewError("couldn't resolve key")
			}
			cmd_channel <- storage_cmd{
				cmd:       LPOP,
				key:       key,
				value:     cmd_arr[2:],
				to:        con,
				timestamp: time.Now(),
			}
		case "LLEN":
			if len(cmd_arr) != 2 {
				return NewError("LLEN accepts exactly 1 arguments")
			}
			key, ok := cmd_arr[1].(string)
			if !ok {
				return NewError("couldn't resolve key")
			}
			cmd_channel <- storage_cmd{
				cmd:       LLEN,
				key:       key,
				value:     cmd_arr[2:],
				to:        con,
				timestamp: time.Now(),
			}
		case "LRANGE":
			if len(cmd_arr) != 4 {
				return NewError("LRANGE accepts exactly 3 arguments")
			}
			key, ok := cmd_arr[1].(string)
			if !ok {
				return NewError("couldn't resolve key")
			}
			cmd_channel <- storage_cmd{
				cmd:       LRANGE,
				key:       key,
				value:     cmd_arr[2:],
				to:        con,
				timestamp: time.Now(),
			}
		default:
			return NewError("unrecognised cmd")
		}
	} else {
		return NewError("cmd not of type string")
	}
	return nil
}
