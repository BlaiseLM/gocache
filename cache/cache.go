package cache

import (
	"sync"
	"sync/atomic"
)

type Metrics struct { 
	TotalSetRequests 		int64
	TotalGetRequests 		int64
	TotalDeleteRequests 	int64
	TotalFlushRequests		int64
	TotalCacheHits			int64
	TotalCacheMiss			int64
	TotalEvictions			int64
}

func (c *Cache) GetMetrics() Stats { 
	c.mu.RLock()
	size := len(c.data)
	c.mu.RUnlock()

	return Stats{ 
		Sets : atomic.LoadInt64(&c.me.TotalSetRequests), 
		Gets: atomic.LoadInt64(&c.me.TotalGetRequests), 
		Deletes: atomic.LoadInt64(&c.me.TotalDeleteRequests), 
		Flushes: atomic.LoadInt64(&c.me.TotalFlushRequests), 
		Hits: atomic.LoadInt64(&c.me.TotalCacheHits), 
		Misses: atomic.LoadInt64(&c.me.TotalCacheMiss), 
		Evictions: atomic.LoadInt64(&c.me.TotalEvictions),
		Size: int64(size), 
		Capacity: int64(c.capacity), 
	}
}

type Stats struct {
	Sets    	int64
	Gets    	int64
	Deletes 	int64
	Flushes 	int64
	Hits    	int64
	Misses  	int64
	Evictions	int64
	Size  		int64
	Capacity 	int64
} 

func (s Stats) GetHitRate() float64{ 
	hits := s.Hits
	total := s.Hits + s.Misses
	if total == 0 { 
		return 0 
	}
	return float64(hits) / float64(total) *100
}

type Node struct {
	Key   string
	Value string
	Next  *Node
	Prev  *Node
}

func NewNode(key string, value string, next *Node, prev *Node) *Node {
	return &Node{
		Key:   key,
		Value: value,
		Next:  next,
		Prev:  prev,
	}
}

type Cache struct {
	capacity int
	data     map[string]*Node
	mu       sync.RWMutex
	me 		 *Metrics
	head     *Node
	tail     *Node
}

func NewCache(capacity int) *Cache {
	return &Cache{
		capacity: capacity,
		data:     make(map[string]*Node),
		me:  &Metrics{}, 
		head:     nil,
		tail:     nil,
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
	if key == "" || c.capacity == 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	node, ok := c.data[key]
	atomic.AddInt64(&c.me.TotalSetRequests, 1)

	if ok {
		if node == nil {
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
			atomic.AddInt64(&c.me.TotalEvictions, 1)
		}
		newNode := NewNode(key, value, nil, nil)
		c.data[key] = newNode
		c.addToFront(newNode)
	}
}

func (c *Cache) Get(key string) (string, bool) {
	if key == "" || c.capacity == 0 {
		return "", false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	node, ok := c.data[key]
	atomic.AddInt64(&c.me.TotalGetRequests, 1)

	if ok {
		if node == nil {
			return "", false
		}
		value := node.Value
		c.removeNode(node)
		c.addToFront(node)
		atomic.AddInt64(&c.me.TotalCacheHits, 1)
		return value, true
	} else {
		atomic.AddInt64(&c.me.TotalCacheMiss, 1)
		return "", false
	}
}

func (c *Cache) Delete(key string) {
	if key == "" || c.capacity == 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	
	node, ok := c.data[key]
	atomic.AddInt64(&c.me.TotalDeleteRequests, 1)

	if ok {
		if node == nil {
			return
		}
		c.removeNode(node)
		delete(c.data, key)
	} else {
		return
	}
}

func (c *Cache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	atomic.AddInt64(&c.me.TotalFlushRequests, 1)
	clear(c.data)
	c.head = nil
	c.tail = nil
}