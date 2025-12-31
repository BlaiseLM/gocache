package main

import (
	"fmt"
	"net"
	"strconv"
)

func main() { 
	listener, _ := net.Listen("tcp", ":8080") // creates server and starts listening for connections
	for { 
		connection, _ := listener.Accept() // catch a connection
		go handleConnection(connection) // allows connections to be handled concurrently (connection 1 is handled at the same time as connection 2, connection 3, etc. )
	}
}

func handleConnection(connection net.Conn) { 
	buffer := make([]byte, 4096)
	read, _ := connection.Read(buffer)

	fmt.Println("Number of bytes read: " + strconv.Itoa(read))
	fmt.Println("Content: " + string(buffer[:read]))

	write, _ := connection.Write(buffer[:read])
	fmt.Println("Number of bytes written: " + strconv.Itoa(write))
	fmt.Println("Content: " + string(buffer[:write]))
}