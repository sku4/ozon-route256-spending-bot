package lru

import (
	"container/list"
	"sync"
)

type Item struct {
	Key   string
	Value interface{}
}

type LRU struct {
	capacity int
	queue    *list.List
	mutex    *sync.RWMutex
	items    map[string]*list.Element
}

func NewLRU(capacity int) *LRU {
	return &LRU{
		capacity: capacity,
		queue:    list.New(),
		mutex:    new(sync.RWMutex),
		items:    make(map[string]*list.Element),
	}
}

func (c *LRU) Add(key string, value interface{}) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if element, exists := c.items[key]; exists {
		c.queue.MoveToFront(element)
		element.Value.(*Item).Value = value
		return true
	}

	if c.queue.Len() == c.capacity {
		c.limitQueueCapacity()
	}

	item := &Item{
		Key:   key,
		Value: value,
	}

	element := c.queue.PushFront(item)
	c.items[item.Key] = element

	return true
}

func (c *LRU) Get(key string) interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	element, exists := c.items[key]
	if !exists {
		return nil
	}

	c.queue.MoveToFront(element)
	return element.Value.(*Item).Value
}

func (c *LRU) Remove(key string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if val, found := c.items[key]; found {
		c.deleteItem(val)
	}

	return true
}

func (c *LRU) Len() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.items)
}

func (c *LRU) limitQueueCapacity() {
	if element := c.queue.Back(); element != nil {
		c.deleteItem(element)
	}
}

func (c *LRU) deleteItem(element *list.Element) {
	item := c.queue.Remove(element).(*Item)
	delete(c.items, item.Key)
}
