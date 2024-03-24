package lru

import (
	"container/list"
)

type Cache struct {
	capacity  int64
	allocated int64
	deque     *list.List
	cache     map[string]*list.Element
	OnEvicted func(key string, value Value)
}

type Value interface {
	Len() int
}

func (c *Cache) Len() int {
	return c.deque.Len()
}

type KV struct {
	key   string
	value Value
}

func New(capacity int64, onEvicted func(string, Value)) *Cache {
	return &Cache{capacity: capacity, allocated: 0, deque: list.New(), cache: make(map[string]*list.Element), OnEvicted: onEvicted}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.deque.MoveToFront(ele)
		kv := ele.Value.(*KV)
		return kv.value, true
	}
	return
}

func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		kv := ele.Value.(*KV)
		kv.value = value
		c.deque.MoveToFront(ele)
		c.allocated += int64(value.Len()) - int64(kv.value.Len())
	} else {
		ele := c.deque.PushFront(&KV{key: key, value: value})
		c.cache[key] = ele
		c.allocated += int64(value.Len()) + int64(len(key))
	}
	for c.capacity != 0 && c.capacity < c.allocated {
		c.RemoveOldest()
	}
}

func (c *Cache) RemoveOldest() {
	ele := c.deque.Back()
	if ele != nil {
		c.deque.Remove(ele)
		kv := ele.Value.(*KV)
		c.allocated -= int64(kv.value.Len()) + int64(len(kv.key))
		delete(c.cache, kv.key)
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}
