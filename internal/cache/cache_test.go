package cache

import (
	"path/filepath"
	"math/rand"
	"strconv"
	"sync"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func wrap(vs ...interface{}) []interface{} {
	return vs
}

var cacheDir = ".cache"

func checkSetItem(t *testing.T, c Cache, key Key, value interface{}, expected bool) {
	wasInCache, err := c.Set(key, value)
	require.NoError(t, err, err)
	if expected {
		require.True(t, wasInCache)
	} else {
		require.False(t, wasInCache)
	}
	checkFileInDir(t, c, key, expected)
}

func checkGetItem(t *testing.T, c Cache, key Key, expected interface{}) {
	if expected != nil {
		require.Equal(t, []interface{}{expected, true}, wrap(c.Get(key)))
	} else {
		require.Equal(t, []interface{}{nil, false}, wrap(c.Get(key)))
	}
	checkFileInDir(t, c, key, expected != nil)
}

func checkFileInDir(t *testing.T, c Cache, key Key, expected bool) {
	_, err := os.Stat(filepath.Join(c.GetDir(), string(key)))
	if expected {
		require.True(t, os.IsExist(err))
	} else {
		require.True(t, os.IsNotExist(err))
	}
}

func TestCache(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		c, err := NewCache(10, cacheDir)
		require.NoError(t, err, err)

		checkGetItem(t, c, "aaa", nil)
		checkGetItem(t, c, "bbb", nil)
		c.Clear()
	})

	t.Run("simple", func(t *testing.T) {
		c, err := NewCache(5, cacheDir)
		require.NoError(t, err, err)

		checkSetItem(t, c, "aaa", "aaav", false)
		checkSetItem(t, c, "bbb", "bbbv", false)

		checkGetItem(t, c, "aaa", "aaav")
		checkGetItem(t, c, "bbb", "bbbv")

		checkSetItem(t, c, "aaa", "aaav", true)

		checkGetItem(t, c, "aaa", "aaav")
		checkGetItem(t, c, "ccc", nil)

		// c.Clear()
	})

	// t.Run("purge logic", func(t *testing.T) {
	// 	c, err := NewCache(10, cacheDir)
	// 	require.NoError(t, err, err)

	// 	checkSetItem(t, c, "aaa", "aaav", false)
	// 	checkSetItem(t, c, "bbb", "bbbv", false)
	// 	checkSetItem(t, c, "ccc", "cccv", false)

	// 	// Check values in the cache
	// 	checkGetItem(t, c, "aaa", "aaav")
	// 	checkGetItem(t, c, "bbb", "bbbv")
	// 	checkGetItem(t, c, "ccc", "cccv")

	// 	c.Clear()

	// 	// Check values in the cache
	// 	checkGetItem(t, c, "aaa", nil)
	// 	checkGetItem(t, c, "bbb", nil)
	// 	checkGetItem(t, c, "ccc", nil)
	// })

	// t.Run("check cache capacity", func(t *testing.T) {
	// 	c, err := NewCache(4, cacheDir)
	// 	require.NoError(t, err, err)

	// 	checkSetItem(t, c, "aaa", "aaav", false)
	// 	checkSetItem(t, c, "bbb", "bbbv", false)
	// 	checkSetItem(t, c, "ccc", "cccv", false)
	// 	checkSetItem(t, c, "ddd", "dddv", false)
	// 	checkSetItem(t, c, "eee", "eeev", false)

	// 	// Item should be removed from the cache
	// 	checkGetItem(t, c, "aaa", nil)
	// 	c.Clear()
	// })
}

func TestCacheMultithreading(t *testing.T) {
	c, err := NewCache(10, cacheDir)
	require.NoError(t, err, err)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Set(Key(strconv.Itoa(i)), i)
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