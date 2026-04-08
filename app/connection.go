package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
)

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
