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

// The client to read data of config,
// such as etcd and file.

package config

import (
	"strings"

	badger "github.com/dgraph-io/badger"
	"github.com/vicanso/pike/log"
)

// BadgerClient badger client
type BadgerClient struct {
	db        *badger.DB
	onChanges map[string]OnKeyChange
}

// NewBadgerClient new badger client
func NewBadgerClient(file string) (client *BadgerClient, err error) {
	options := badger.DefaultOptions(file)
	options = options.WithLogger(log.BadgerLogger())
	db, err := badger.Open(options)
	if err != nil {
		return
	}
	client = &BadgerClient{
		db: db,
	}
	return
}

// Get get data form badger
func (bc *BadgerClient) Get(key string) (data []byte, err error) {
	err = bc.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		valueCopy, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		data = valueCopy
		return nil
	})
	if err == badger.ErrKeyNotFound {
		err = nil
	}
	return
}

func (bc *BadgerClient) emit(key string) {
	for prefix, onChange := range bc.onChanges {
		if strings.HasPrefix(key, prefix) {
			onChange(key)
		}
	}
}

// Set set data to badger
func (bc *BadgerClient) Set(key string, data []byte) (err error) {
	err = bc.db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), data)
		return err
	})
	if err == nil {
		bc.emit(key)
	}
	return
}

// List list all key of the prefix from badger
func (bc *BadgerClient) List(prefix string) (keys []string, err error) {
	err = bc.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		pre := []byte(prefix)
		for it.Seek(pre); it.ValidForPrefix(pre); it.Next() {
			item := it.Item()
			keys = append(keys, string(item.Key()))
		}
		return nil
	})
	return
}

// Delete delete the data of key
func (bc *BadgerClient) Delete(key string) (err error) {
	err = bc.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
	if err == nil {
		bc.emit(key)
	}
	return
}

// Watch watch config change
func (bc *BadgerClient) Watch(key string, onChange OnKeyChange) {
	if bc.onChanges == nil {
		bc.onChanges = make(map[string]OnKeyChange)
	}
	bc.onChanges[key] = onChange
}

// Close close the etcd client
func (bc *BadgerClient) Close() error {
	return bc.db.Close()
}
