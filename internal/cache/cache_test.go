package cache

import (
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func wrap(vs ...interface{}) []interface{} {
	return vs
}

var cacheDir = ".cache"

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func checkSetItem(t *testing.T, c Cache, key Key, expected bool) {
	checkFileInDir(t, c, key, expected)
	wasInCache, err := c.Set(key, string(key))
	require.NoError(t, err, err)
	if expected {
		require.True(t, wasInCache)
	} else {
		require.False(t, wasInCache)
	}
}

func checkGetItem(t *testing.T, c Cache, key Key, expected bool) {
	if expected {
		require.Equal(t, []interface{}{string(key), true}, wrap(c.Get(key)))
	} else {
		require.Equal(t, []interface{}{nil, false}, wrap(c.Get(key)))
	}
	checkFileInDir(t, c, key, expected)
}

func checkFileInDir(t *testing.T, c Cache, key Key, expected bool) {
	_, err := os.Stat(filepath.Join(c.GetDir(), string(key)))
	if expected {
		require.NoError(t, err)
	} else {
		require.True(t, os.IsNotExist(err))
	}
}

func TestCache(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		c, err := NewCache(10, cacheDir)
		require.NoError(t, err, err)

		checkGetItem(t, c, "aaa", false)
		checkGetItem(t, c, "bbb", false)
		c.Clear()
	})
	t.Run("simple", func(t *testing.T) {
		c, err := NewCache(5, cacheDir)
		require.NoError(t, err, err)

		checkSetItem(t, c, "aaa", false)
		checkSetItem(t, c, "bbb", false)

		checkGetItem(t, c, "aaa", true)
		checkGetItem(t, c, "bbb", true)

		checkSetItem(t, c, "aaa", true)

		checkGetItem(t, c, "aaa", true)
		checkGetItem(t, c, "ccc", false)

		c.Clear()
	})
	t.Run("purge logic", func(t *testing.T) {
		c, err := NewCache(10, cacheDir)
		require.NoError(t, err, err)

		checkSetItem(t, c, "aaa", false)
		checkSetItem(t, c, "bbb", false)
		checkSetItem(t, c, "ccc", false)

		// Check values in the cache
		checkGetItem(t, c, "aaa", true)
		checkGetItem(t, c, "bbb", true)
		checkGetItem(t, c, "ccc", true)

		c.Clear()

		// Check values in the cache
		checkGetItem(t, c, "aaa", false)
		checkGetItem(t, c, "bbb", false)
		checkGetItem(t, c, "ccc", false)
	})
}

func TestCacheCapacity(t *testing.T) {
	t.Run("check cache capacity", func(t *testing.T) {
		c, err := NewCache(4, cacheDir)
		require.NoError(t, err, err)

		checkSetItem(t, c, "aaa", false)
		checkSetItem(t, c, "bbb", false)
		checkSetItem(t, c, "ccc", false)
		checkSetItem(t, c, "ddd", false)
		checkSetItem(t, c, "eee", false)

		// Item should be removed from the cache
		checkGetItem(t, c, "aaa", false)
		c.Clear()
	})
}

func TestCacheMultithreading(t *testing.T) {
	c, err := NewCache(10, cacheDir)
	require.NoError(t, err, err)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			_, _ = c.Set(Key(strconv.Itoa(i)), i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Get(Key(strconv.Itoa(rand.Intn(1_000_000))))
		}
	}()

	wg.Wait()
	c.Clear()
}
