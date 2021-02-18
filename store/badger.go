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
	"fmt"
	"strings"
	"time"

	badger "github.com/dgraph-io/badger/v3"
	"github.com/vicanso/pike/log"
	"go.uber.org/zap"
)

type badgerStore struct {
	db *badger.DB
}

type badgerLogger struct{}

func (bl *badgerLogger) Errorf(format string, args ...interface{}) {
	msg := strings.TrimSpace(fmt.Sprintf(format, args...))
	log.Default().Error(msg,
		zap.String("category", "badger"),
	)
}
func (bl *badgerLogger) Warningf(format string, args ...interface{}) {
	msg := strings.TrimSpace(fmt.Sprintf(format, args...))
	log.Default().Warn(msg,
		zap.String("category", "badger"),
	)
}
func (bl *badgerLogger) Infof(format string, args ...interface{}) {
	msg := strings.TrimSpace(fmt.Sprintf(format, args...))
	log.Default().Info(msg,
		zap.String("category", "badger"),
	)
}
func (bl *badgerLogger) Debugf(format string, args ...interface{}) {
	msg := strings.TrimSpace(fmt.Sprintf(format, args...))
	log.Default().Debug(msg,
		zap.String("category", "badger"),
	)
}

// newBadgerStore create a new badger store
func newBadgerStore(path string) (*badgerStore, error) {
	options := badger.DefaultOptions(path)
	options.Logger = &badgerLogger{}
	db, err := badger.Open(options)
	if err != nil {
		return nil, err
	}
	return &badgerStore{
		db: db,
	}, nil
}

// Get get data from badger
func (bs *badgerStore) Get(key []byte) (data []byte, err error) {
	err = bs.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				err = ErrNotFound
			}
			return err
		}
		return item.Value(func(val []byte) error {
			data = append([]byte{}, val...)
			return nil
		})
	})
	if err != nil {
		return
	}
	return
}

// Set set data to badger
func (bs *badgerStore) Set(key []byte, data []byte, ttl time.Duration) (err error) {
	return bs.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry(key, data).
			WithTTL(ttl)
		return txn.SetEntry(e)
	})
}

// Delete delete data from store
func (bs *badgerStore) Delete(key []byte) (err error) {
	return bs.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

// Close close badger
func (bs *badgerStore) Close() error {
	return bs.db.Close()
}
