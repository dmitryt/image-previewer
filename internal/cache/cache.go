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
	GetFile2(utils.URLParams) (*os.File, error)
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
			_, err = c.Set(Key(filepath.Base(path)), path)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (c *lruCache) GetFilePath(up utils.URLParams) string {
	return filepath.Join(c.GetDir(), string(c.GetKey(up)))
}

func (c *lruCache) GetFile(up utils.URLParams) (*os.File, error) {
	fpath := c.GetFilePath(up)
	log.Debug().Msgf("Getting cache file %s", fpath)
	return os.OpenFile(fpath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
}

func (c *lruCache) GetFile2(up utils.URLParams) (*os.File, error) {
	fpath := c.GetFilePath(up)
	log.Debug().Msgf("Getting cache file %s", fpath)
	return os.Open(fpath)
}

func (c *lruCache) GetKey(up utils.URLParams) Key {
	h := sha512.New()
	str := fmt.Sprintf("%s/%dx%d", up.ExternalURL, up.Width, up.Height)
	_, _ = io.WriteString(h, str)
	result := Key([]rune(fmt.Sprintf("%x", h.Sum(nil)))[0:64])
	log.Debug().Msgf("Getting cache key %s %s", str, result)
	return result
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
		err = c.AddFile(filepath.Join(c.dir, fpath))
		if err != nil {
			return
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
