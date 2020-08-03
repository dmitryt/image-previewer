package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	GetDir() string
	Clear()
}

type lruCache struct {
	dir      string
	capacity int
	queue    List
	items    map[Key]*listItem
	mux      sync.Mutex
}

type cacheItem struct {
	key   Key
	value interface{}
}

func NewCache(capacity int, dir string) (Cache, error) {
	cache := &lruCache{capacity: capacity, dir: dir, queue: NewList(), items: make(map[Key]*listItem)}
	err := cache.Init()
	return cache, err
}

func (c *lruCache) GetDir() string {
	return c.dir
}

func (c *lruCache) Init() error {
	// Prepare dir
	err := os.MkdirAll(c.dir, 0755)
	if err != nil {
		return err
	}
	return filepath.Walk(c.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		fmt.Println(path, info.IsDir(), strings.HasPrefix(filepath.Base(path), "."))
		if !info.IsDir() && !strings.HasPrefix(filepath.Base(path), ".") {
			c.Set(Key(filepath.Base(path)), path)
		}
		return nil
	})
}

func (c *lruCache) Get(key Key) (interface{}, bool) {
	c.mux.Lock()
	defer c.mux.Unlock()
	item, found := c.items[key]
	if !found {
		return nil, false
	}
	c.queue.MoveToFront(item)
	return item.Value.(cacheItem).value, true
}

func (c *lruCache) Set(key Key, value interface{}) bool {
	c.mux.Lock()
	defer c.mux.Unlock()
	item, found := c.items[key]
	if found {
		c.queue.Remove(item)
	}
	c.items[key] = c.queue.PushFront(cacheItem{key: key, value: value})
	if c.queue.Len() > c.capacity {
		itemToRemove := c.queue.Back()
		delete(c.items, itemToRemove.Value.(cacheItem).key)
		c.queue.Remove(itemToRemove)
	}
	return found
}

func (c *lruCache) Clear() {
	c.queue = nil
	c.items = nil
}
