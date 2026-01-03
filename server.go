package main

import (
	"log"
	"net"
	"github.com/BlaiseLM/gocache/cache"
	"strings"
)

func main() { 
    cache := cache.NewCache(1024)
    listener, err := net.Listen("tcp", ":8080") 
    if err != nil { 
        log.Fatalf("FATAL ERROR: Unable to start server [%v]\n", err)
    }
    for { 
        connection, err := listener.Accept()
        if err != nil {
            log.Printf("ERROR: Failed to accept connection [%v]\n", err)
            continue 
        }
        go handleConnection(connection, cache)
    }
}

func handleConnection(connection net.Conn, cache *cache.Cache) { 
	defer connection.Close()
	for { 
		buffer := make([]byte, 4096)
		read, err := connection.Read(buffer)
		if err != nil {
			log.Printf("ERROR: Unable to read requests [%v]", err)
			return
		}
		if read == 0 {
			return
		}
		request := strings.Fields(string(buffer[:read]))
		if len(request) == 0 { 
			continue
		}
		switch method := strings.ToLower(request[0]); method { 
		case "get" :
			if len(request) < 2 {
				connection.Write([]byte("ERROR: GET requires a key\n"))
				continue
			}
			key := request[1]
			value, ok := cache.Get(key)
			if ok { 
				connection.Write([]byte(value + "\n"))  
			} else {
				connection.Write([]byte("(nil)\n"))
			}
		case "set" :
			if len(request) < 3 {
				connection.Write([]byte("ERROR: SET requires key and value\n"))
				continue
			}
			key := request[1]
			value := request[2]
			cache.Set(key, value)
			connection.Write([]byte("OK\n"))  
		case "end" : 
			connection.Write([]byte("Closing connection\n"))
			return

		default: 
			connection.Write([]byte("ERROR: Unknown command\n"))
		}	
	}
}
