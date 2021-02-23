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
	"errors"
	"net/url"
	"sync"
	"time"

	"github.com/vicanso/hes"
	"github.com/vicanso/pike/log"
	"go.uber.org/zap"
)

type Store interface {
	// Get get data from store
	Get(key []byte) (data []byte, err error)
	// Set set data from store
	Set(key []byte, data []byte, ttl time.Duration) (err error)
	// Delete delete data from store
	Delete(key []byte) (err error)
	// Close close the store
	Close() error
}

var ErrNotFound = errors.New("Not found")

var stores = sync.Map{}

var newStoreLock = sync.Mutex{}

// NewStore create a new store
func NewStore(storeURL string) (store Store, err error) {
	// 保证new store只允许一个实例操作
	newStoreLock.Lock()
	defer newStoreLock.Unlock()

	// 如果该store已存在，直接返回
	store = GetStore(storeURL)
	if store != nil {
		return
	}

	// 初始化新的store
	urlInfo, err := url.Parse(storeURL)
	if err != nil {
		return
	}
	switch urlInfo.Scheme {
	case "badger":
		store, err = newBadgerStore(urlInfo.Path)
		if err != nil {
			return
		}
	case "redis":
		store, err = newRedisStore(storeURL)
		if err != nil {
			return
		}
	}
	// 保存store
	if store != nil {
		stores.Store(storeURL, store)
	}
	return
}

// GetStore get store
func GetStore(storeURL string) Store {
	value, ok := stores.Load(storeURL)
	if !ok {
		return nil
	}
	s, ok := value.(Store)
	if !ok {
		return nil
	}
	return s
}

// Close close stores
func Close() error {
	he := &hes.Error{
		Message: "close stores fail",
	}
	stores.Range(func(_ interface{}, value interface{}) bool {
		err := value.(Store).Close()
		if err != nil {
			he.Add(err)
		}
		return true
	})
	if he.IsNotEmpty() {
		// 由于close是在程序退出时调用，因此如果失败，则先输出出错日志再返回出错
		log.Default().Error("close stores fail",
			zap.Error(he),
		)
		return he
	}
	return nil
}
