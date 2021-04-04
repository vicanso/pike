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

func TestFillMongoStoreOptions(t *testing.T) {
	assert := assert.New(t)

	ms := &mongoStore{}
	fillMongoStoreOptions("", ms)
	assert.Equal(defaultMongoDatabase, ms.db)
	assert.Equal(3*time.Second, ms.timeout)

	ms = &mongoStore{}
	fillMongoStoreOptions("mongodb://localhost:27017", ms)
	assert.Equal(defaultMongoDatabase, ms.db)
	assert.Equal(3*time.Second, ms.timeout)

	ms = &mongoStore{}
	fillMongoStoreOptions("mongodb://localhost:27017/abc?timeout=5s", ms)
	assert.Equal("abc", ms.db)
	assert.Equal(5*time.Second, ms.timeout)
}

func TestNewMongoStore(t *testing.T) {
	assert := assert.New(t)
	store, err := newMongoStore("mongodb://localhost:27017/pike")
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
	store.Close()
}
