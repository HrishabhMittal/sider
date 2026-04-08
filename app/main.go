package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)
func handleConnection(con net.Conn) {
	defer con.Close()
	reader := bufio.NewReader(con)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("connection error: ",err)
			return
		}
		if len(line)<2 {
			fmt.Println("recieved bogus data")
			continue
		}
		end := line[len(line)-2:]
		if end != "\r\n" {
			fmt.Println("recieved bogus data")
			continue
		}
		line = line[:len(line)-2]
		if line == "+PING" {
			con.Write([]byte("+PONG\r\n"))
		}
	}
}


func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	for {
		con, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(con)
	}
}
