# Go Cache

Building a cache server from scratch using Go. 

## Description
So far, I've built an in-memory LRU cache. It is implemented using a hash map for O(1) lookups and a doubly linked list for O(1) data manipulation. The map supports two operations, Get() and Set(). Get() returns the value associated with a key. Set() updates the value of a key and/or creates a new key-value pair. The cache uses LRU (Least Recently Used) eviction to remove the item that hasn't been accessed for the longest time when it reaches capacity. 

## Getting Started

### Dependencies

* [Stable version of Go](https://go.dev/dl/)

### Installing

1. Create new local repository: 
   ```
   mkdir gocache
   cd gocache
   ```
2. Clone remote repository: 
   ```
    git clone <method> <project-name>.
   ```
3. Open project

### Executing program
To run the project, run the following command:  
```
go run cache.go
```


## Roadmap: 
- [ ] Networking
  - [ ] Simple TCP server
  - [ ] Parse GET/SET requests
  - [ ] Call corresponding cache methods
  - [ ] Error handling
