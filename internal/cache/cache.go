package cache

import (
	"crypto/sha512"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"

	utils "github.com/dmitryt/image-previewer/internal/utils"
)

var (
	ErrIncorrectFilePath = errors.New("incorrect file path")
)

type Key string

type Cache interface {
	Set(key Key, value interface{}) (bool, error)
	Get(key Key) (interface{}, bool)
	GetKey(utils.URLParams) Key
	GetDir() string
	GetFile(utils.URLParams) (*os.File, error)
	GetFilePath(utils.URLParams) string
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
		if !info.IsDir() && !strings.HasPrefix(filepath.Base(path), ".") {
			c.Set(Key(filepath.Base(path)), path)
		}
		return nil
	})
}

func (c *lruCache) GetFilePath(up utils.URLParams) string {
	return filepath.Join(c.GetDir(), string(c.GetKey(up)))
}

func (c *lruCache) GetFile(up utils.URLParams) (*os.File, error) {
	return os.Open(c.GetFilePath(up))
}

func (c *lruCache) GetKey(up utils.URLParams) Key {
	h := sha512.New()
	_, _ = io.WriteString(h, fmt.Sprintf("%s/%dx%d", up.ExternalURL, up.Width, up.Height))
	return Key([]rune(fmt.Sprintf("%x", h.Sum(nil)))[0:64])
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

func (c *lruCache) AddFile(value interface{}) (err error) {
	fpath, ok := value.(string)
	if !ok {
		return ErrIncorrectFilePath
	}
	f, err := os.Create(filepath.Join(c.dir, fpath))
	if err != nil {
		return err
	}
	log.Debug().Msgf("created file %s", fpath)
	defer f.Close()
	return nil
}

func (c *lruCache) RemoveFile(value interface{}) (err error) {
	fpath, ok := value.(string)
	if !ok {
		return ErrIncorrectFilePath
	}
	err = os.Remove(filepath.Join(c.dir, fpath))
	if err != nil {
		return err
	}
	log.Debug().Msgf("removed file %s", fpath)
	return nil
}

func (c *lruCache) Set(key Key, value interface{}) (found bool, err error) {
	c.mux.Lock()
	defer c.mux.Unlock()
	item, found := c.items[key]
	if found {
		c.queue.Remove(item)
	} else {
		err = c.AddFile(value)
		if err != nil {
			return
		}
	}
	c.items[key] = c.queue.PushFront(cacheItem{key: key, value: value})
	if c.queue.Len() > c.capacity {
		itemToRemove := c.queue.Back()
		delete(c.items, itemToRemove.Value.(cacheItem).key)
		c.queue.Remove(itemToRemove)
		err = c.RemoveFile(value)
		if err != nil {
			return
		}
	}
	return
}

func (c *lruCache) Clear() {
	c.queue = nil
	c.items = nil
	os.RemoveAll(c.dir)
}
