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
	// Close close the store
	Close() error
}

var ErrNotFound = errors.New("Not found")

var stores = sync.Map{}

// NewStore create a new store
func NewStore(storeURL string) (store Store, err error) {
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
	}
	value, ok := stores.Load(storeURL)
	if ok {
		value.(Store).Close()
	}
	stores.Store(storeURL, store)
	return
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
		log.Default().Error("close stores fail",
			zap.Error(he),
		)
		return he
	}
	return nil
}
