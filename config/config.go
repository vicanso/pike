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
	"time"

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
)

type (
	// Admin admin config
	Admin struct {
		Prefix   string `yaml:"prefix,omitempty" json:"prefix,omitempty"`
		User     string `yaml:"user,omitempty" json:"user,omitempty"`
		Password string `yaml:"password,omitempty" json:"password,omitempty"`
	}
	// Server server config
	Server struct {
		Name               string        `yaml:"-" json:"name,omitempty"`
		Addr               string        `yaml:"Addr,omitempty" json:"Addr,omitempty"`
		Concurrency        int           `yaml:"concurrency,omitempty" json:"concurrency,omitempty"`
		EnableServerTiming bool          `yaml:"enableServerTiming,omitempty" json:"enableServerTiming,omitempty"`
		ReadTimeout        time.Duration `yaml:"readTimeout,omitempty" json:"readTimeout,omitempty"`
		ReadHeaderTimeout  time.Duration `yaml:"readHeaderTimeout,omitempty" json:"readHeaderTimeout,omitempty"`
		WriteTimeout       time.Duration `yaml:"writeTimeout,omitempty" json:"writeTimeout,omitempty"`
		IdleTimeout        time.Duration `yaml:"idleTimeout,omitempty" json:"idleTimeout,omitempty"`
		MaxHeaderBytes     int           `yaml:"maxHeaderBytes,omitempty" json:"maxHeaderBytes,omitempty"`
	}
	// Compress compress config
	Compress struct {
		Name      string `yaml:"-" json:"name,omitempty"`
		Level     int    `yaml:"level,omitempty" json:"level,omitempty"`
		MinLength int    `yaml:"minLength,omitempty" json:"minLength,omitempty"`
		Filter    string `yaml:"filter,omitempty" json:"filter,omitempty"`
	}
	// Cache cache config
	Cache struct {
		Name       string `yaml:"-" json:"name,omitempty"`
		Zone       int    `yaml:"zone,omitempty" json:"zone,omitempty"`
		Size       int    `yaml:"size,omitempty" json:"size,omitempty"`
		HitForPass int    `yaml:"hitForPass,omitempty" json:"hitForPass,omitempty"`
	}
	// UpstreamServer upstream server
	UpstreamServer struct {
		Addr   string `yaml:"addr,omitempty" json:"addr,omitempty"`
		Weight int    `yaml:"weight,omitempty" json:"weight,omitempty"`
		Backup bool   `yaml:"backup,omitempty" json:"backup,omitempty"`
	}
	// Upstream upstream config
	Upstream struct {
		HealthCheck string           `yaml:"healthCheck,omitempty" json:"healthCheck,omitempty"`
		Policy      string           `yaml:"policy,omitempty" json:"policy,omitempty"`
		Name        string           `yaml:"-" json:"name,omitempty"`
		Servers     []UpstreamServer `yaml:"servers,omitempty" json:"servers,omitempty"`
	}
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
	result := make([]string, len(keys))

	keyLength := len(key)
	for index, item := range keys {
		result[index] = item[keyLength+1:]
	}
	return result, nil
}

// Fetch fetch admin config
func (admin *Admin) Fetch() (err error) {
	err = fetchConfig(admin, defaultAdminKey)
	if err != nil {
		return
	}
	if admin.Prefix == "" {
		admin.Prefix = defaultAdminPrefix
	}
	return
}

// Save save admin config
func (admin *Admin) Save() (err error) {
	err = saveConfig(admin, defaultAdminKey)
	return
}

// Delete delete admin config
func (admin *Admin) Delete() (err error) {
	return deleteConfig(defaultAdminKey)
}

// Fetch fetch server config
func (s *Server) Fetch() (err error) {
	return fetchConfig(s, defaultServerPath, s.Name)
}

// Save save server config
func (s *Server) Save() (err error) {
	return saveConfig(s, defaultServerPath, s.Name)
}

// Delete delete server config
func (s *Server) Delete() (err error) {
	return deleteConfig(defaultServerPath, s.Name)
}

// Fetch fetch compress config
func (c *Compress) Fetch() (err error) {
	return fetchConfig(c, defaultCompressPath, c.Name)
}

// Save save compress config
func (c *Compress) Save() (err error) {
	return saveConfig(c, defaultCompressPath, c.Name)
}

// Delete delete compress config
func (c *Compress) Delete() (err error) {
	return deleteConfig(defaultCompressPath, c.Name)
}

// Fetch fetch cache config
func (c *Cache) Fetch() (err error) {
	return fetchConfig(c, defaultCachePath, c.Name)
}

// Save save ccache config
func (c *Cache) Save() (err error) {
	return saveConfig(c, defaultCachePath, c.Name)
}

// Delete delete compress config
func (c *Cache) Delete() (err error) {
	return deleteConfig(defaultCachePath, c.Name)
}

// Fetch fetch upstream config
func (u *Upstream) Fetch() (err error) {
	return fetchConfig(u, defaultUpstreamPath, u.Name)
}

// Save save upstream config
func (u *Upstream) Save() (err error) {
	return saveConfig(u, defaultUpstreamPath, u.Name)
}

// Delete delete upsteram config
func (u *Upstream) Delete() (err error) {
	return deleteConfig(defaultUpstreamPath, u.Name)
}

// GetUpstreams get all upstream config
func GetUpstreams() (upstreams []Upstream, err error) {
	keys, err := listKeysExcludePrefix(defaultUpstreamPath)
	upstreams = make([]Upstream, 0, len(keys))
	for _, key := range keys {
		u := Upstream{
			Name: key,
		}
		err = u.Fetch()
		if err != nil {
			return
		}
		upstreams = append(upstreams, u)
	}
	return
}

// GetServers get all server config
func GetServers() (servers []Server, err error) {
	keys, err := listKeysExcludePrefix(defaultServerPath)
	if err != nil {
		return
	}
	servers = make([]Server, 0, len(keys))
	for _, key := range keys {
		s := Server{
			Name: key,
		}
		err = s.Fetch()
		if err != nil {
			return
		}
		servers = append(servers, s)
	}
	return
}
