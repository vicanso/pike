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
	errServerNameIsNil   = errors.New("server's name can't be nil")
	errCompressNameIsNil = errors.New("compress's name can't be nil")
	errCacheNameIsNil    = errors.New("cache's name can't be nil")
	errUpstreamNameIsNil = errors.New("upstream's name catn't be nil")
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
		Port               int           `yaml:"port,omitempty" json:"port,omitempty"`
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
		Name    string           `yaml:"-" json:"name,omitempty"`
		Servers []UpstreamServer `yaml:"servers,omitempty" json:"servers,omitempty"`
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

func getKey(elem ...string) string {
	arr := []string{
		basePath,
	}
	arr = append(arr, elem...)
	return filepath.Join(arr...)
}

func fetchConfig(v interface{}, keys ...string) (err error) {
	data, err := configClient.Get(getKey(keys...))
	if err != nil {
		return
	}
	err = yaml.Unmarshal(data, v)
	return
}

func saveConfig(v interface{}, keys ...string) (err error) {
	data, err := yaml.Marshal(v)
	if err != nil {
		return
	}
	err = configClient.Set(getKey(keys...), data)
	return
}

func deleteConfig(key ...string) (err error) {
	return configClient.Delete(getKey(key...))
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
	if s.Name == "" {
		err = errServerNameIsNil
		return
	}
	err = fetchConfig(s, defaultServerPath, s.Name)
	return
}

// Save save server config
func (s *Server) Save() (err error) {
	if s.Name == "" {
		err = errServerNameIsNil
		return
	}
	err = saveConfig(s, defaultServerPath, s.Name)
	return
}

// Delete delete server config
func (s *Server) Delete() (err error) {
	if s.Name == "" {
		err = errServerNameIsNil
		return
	}
	return deleteConfig(defaultServerPath, s.Name)
}

// Fetch fetch compress config
func (c *Compress) Fetch() (err error) {
	if c.Name == "" {
		err = errCompressNameIsNil
		return
	}
	err = fetchConfig(c, defaultCompressPath, c.Name)
	return
}

// Save save compress config
func (c *Compress) Save() (err error) {
	if c.Name == "" {
		err = errCompressNameIsNil
		return
	}
	err = saveConfig(c, defaultCompressPath, c.Name)
	return
}

// Delete delete compress config
func (c *Compress) Delete() (err error) {
	if c.Name == "" {
		err = errCompressNameIsNil
		return
	}
	return deleteConfig(defaultCompressPath, c.Name)
}

// Fetch fetch cache config
func (c *Cache) Fetch() (err error) {
	if c.Name == "" {
		err = errCacheNameIsNil
		return
	}
	err = fetchConfig(c, defaultCachePath, c.Name)
	return
}

// Save save ccache config
func (c *Cache) Save() (err error) {
	if c.Name == "" {
		err = errCacheNameIsNil
		return
	}
	err = saveConfig(c, defaultCachePath, c.Name)
	return
}

// Delete delete compress config
func (c *Cache) Delete() (err error) {
	if c.Name == "" {
		err = errCacheNameIsNil
		return
	}
	return deleteConfig(defaultCachePath, c.Name)
}

// Fetch fetch upstream config
func (u *Upstream) Fetch() (err error) {
	if u.Name == "" {
		err = errUpstreamNameIsNil
		return
	}
	err = fetchConfig(u, defaultUpstreamPath, u.Name)
	return
}

// Save save upstream config
func (u *Upstream) Save() (err error) {
	if u.Name == "" {
		err = errUpstreamNameIsNil
		return
	}
	err = saveConfig(u, defaultUpstreamPath, u.Name)
	return
}

// Delete delete upsteram config
func (u *Upstream) Delete() (err error) {
	if u.Name == "" {
		err = errUpstreamNameIsNil
		return
	}
	return deleteConfig(defaultUpstreamPath, u.Name)
}
