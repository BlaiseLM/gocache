package cache

import (
	"reflect"
	"testing"
	"sync"
	"fmt"
)

func TestSetToCache(t *testing.T){ 
	cache := NewCache(3)
	node := NewNode("test", "value", nil, nil)

	cache.Set("test", "value")
	result := cache.data["test"]

	if !reflect.DeepEqual(node, result){ 
		t.Errorf("Error occured while adding item to cache. Expected: %v, Got: %v", node, result)
	}
}

func TestGetFromCache(t *testing.T){ 
	cache := NewCache(3)
	cache.Set("test", "value")

	result, ok := cache.Get("test")

	if result != "value" || !ok{ 
		t.Errorf("Error occured while getting item from cache. Expected: %v, Got: %v", "value", result)
	}
}

func TestEviction(t *testing.T){ 
	cache := NewCache(3)
	cache.Set("test1", "value1")
	cache.Set("test2", "value2")
	cache.Set("test3", "value3")
	cache.Set("test4", "value4")

	if cache.tail.Key == "test1" || cache.capacity != 3{ 
		t.Errorf("Error occured while evicting LRU item from cache.")
	}
}

func TestDeleteFromCache(t *testing.T){ 
	cache := NewCache(3)
	cache.Set("test", "value")
	cache.Delete("test")
	_, ok := cache.data["test"] 

	if ok{ 
		t.Errorf("Error occured while deleting item from cache.")
	}
}

func TestConcurrentAccess(t *testing.T) {
	cache := NewCache(100)

	var wg sync.WaitGroup

	setOperations := func() { 
		for i := 0; i < 100; i++{
			key := fmt.Sprintf("test%d", i) 
			value := fmt.Sprintf("value%d", i)
			cache.Set(key, value)
		}
	}

	getOperations := func() { 
		for j := 0; j < 100; j++{
			key := fmt.Sprintf("test%d", j) 
			cache.Get(key)
		}
	}

	for k := 0; k < 10; k++{ 
		wg.Add(1)
		go func(){ 
			defer wg.Done()

			setOperations()
			getOperations()
		}()
	}

	wg.Wait()
}