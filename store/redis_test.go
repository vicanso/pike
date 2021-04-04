// MIT License

// Copyright (c) 2021 Tree Xie

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

package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRedisStore(t *testing.T) {
	assert := assert.New(t)

	store, err := newRedisStore("redis://user:pwd@127.0.0.1:6379/?db=1&timeout=5s&prefix=test")
	assert.Nil(err)
	rs := store.(*redisStore)
	assert.Equal(5*time.Second, rs.timeout)
	assert.Equal("test", rs.prefix)
	store.Close()

	store, err = newRedisStore("redis://user:pwd@127.0.0.1:6379,127.0.0.1:6380/?master=master")
	assert.Nil(err)
	rs = store.(*redisStore)
	assert.NotNil(rs.client)
	assert.Nil(rs.cluster)
	store.Close()

	store, err = newRedisStore("redis://user:pwd@127.0.0.1:6379,127.0.0.1:6380/?mode=cluster")
	assert.Nil(err)
	rs = store.(*redisStore)
	assert.NotNil(rs.cluster)
	assert.Nil(rs.client)
	store.Close()

	store, err = newRedisStore("redis://127.0.0.1:6379/")
	assert.Nil(err)
	key := []byte("key")
	value := []byte("value")
	_, err = store.Get(key)
	assert.Equal(ErrNotFound, err)

	err = store.Set(key, value, time.Second)
	assert.Nil(err)

	data, err := store.Get(key)
	assert.Nil(err)
	assert.Equal(value, data)

	err = store.Delete(key)
	assert.Nil(err)

	// 数据过期
	err = store.Set(key, value, 10*time.Millisecond)
	assert.Nil(err)
	time.Sleep(20 * time.Millisecond)
	_, err = store.Get(key)
	assert.Equal(ErrNotFound, err)

	store.Close()
}
