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
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetKey(t *testing.T) {
	assert := assert.New(t)
	key, err := getKey("test", "")
	assert.Equal(errKeyIsNil, err)
	assert.Empty(key)

	key, _ = getKey("test")
	assert.Equal(basePath+"/test", key)

	key, _ = getKey("/test")
	assert.Equal(basePath+"/test", key)

	key, _ = getKey("test", "1")
	assert.Equal(basePath+"/test/1", key)
}

func TestConfig(t *testing.T) {
	assert := assert.New(t)
	prefix := "foo"
	key := filepath.Join(prefix, "1")
	value := map[string]string{
		"a": "1",
	}
	// 保存当前配置
	err := saveConfig(value, key)
	assert.Nil(err)

	// 从存储中读取
	result := make(map[string]string)
	err = fetchConfig(result, key)
	assert.Nil(err)
	assert.Equal(value, result)

	// 获取当前前缀下在所有keys
	keys, err := listKeys(key)
	assert.Nil(err)
	assert.NotEmpty(keys)
	for _, item := range keys {
		assert.True(strings.HasPrefix(item, filepath.Join(basePath, key)))
	}

	// 获取当前前缀下的所有keys，并去除前缀
	keys, err = listKeysExcludePrefix(prefix)
	assert.Nil(err)
	assert.Equal(1, len(keys))
	assert.Equal("1", keys[0])

	// 删除数据并校验删除后是否为空
	err = deleteConfig(key)
	assert.Nil(err)
	keys, err = listKeys(key)
	assert.Nil(err)
	assert.Empty(keys)
}
