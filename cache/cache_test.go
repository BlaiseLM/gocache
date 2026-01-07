package cache

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
)

func TestSetToCache(t *testing.T) {
	cache := NewCache(3, &NoOpRegisterer{})
	node := NewNode("test", "value", nil, nil)
	cache.Set("test", "value")
	result := cache.data["test"]
	if !reflect.DeepEqual(node, result) {
		t.Errorf("Error occurred while adding item to cache. Expected: %v, Got: %v", node, result)
	}
}

func TestGetFromCache(t *testing.T) {
	cache := NewCache(3, &NoOpRegisterer{})
	cache.Set("test", "value")
	result, ok := cache.Get("test")
	if result != "value" || !ok {
		t.Errorf("Error occurred while getting item from cache. Expected: %v, Got: %v", "value", result)
	}
}

func TestEviction(t *testing.T) {
	cache := NewCache(3, &NoOpRegisterer{})
	cache.Set("test1", "value1")
	cache.Set("test2", "value2")
	cache.Set("test3", "value3")
	cache.Set("test4", "value4")
	if cache.tail.Key == "test1" || cache.capacity != 3 {
		t.Errorf("Error occurred while evicting LRU item from cache.")
	}
}

func TestDeleteFromCache(t *testing.T) {
	cache := NewCache(3, &NoOpRegisterer{})
	cache.Set("test", "value")
	cache.Delete("test")
	_, ok := cache.data["test"]
	if ok {
		t.Errorf("Error occurred while deleting item from cache.")
	}
}

func TestDeleteNonExistentKey(t *testing.T) {
	cache := NewCache(3, &NoOpRegisterer{})
	cache.Set("test", "value")
	cache.Delete("nonexistent")
	_, ok := cache.data["test"]
	if !ok {
		t.Errorf("Error occurred while deleting non-existent key from cache.")
	}
}

func TestGetNonExistentKey(t *testing.T) {
	cache := NewCache(3, &NoOpRegisterer{})
	_, ok := cache.Get("nonexistent")
	if ok {
		t.Errorf("Error occurred while getting non-existent key from cache.")
	}
}

func TestSetEmptyKey(t *testing.T) {
	cache := NewCache(3, &NoOpRegisterer{})
	cache.Set("", "value")
	_, ok := cache.data[""]
	if ok {
		t.Errorf("Error occurred while setting empty key in cache.")
	}
}

func TestGetEmptyKey(t *testing.T) {
	cache := NewCache(3, &NoOpRegisterer{})
	_, ok := cache.Get("")
	if ok {
		t.Errorf("Error occurred while getting empty key from cache.")
	}
}

func TestDeleteEmptyKey(t *testing.T) {
	cache := NewCache(3, &NoOpRegisterer{})
	cache.Set("test", "value")
	cache.Delete("")
	_, ok := cache.data["test"]
	if !ok {
		t.Errorf("Error occurred while deleting empty key from cache.")
	}
}

func TestSetUpdateValue(t *testing.T) {
	cache := NewCache(3, &NoOpRegisterer{})
	cache.Set("test", "value1")
	cache.Set("test", "value2")
	result, ok := cache.Get("test")
	if !ok || result != "value2" {
		t.Errorf("Error occurred while updating value for existing key in cache. Expected: %v, Got: %v", "value2", result)
	}
}

func TestEvictionOrder(t *testing.T) {
	cache := NewCache(2, &NoOpRegisterer{})
	cache.Set("test1", "value1")
	cache.Set("test2", "value2")
	cache.Get("test1")
	cache.Set("test3", "value3")
	_, ok := cache.Get("test2")
	if ok {
		t.Errorf("Error occurred while evicting LRU item from cache. 'test2' should have been evicted.")
	}
}

func TestCapacityZero(t *testing.T) {
	cache := NewCache(0, &NoOpRegisterer{})
	cache.Set("test", "value")
	_, ok := cache.Get("test")
	if ok {
		t.Errorf("Error occurred while handling zero capacity cache. No items should be stored.")
	}
}

func TestCapacityOne(t *testing.T) {
	cache := NewCache(1, &NoOpRegisterer{})
	cache.Set("test1", "value1")
	cache.Set("test2", "value2")
	_, ok1 := cache.Get("test1")
	value2, ok2 := cache.Get("test2")
	if ok1 {
		t.Errorf("Error occurred while handling capacity one cache. 'test1' should have been evicted.")
	}
	if !ok2 || value2 != "value2" {
		t.Errorf("Error occurred while handling capacity one cache. 'test2' should be present with correct value.")
	}
}

