// Copyright 2019 tree xie
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusLRU(t *testing.T) {
	assert := assert.New(t)
	lru := NewHTTPCacheLRU(10)
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
	assert.False(ok)
	assert.Nil(v, "remove cache fail")

	lru.Add(key1, value1)
	assert.Equal(1, lru.Len(), "get lru len fail")

	count := 0
	lru.ForEach(func(key string, value *HTTPCache) {
		count++
	})
	assert.Equal(count, lru.Len(), "lru forEach fail")

	lru.Clear()
	assert.Equal(0, lru.Len(), "lru clear fail")

	v = lru.FindOrCreate(key1)
	assert.NotNil(v)
}
