package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func handleConnection(con net.Conn) {
	defer con.Close()
	reader := NewRespReader(bufio.NewReader(con))
	for {
		obj, err := reader.decodeObj()
		if err != nil {
			fmt.Println("Error: ", err)
			return
		}
		switch v := obj.(type) {
		case []any:
			if len(v) <= 0 {
				continue
			}
			cmd, ok := v[0].(string)
			if ok {
				if cmd=="PING" {
					con.Write([]byte("+PONG\r\n"))
				} else {
					fmt.Println("error: unrecognised cmd")
					return
				}
			} else {
				fmt.Println("error: cmd not of type string")
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
	for {
		con, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(con)
	}
}
