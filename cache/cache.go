package cache

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

type Metrics struct {
	TotalSetRequests    prometheus.Counter
	TotalGetRequests    prometheus.Counter
	TotalDeleteRequests prometheus.Counter
	TotalFlushRequests  prometheus.Counter
	TotalCacheHits      prometheus.Counter
	TotalCacheMisses    prometheus.Counter
	TotalEvictions      prometheus.Counter
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		TotalSetRequests: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "total_set_requests",
			Help: "Number of set requests.",
		}),
		TotalGetRequests: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "total_get_requests",
			Help: "Number of get requests.",
		}),
		TotalDeleteRequests: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "total_delete_requests",
			Help: "Number of delete requests.",
		}),
		TotalFlushRequests: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "total_flush_requests",
			Help: "Number of flush requests.",
		}),
		TotalCacheHits: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "total_cache_hits",
			Help: "Number of cache hits.",
		}),
		TotalCacheMisses: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "total_cache_misses",
			Help: "Number of cache misses.",
		}),
		TotalEvictions: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "total_evictions",
			Help: "Number of evictions.",
		}),
	}
	reg.MustRegister(
		m.TotalSetRequests,
		m.TotalGetRequests,
		m.TotalDeleteRequests,
		m.TotalFlushRequests,
		m.TotalCacheHits,
		m.TotalCacheMisses,
		m.TotalEvictions,
	)
	return m
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
	me       *Metrics
	head     *Node
	tail     *Node
}

func NewCache(capacity int, reg prometheus.Registerer) *Cache {
	return &Cache{
		capacity: capacity,
		data:     make(map[string]*Node),
		me:       NewMetrics(reg),
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
	c.me.TotalSetRequests.Inc()
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
			c.me.TotalEvictions.Inc()
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
	c.me.TotalGetRequests.Inc()
	if ok {
		if node == nil {
			return "", false
		}
		value := node.Value
		c.removeNode(node)
		c.addToFront(node)
		c.me.TotalCacheHits.Inc()
		return value, true
	} else {
		c.me.TotalCacheMisses.Inc()
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
	c.me.TotalDeleteRequests.Inc()

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
	c.me.TotalFlushRequests.Inc()
	clear(c.data)
	c.head = nil
	c.tail = nil
}

// Dummy Prometheus registerer for testing purposes only
type NoOpRegisterer struct{}

func (n *NoOpRegisterer) Register(prometheus.Collector) error {
	return nil
}

func (n *NoOpRegisterer) MustRegister(...prometheus.Collector) {}

func (n *NoOpRegisterer) Unregister(prometheus.Collector) bool {
	return true
}
