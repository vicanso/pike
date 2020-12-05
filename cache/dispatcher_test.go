// MIT License

// Copyright (c) 2020 Tree Xie

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLRUGetCache(t *testing.T) {
	assert := assert.New(t)
	httpLRU := newHTTPLRUCache(10)
	key := []byte("abcd")
	c, ok := httpLRU.getCache(key)
	assert.False(ok)
	assert.Nil(c)

	httpLRU.cache.Add(string(key), "abc")
	c, ok = httpLRU.getCache(key)
	assert.False(ok)
	assert.Nil(c)

	hc := &httpCache{}
	httpLRU.cache.Add(string(key), hc)
	c, ok = httpLRU.getCache(key)
	assert.True(ok)
	assert.Equal(hc, c)
}

func TestDispatcher(t *testing.T) {
	assert := assert.New(t)
	d := NewDispatcher(0, 30)
	assert.Equal(30, d.GetHitForPass())
	key := []byte("key")
	c := d.GetHTTPCache(key)
	c.createdAt = 1
	for i := 0; i < 10; i++ {
		assert.Equal(c, d.GetHTTPCache([]byte("key")))
	}
	d.RemoveHTTPCache(key)
	hc := d.GetHTTPCache(key)
	assert.NotNil(hc)
	assert.Empty(hc.createdAt)
}

func TestDispatchers(t *testing.T) {
	assert := assert.New(t)
	name1 := "test1"
	name2 := "test2"
	ds := NewDispatchers([]DispatcherOption{
		{
			Name: name1,
			Size: 100,
		},
	})
	assert.NotNil(ds.Get(name1))
	// 第一次reset，清除name1，添加name2
	ds.Reset([]DispatcherOption{
		{
			Name: name2,
			Size: 100,
		},
	})
	assert.Nil(ds.Get(name1))
	assert.NotNil(ds.Get(name2))

	// 再次reset
	ds.Reset([]DispatcherOption{
		{
			Name: name2,
			Size: 100,
		},
	})
	assert.Nil(ds.Get(name1))
	assert.NotNil(ds.Get(name2))

	key := []byte("abc")
	hc := ds.Get(name2).GetHTTPCache(key)
	assert.NotNil(hc)
	hc.createdAt = 1

	ds.RemoveHTTPCache("", key)
	hc1 := ds.Get(name2).GetHTTPCache(key)
	assert.Empty(hc1.createdAt)

}
