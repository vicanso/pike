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

// The config of pike, include admin, server and upstreams.

package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-yaml/yaml"
)

var (
	basePath string

	configClient Client
)

var (
	errKeyIsNil = errors.New("key can't be nil")
)

const (
	defaultBasePath = "/pike"

	defaultAdminKey    = "admin"
	defaultAdminPrefix = "/admin"

	defaultServerPath   = "servers"
	defaultCompressPath = "compresses"
	defaultCachePath    = "caches"
	defaultUpstreamPath = "upstreams"
	defaultLocationPath = "locations"
)

func init() {
	basePath = os.Getenv("BASE_PATH")
	if basePath == "" {
		basePath = defaultBasePath
	}
	configPath := os.Getenv("CONFIG")
	if configPath == "" {
		panic(errors.New("config path can't be nil"))
	}
	if strings.HasPrefix(configPath, "etcd://") {
		etcdClient, err := NewEtcdClient(configPath)
		if err != nil {
			panic(err)
		}
		configClient = etcdClient
	} else {
		// TODO 支持文件配置
	}
}

func getKey(elem ...string) (string, error) {
	for _, item := range elem {
		if item == "" {
			return "", errKeyIsNil
		}
	}
	arr := []string{
		basePath,
	}
	arr = append(arr, elem...)
	return filepath.Join(arr...), nil
}

func fetchConfig(v interface{}, keys ...string) (err error) {
	key, err := getKey(keys...)
	if err != nil {
		return
	}
	data, err := configClient.Get(key)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(data, v)
	return
}

func saveConfig(v interface{}, keys ...string) (err error) {
	key, err := getKey(keys...)
	if err != nil {
		return
	}
	data, err := yaml.Marshal(v)
	if err != nil {
		return
	}
	err = configClient.Set(key, data)
	return
}

func deleteConfig(keys ...string) (err error) {
	key, err := getKey(keys...)
	if err != nil {
		return
	}
	return configClient.Delete(key)
}

func listKeys(keyPath string) ([]string, error) {
	key, err := getKey(keyPath)
	if err != nil {
		return nil, err
	}
	return configClient.List(key)
}

func listKeysExcludePrefix(keyPath string) ([]string, error) {
	key, err := getKey(keyPath)
	if err != nil {
		return nil, err
	}
	keys, err := configClient.List(key)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(keys))

	keyLength := len(key)
	for _, item := range keys {
		if len(item) == keyLength {
			continue
		}
		result = append(result, item[keyLength+1:])
	}
	return result, nil
}
