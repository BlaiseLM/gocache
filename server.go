package main

import (
	"github.com/BlaiseLM/gocache/cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/joho/godotenv"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("WARNING: Unable to load .env [%v]\n", err)
	}
}

func main() {
	capacity, err := strconv.Atoi(os.Getenv("CAPACITY"))
	if err != nil {
		log.Fatalf("FATAL ERROR: Invalid CAPACITY value in .env [%v]\n", err)
	}

	bufferSize, err := strconv.Atoi(os.Getenv("BUFFER_SIZE"))
	if err != nil {
		log.Fatalf("FATAL ERROR: Invalid BUFFER_SIZE value in .env [%v]\n", err)
	}

	cache := cache.NewCache(capacity, prometheus.DefaultRegisterer)
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(os.Getenv("P_ADDR"), nil))
	}()
	listener, err := net.Listen("tcp", os.Getenv("C_ADDR"))
	if err != nil {
		log.Fatalf("FATAL ERROR: Unable to start server [%v]\n", err)
	}
	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Printf("ERROR: Failed to accept connection [%v]\n", err)
			continue
		}
		go handleConnection(connection, cache, bufferSize)
	}
}

func handleConnection(connection net.Conn, cache *cache.Cache, bufferSize int) {
	defer connection.Close()
	for {
		buffer := make([]byte, bufferSize)
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
		case "get":
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
		case "set":
			if len(request) < 3 {
				connection.Write([]byte("ERROR: SET requires key and value\n"))
				continue
			}
			key := request[1]
			value := request[2]
			cache.Set(key, value)
			connection.Write([]byte("OK\n"))
		case "delete":
			if len(request) < 2 {
				connection.Write([]byte("ERROR: DELETE requires a key\n"))
				continue
			}
			key := request[1]
			cache.Delete(key)
			connection.Write([]byte("OK\n"))
		case "flush":
			if len(request) > 1 {
				connection.Write([]byte("ERROR: FLUSH doesn't require key and/or value\n"))
				continue
			}
			cache.Flush()
			connection.Write([]byte("OK\n"))
		case "end":
			connection.Write([]byte("Closing connection\n"))
			return
		default:
			connection.Write([]byte("ERROR: Unknown command\n"))
		}
	}
}
