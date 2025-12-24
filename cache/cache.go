package cache

import (
	"fmt"
)


type Node struct { 
    Key string
    Value string
    Next *Node
    Prev *Node 
}

func NewNode(key string, value string, next *Node, prev *Node) *Node { 
    return &Node{ 
        Key: key,
        Value: value,
        Next: next, 
        Prev: prev,  
    }
}

type Cache struct {
    capacity int
    data map[string]*Node
    head *Node
    tail *Node
}

func NewCache(capacity int) *Cache {
    return &Cache{
        capacity: capacity,
        data: make(map[string]*Node),
        head: nil, 
        tail: nil, 
    }
}

func (c *Cache) addToFront(node *Node) {
    if c.head == nil {
        c.head = node 
        c.tail = node
        node.Next = nil
        node.Prev = nil
    } else {
        node.Next = c.head
        node.Prev = nil
        c.head.Prev = node
        c.head = node
    }
}

func (c *Cache) removeNode(node *Node) {
    if c.head == nil {
        return
    } else if c.head == c.tail && c.head != nil {
        c.head = nil
        c.tail = nil
    } else if c.head == node && c.tail != node {
        c.head = node.Next
        node.Next = nil
        node.Prev = nil
        c.head.Prev = nil
    } else if c.tail == node && c.head != node {
        c.tail = node.Prev
        node.Next = nil
        node.Prev = nil
        c.tail.Next = nil
    } else {
        prev := node.Prev
        next := node.Next
        node.Prev = nil
        node.Next = nil
        next.Prev = prev
        prev.Next = next
    }
}

func (c *Cache) Set(key, value string) {
    if key == "" {
        fmt.Println("Key cannot be empty")
        return
    }
    node, ok := c.data[key]
    if ok {
        if node == nil {
            fmt.Println("Node cannot be null")
            return
        }
        c.removeNode(node)
        node.Value = value
        c.addToFront(node)
    } else {
        if len(c.data) >= c.capacity {
            evictKey := c.tail.Key
            c.removeNode(c.tail)
            delete(c.data, evictKey)
        }
        newNode := NewNode(key, value, nil, nil)
        c.data[key] = newNode
        c.addToFront(newNode)
    }
}

func (c *Cache) Get(key string) (string, bool) {
    if key == "" {
        fmt.Println("Key cannot be empty")
        return "", false
    }

    node, ok := c.data[key]
    if ok {
        if node == nil {
            return "", false
        }
        value := node.Value
        c.removeNode(node)
        c.addToFront(node)
        return value, true
    } else {
        fmt.Println("Key doesn't exist")
        return "", false
    }
}