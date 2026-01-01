package main

import (
	"log"
	"net"
	"github.com/BlaiseLM/gocache/cache"
	"strings"
)

func main() { 
	cache := cache.NewCache(1024)
	listener, error := net.Listen("tcp", ":8080") 
	if error != nil { 
		log.Fatalf("FATAL ERROR: Unable to start server [%v]\n", error)
	}
	for { 
		connection, error := listener.Accept() 
		if error != nil { 
			log.Printf("ERROR: Unable to accept connection [%v]\n", error)
			// TODO: Retry accepting connection (with exp backoff?)
		}
		go handleConnection(connection, cache) 
	}
}

func handleConnection(connection net.Conn, cache *cache.Cache) { 
	// TODO: Check for broken conections
	defer connection.Close()
	for { 
		buffer := make([]byte, 4096)
		read, error := connection.Read(buffer)
		if error != nil {
			log.Printf("ERROR: Unable to read requests [%v]", error)
			return
		}
		if read == 0 {
			return
		}
		request := strings.Fields(string(buffer[:read]))
		if len(request) == 0 { 
			continue
		}
		// TODO: Handle malformed requests (missing args, etc)
		switch method := strings.ToLower(request[0]); method { 
		case "get" :
			if len(request) < 2 {
				connection.Write([]byte("ERROR: GET requires a key\n"))
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