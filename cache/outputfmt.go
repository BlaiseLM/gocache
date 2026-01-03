package cache

import (
	"fmt"
)

func (c *Cache) Print() {
	fmt.Println("Cache contents (head to tail):")
	current := c.head
	for current != nil {
		fmt.Printf("  %s: %s\n", current.Key, current.Value)
		current = current.Next
	}
	fmt.Printf("Size: %d/%d\n\n", len(c.data), c.capacity)
}
