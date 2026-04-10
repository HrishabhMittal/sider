package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func send(con net.Conn, stuff []string) {
	var sb strings.Builder
	sb.WriteString("*")
	sb.WriteString(strconv.Itoa(len(stuff)))
	sb.WriteString("\r\n")

	for _, val := range stuff {
		sb.WriteString("$")
		sb.WriteString(strconv.Itoa(len(val)))
		sb.WriteString("\r\n")
		sb.WriteString(val)
		sb.WriteString("\r\n")
	}

	con.Write([]byte(sb.String()))
}
func main() {
	con, err := net.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		os.Exit(1)
	}
	defer con.Close()

	cmds := [][]string{
		{"LPUSH", "apple", "grape"},
		{"LPUSH", "apple", "orange", "rasberry"},
		{"LRANGE", "apple", "0", "-1"},
		{"LLEN", "apple"},
		{"LLEN", "diddle"},
	}
	buf := make([]byte, 4096) // nobody cares
	for i, v := range cmds {
		fmt.Printf("%d: %v\n", i+1, v)
		send(con, v)
		n, err := con.Read(buf)
		if err != nil {
			fmt.Printf("read Error: %v\n", err)
			break
		}
		fmt.Printf("recv: %q\n\n", string(buf[:n]))
	}
}
