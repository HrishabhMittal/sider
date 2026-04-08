package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
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
func handleConnection(con net.Conn, cmd_channel chan storage_cmd) {
	defer con.Close()
	reader := NewRespReader(bufio.NewReader(con))
	for {
		obj, err := reader.decodeObj()
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error: ", err)
			}
			return
		}
		switch v := obj.(type) {
		case []any:
			err := handleCommand(v, cmd_channel, con)
			if err != nil {
				fmt.Println("Error: ", err)
				return
			}
		}
	}
}
func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	fmt.Println("listening on port 6379")
	cmd_channel := make(chan storage_cmd, 10)
	go handleStorage(cmd_channel)
	for {
		con, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(con, cmd_channel)
	}
}
