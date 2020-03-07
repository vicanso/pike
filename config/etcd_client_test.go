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

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vicanso/pike/util"
)

func TestEtdClient(t *testing.T) {
	assert := assert.New(t)
	name := util.RandomString(10)
	client, err := NewEtcdClient("etcd://127.0.0.1:2379")
	assert.Nil(err)

	upstream1Name := name + "/upstreams/1"
	upstream1Data := []byte("abcd")
	t.Run("set", func(t *testing.T) {
		err := client.Set(upstream1Name, upstream1Data)
		assert.Nil(err)
	})
	t.Run("get", func(t *testing.T) {
		data, err := client.Get(upstream1Name)
		assert.Nil(err)
		assert.Equal(upstream1Data, data)
	})
	t.Run("list", func(t *testing.T) {
		keys, err := client.List(name)
		assert.Nil(err)
		assert.Equal(1, len(keys))
		assert.Equal(upstream1Name, keys[0])
	})
	t.Run("delete", func(t *testing.T) {
		err := client.Delete(upstream1Name)
		assert.Nil(err)

		data, err := client.Get(upstream1Name)
		assert.Nil(err)
		assert.Empty(data)
	})
}
