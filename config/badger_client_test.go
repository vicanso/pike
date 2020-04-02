// Copyright 2020 tree xie
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

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBadger(t *testing.T) {
	assert := assert.New(t)
	client, err := NewBadgerClient(os.TempDir())
	defer func() {
		_ = client.Close()
	}()
	assert.Nil(err)

	prefix := "/prefix/"
	key := prefix + "foo"
	value := []byte("bar")

	t.Run("set", func(t *testing.T) {
		err := client.Set(key, value)
		assert.Nil(err)
	})

	t.Run("get", func(t *testing.T) {
		data, err := client.Get(key)
		assert.Nil(err)
		assert.Equal(value, data)
	})

	t.Run("list", func(t *testing.T) {
		keys, err := client.List(prefix)
		assert.Nil(err)
		assert.Equal(1, len(keys))
		assert.Equal(key, keys[0])
	})

	t.Run("delete", func(t *testing.T) {
		err := client.Delete(key)
		assert.Nil(err)
		data, err := client.Get(key)
		assert.Nil(err)
		assert.Nil(data)
	})

	t.Run("watch", func(t *testing.T) {
		changed := false
		client.Watch(prefix, func(k string) {
			changed = true
		})
		err := client.Set(key, value)
		assert.Nil(err)
		assert.True(changed)
	})

}