func TestFlushCache(t *testing.T) {
	cache := NewCache(3, &NoOpRegisterer{})
	cache.Set("test1", "value1")
	cache.Set("test2", "value2")
	cache.Flush()
	if len(cache.data) != 0 {
		t.Errorf("Error occurred while flushing the cache. Cache should be empty.")
	}
}

func TestFlushEmptyCache(t *testing.T) {
	cache := NewCache(3, &NoOpRegisterer{})
	cache.Flush()
	if len(cache.data) != 0 {
		t.Errorf("Error occurred while flushing an already empty cache. Cache should remain empty.")
	}
}

func TestConcurrentAccess(t *testing.T) {
	cache := NewCache(100, &NoOpRegisterer{})
	var wg sync.WaitGroup
	setOperations := func() {
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("test%d", i)
			value := fmt.Sprintf("value%d", i)
			cache.Set(key, value)
		}
	}
	getOperations := func() {
		for j := 0; j < 100; j++ {
			key := fmt.Sprintf("test%d", j)
			cache.Get(key)
		}
	}
	for k := 0; k < 10; k++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			setOperations()
			getOperations()
		}()
	}
	wg.Wait()
}

func BenchmarkSetNoEviction(b *testing.B) { 
    cache := NewCache(1000000, &NoOpRegisterer{}) 
	b.ResetTimer()
	i := 0
    for b.Loop() {
		key := fmt.Sprintf("Key%d", i)
		value := fmt.Sprintf("Value%d", i)
		cache.Set(key, value)
		i++
    }
	/* 
	for range b.N: 498.2 ns/op
	b.Loop(): 573.6 ns/op 
	*/ 
}

func BenchmarkSetWithEviction(b *testing.B) { 
	cache := NewCache(1000, &NoOpRegisterer{})
	func(){ 
		for i := 0; i < 1000; i++{ 
			key := fmt.Sprintf("Key%d", i)
			value := fmt.Sprintf("Value%d", i)
			cache.Set(key, value)
		}
	}()
	b.ResetTimer()
	j := 1
	for b.Loop(){
		key := fmt.Sprintf("Key%d", 1000 + j)
		value := fmt.Sprintf("Value%d", 1000 + j)
		cache.Set(key, value)
		j++
	}
	/* 
	for range b.N: 242.4 ns/op
	b.Loop(): 250.4 ns/op 
	*/
}

func BenchmarkGetHit(b *testing.B) { 
	cache := NewCache(1000, &NoOpRegisterer{})
	func(){ 
		for i := 0; i < 100; i++{ 
			key := fmt.Sprintf("Key%d", i)
			value := fmt.Sprintf("Value%d", i)
			cache.Set(key, value)
		}
	}()
	b.ResetTimer()
	j := 0
	for b.Loop(){ 
		key := fmt.Sprintf("Key%d", j % 100)
		cache.Get(key)
		j++
	}
	/*
	for range b.N: 101.8 ns/op
	b.Loop(): 187.6 ns/op
	*/
}

func BenchmarkGetMiss(b *testing.B) { 
	cache := NewCache(1000, &NoOpRegisterer{})
	func(){ 
		for i := 0; i < 100; i++{ 
			key := fmt.Sprintf("Key%d", i)
			value := fmt.Sprintf("Value%d", i)
			cache.Set(key, value)
		}
	}()
	b.ResetTimer()
	j := 0
	for b.Loop(){ 
		key := fmt.Sprintf("Key%d", 100 + j)
		cache.Get(key)
		j++
	}
	/*
	for range b.N: 141.7 ns/op
	b.Loop(): 156.2  ns/op
	*/
}

func BenchmarkConcurrent(b *testing.B){ 
    cache := NewCache(10000, &NoOpRegisterer{})
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        i := 0
        for pb.Next() {
            if i % 5 == 0 {
                cache.Set(fmt.Sprintf("key%d", i), "value")
            } else {
                cache.Get(fmt.Sprintf("key%d", i))
            }
            i++
        }
    })
	/*
	381.6 ns/op
	*/
}