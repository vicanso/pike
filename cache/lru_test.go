package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLRU(t *testing.T) {
	assert := assert.New(t)
	lru := NewLRU(10)
	key1 := "abcd"
	value1 := &HTTPCache{}
	key2 := "defg"
	value2 := &HTTPCache{}

	lru.Add(key1, value1)
	lru.Add(key1, value1)
	lru.Add(key2, value2)
	v, ok := lru.Get(key1)
	assert.True(ok)
	assert.Equal(v, value1)

	lru.RemoveOldest()
	v, _ = lru.Get(key2)
	assert.Nil(v, "oldest cache should be removed")

	lru.Remove(key1)
	v, ok = lru.Get(key1)
	assert.Nil(v, "remove cache fail")

	lru.Add(key1, value1)
	assert.Equal(lru.Len(), 1, "get lru len fail")

	count := 0
	lru.ForEach(func(key string, value *HTTPCache) {
		count++
	})
	assert.Equal(count, lru.Len(), "lru forEach fail")

	lru.Clear()
	assert.Equal(lru.Len(), 0, "lru clear fail")
}
