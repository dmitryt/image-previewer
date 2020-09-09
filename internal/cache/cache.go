package cache

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
)

var ErrIncorrectFilePath = errors.New("incorrect file path")

type Key string

type Cache interface {
	Set(key Key, value interface{}) (bool, error)
	Get(key Key) (interface{}, bool)
	GetDir() string
	GetFile(key Key, flag int) (*os.File, error)
	GetFilePath(key Key) string
	HasFilePath(key Key) bool
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

func New(capacity int, dir string) (Cache, error) {
	cache := &lruCache{capacity: capacity, dir: dir, queue: NewList(), items: make(map[Key]*listItem)}
	err := cache.Init()

	return cache, err
}

func (c *lruCache) GetDir() string {
	return c.dir
}

func (c *lruCache) Init() error {
	// Prepare dir
	err := os.MkdirAll(c.dir, 0o755)
	if err != nil {
		return err
	}

	return filepath.Walk(c.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && !strings.HasPrefix(filepath.Base(path), ".") {
			_, err = c.Set(Key(filepath.Base(path)), filepath.Base(path))
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (c *lruCache) GetFilePath(key Key) string {
	return filepath.Join(c.GetDir(), string(key))
}

func (c *lruCache) HasFilePath(key Key) bool {
	filePath := filepath.Join(filepath.Join(c.GetDir(), string(key)))
	if _, err := os.Stat(filePath); err == nil {
		return true
	}

	return false
}

func (c *lruCache) GetFile(key Key, flag int) (*os.File, error) {
	fpath := c.GetFilePath(key)
	log.Debug().Msgf("Getting cache file %s", fpath)

	return os.OpenFile(fpath, flag, os.ModeAppend)
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

func (c *lruCache) AddFile(fpath string) (err error) {
	log.Debug().Msgf("creating file %s", fpath)
	f, err := os.Create(fpath)
	if err != nil {
		return err
	}
	log.Debug().Msgf("created file %s", fpath)
	defer f.Close()

	return nil
}

func (c *lruCache) RemoveFile(fpath string) (err error) {
	log.Debug().Msgf("removing file %s", fpath)
	err = os.Remove(filepath.Join(c.dir, fpath))
	if err != nil {
		return err
	}
	log.Debug().Msgf("removed file %s", fpath)

	return nil
}

func (c *lruCache) processAddFile(fpath string) error {
	if _, err := os.Stat(filepath.Join(c.dir, fpath)); os.IsNotExist(err) {
		err = c.AddFile(filepath.Join(c.dir, fpath))
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *lruCache) Set(key Key, value interface{}) (found bool, err error) {
	c.mux.Lock()
	defer c.mux.Unlock()
	item, found := c.items[key]
	if found {
		c.queue.Remove(item)
	} else {
		fpath, ok := value.(string)
		if !ok {
			return found, ErrIncorrectFilePath
		}
		err = c.processAddFile(fpath)
		if err != nil {
			return found, err
		}
	}
	c.items[key] = c.queue.PushFront(cacheItem{key: key, value: value})
	if c.queue.Len() > c.capacity {
		itemToRemove := c.queue.Back()
		fpath, ok := itemToRemove.Value.(cacheItem).value.(string)
		if !ok {
			return found, ErrIncorrectFilePath
		}
		err = c.RemoveFile(fpath)
		if err != nil {
			return
		}
		delete(c.items, itemToRemove.Value.(cacheItem).key)
		c.queue.Remove(itemToRemove)
	}

	return
}

func (c *lruCache) Clear() {
	c.queue = nil
	c.items = nil
	os.RemoveAll(c.dir)
}
