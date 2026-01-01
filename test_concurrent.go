package main

// TODO: Rewrite from scratch 

import (
    "fmt"
    "net"
    "sync"
    "time"
)

func client(id int, wg *sync.WaitGroup) {
    defer wg.Done()
    
    conn, err := net.Dial("tcp", "localhost:8080")
    if err != nil {
        fmt.Printf("Client %d: connection failed\n", id)
        return
    }
    defer conn.Close()
    
    // Send SET command
    msg := fmt.Sprintf("SET key%d value%d\n", id, id)
    conn.Write([]byte(msg))
    
    time.Sleep(10 * time.Millisecond)
    
    // Send GET command
    msg = fmt.Sprintf("GET key%d\n", id)
    conn.Write([]byte(msg))
    
    buffer := make([]byte, 1024)
    conn.Read(buffer)
    fmt.Printf("Client %d got: %s", id, buffer)
}

func main() {
    var wg sync.WaitGroup
    
    // Launch 100 concurrent clients
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go client(i, &wg)
    }
    
    wg.Wait()
    fmt.Println("All clients done")
}