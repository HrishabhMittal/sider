package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func handleCommand(cmd_arr []any) (string, error) {
	if len(cmd_arr) <= 0 {
		return "", NewError("command too short")
	}
	cmd, ok := cmd_arr[0].(string)
	cmd = strings.ToUpper(cmd)
	if ok {
		switch cmd {
		case "PING":
			return encodeObj(any("PONG"))
		case "ECHO":
			if len(cmd_arr) != 2 {
				return "", NewError("ECHO accepts exactly 1 argument")
			}
			return encodeObj(cmd_arr[1])
		default:
			return "", NewError("unrecognised cmd")
		}
	} else {
		return "", NewError("cmd not of type string")
	}
}
func handleConnection(con net.Conn) {
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
			resp, err := handleCommand(v)
			if err != nil {
				fmt.Println("Error: ", err)
				return
			}
			con.Write([]byte(resp))
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
	for {
		con, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(con)
	}
}
