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
	"path"
	"path/filepath"
	"strings"

	"github.com/go-yaml/yaml"
)

var (
	errKeyIsNil = errors.New("key can't be nil")
)

const (
	defaultAdminPrefix = "/pike"

	// ServersCategory servers category
	ServersCategory = "servers"
	// CompressesCategory compresses category
	CompressesCategory = "compresses"
	// CachesCategory caches category
	CachesCategory = "caches"
	// UpstreamsCategory upstreams category
	UpstreamsCategory = "upstreams"
	// LocationsCategory locations category
	LocationsCategory = "locations"
	// CertsCategory cert category
	CertsCategory = "certs"
	// AdminCategory admin category
	AdminCategory = "admin"
)

// IConfig config interface
type IConfig interface {
	Fetch() error
	Save() error
	Delete() error
}

// ChangeType change key's type
type ChangeType int

const (
	// UnknownChange unknown change
	UnknownChange ChangeType = iota
	// ServerChange server's config change
	ServerChange
	// CompressChange compress's config change
	CompressChange
	// CacheChange cache's config change
	CacheChange
	// UpstreamChange upstream's config change
	UpstreamChange
	// LocationChange location's config change
	LocationChange
	// AdminChange admin's config chage
	AdminChange
)

type (
	// OnChange config change's event handler
	OnChange func(ChangeType, string)
	Config   struct {
		changeTypeKeyMap map[ChangeType]string
		basePath         string
		client           Client
		events           []OnChange
	}
)

func NewTestConfig() *Config {
	cfg, _ := NewConfig("etcd://127.0.0.1:2379/test-pike")
	return cfg
}

// NewConfig new a config instance
func NewConfig(configPath string) (cfg *Config, err error) {
	basePath := "/" + path.Base(configPath)
	if len(basePath) <= 1 {
		err = errors.New("path of config can't be null")
		return
	}
	var configClient Client
	if strings.HasPrefix(configPath, "etcd://") {
		etcdClient, err := NewEtcdClient(configPath)
		if err != nil {
			return nil, err
		}
		configClient = etcdClient
	} else {
		badgerClient, err := NewBadgerClient(configPath)
		if err != nil {
			return nil, err
		}
		configClient = badgerClient
	}
	changeTypeKeyMap := make(map[ChangeType]string)
	changeTypeKeyMap[ServerChange] = filepath.Join(basePath, ServersCategory)
	changeTypeKeyMap[CompressChange] = filepath.Join(basePath, CompressesCategory)
	changeTypeKeyMap[CacheChange] = filepath.Join(basePath, CachesCategory)
	changeTypeKeyMap[UpstreamChange] = filepath.Join(basePath, UpstreamsCategory)
	changeTypeKeyMap[LocationChange] = filepath.Join(basePath, LocationsCategory)
	changeTypeKeyMap[AdminChange] = filepath.Join(basePath, AdminCategory)
	cfg = &Config{
		client:           configClient,
		basePath:         basePath,
		changeTypeKeyMap: changeTypeKeyMap,
	}
	go configClient.Watch(basePath, func(key string) {
		for t, prefix := range changeTypeKeyMap {
			if strings.HasPrefix(key, prefix) {
				value := ""
				if len(key) > len(prefix) {
					value = key[len(prefix)+1:]
				}
				for _, onChange := range cfg.events {
					onChange(t, value)
				}
			}
		}
	})
	return
}

func (cfg *Config) getKey(elem ...string) (string, error) {
	for _, item := range elem {
		if item == "" {
			return "", errKeyIsNil
		}
	}
	arr := []string{
		cfg.basePath,
	}
	arr = append(arr, elem...)
	return filepath.Join(arr...), nil
}

func (cfg *Config) fetchConfig(v interface{}, keys ...string) (err error) {
	key, err := cfg.getKey(keys...)
	if err != nil {
		return
	}
	data, err := cfg.client.Get(key)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(data, v)
	return
}

func (cfg *Config) saveConfig(v interface{}, keys ...string) (err error) {
	key, err := cfg.getKey(keys...)
	if err != nil {
		return
	}
	data, err := yaml.Marshal(v)
	if err != nil {
		return
	}
	err = cfg.client.Set(key, data)
	return
}

func (cfg *Config) deleteConfig(keys ...string) (err error) {
	key, err := cfg.getKey(keys...)
	if err != nil {
		return
	}
	return cfg.client.Delete(key)
}

func (cfg *Config) listKeys(keyPath string) ([]string, error) {
	key, err := cfg.getKey(keyPath)
	if err != nil {
		return nil, err
	}
	return cfg.client.List(key)
}

func (cfg *Config) listKeysExcludePrefix(keyPath string) ([]string, error) {
	key, err := cfg.getKey(keyPath)
	if err != nil {
		return nil, err
	}
	keys, err := cfg.client.List(key)
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

// Watch watch config change
func (cfg *Config) Watch(onChange OnChange) {
	if cfg.events == nil {
		cfg.events = make([]OnChange, 0)
	}
	cfg.events = append(cfg.events, onChange)
}

// Close close config client
func (cfg *Config) Close() error {
	return cfg.client.Close()
}

// GetAdmin get admin config
func (cfg *Config) GetAdmin() (*Admin, error) {
	admin := new(Admin)
	admin.cfg = cfg
	err := admin.Fetch()
	return admin, err
}

// GetCaches get all config config
func (cfg *Config) GetCaches() (caches Caches, err error) {
	keys, err := cfg.listKeysExcludePrefix(CachesCategory)
	if err != nil {
		return
	}
	caches = make(Caches, 0, len(keys))
	for _, key := range keys {
		c := &Cache{
			Name: key,
			cfg:  cfg,
		}
		err = c.Fetch()
		if err != nil {
			return
		}
		caches = append(caches, c)
	}
	return
}

// GetCompresses get all compress config
func (cfg *Config) GetCompresses() (compresses Compresses, err error) {
	keys, err := cfg.listKeysExcludePrefix(CompressesCategory)
	if err != nil {
		return
	}
	compresses = make(Compresses, 0, len(keys))
	for _, key := range keys {
		c := &Compress{
			Name: key,
			cfg:  cfg,
		}
		err = c.Fetch()
		if err != nil {
			return
		}
		compresses = append(compresses, c)
	}
	return
}

// GetLocations get locations
// *Location for better performance)
func (cfg *Config) GetLocations() (locations Locations, err error) {
	keys, err := cfg.listKeysExcludePrefix(LocationsCategory)
	if err != nil {
		return
	}
	locations = make(Locations, 0, len(keys))
	for _, key := range keys {
		l := &Location{
			Name: key,
			cfg:  cfg,
		}
		err = l.Fetch()
		if err != nil {
			return
		}
		locations = append(locations, l)
	}
	return
}

// GetServers get all server config
func (cfg *Config) GetServers() (servers Servers, err error) {
	keys, err := cfg.listKeysExcludePrefix(ServersCategory)
	if err != nil {
		return
	}
	servers = make(Servers, 0, len(keys))
	for _, key := range keys {
		s := &Server{
			Name: key,
			cfg:  cfg,
		}
		err = s.Fetch()
		if err != nil {
			return
		}
		servers = append(servers, s)
	}
	return
}

// GetUpstreams get all upstream config
func (cfg *Config) GetUpstreams() (upstreams Upstreams, err error) {
	keys, err := cfg.listKeysExcludePrefix(UpstreamsCategory)
	upstreams = make(Upstreams, 0, len(keys))
	for _, key := range keys {
		u := &Upstream{
			Name: key,
			cfg:  cfg,
		}
		err = u.Fetch()
		if err != nil {
			return
		}
		upstreams = append(upstreams, u)
	}
	return
}

// GetCerts get all cert config
func (cfg *Config) GetCerts() (certs Certs, err error) {
	keys, err := cfg.listKeysExcludePrefix(CertsCategory)
	if err != nil {
		return
	}
	certs = make(Certs, 0, len(keys))
	for _, key := range keys {
		c := &Cert{
			Name: key,
			cfg:  cfg,
		}
		err = c.Fetch()
		if err != nil {
			return
		}
		certs = append(certs, c)
	}
	return
}

// NewCacheConfig new cache config
func (cfg *Config) NewCacheConfig(name string) *Cache {
	return &Cache{
		Name: name,
		cfg:  cfg,
	}
}

// NewCompressConfig new compress config
func (cfg *Config) NewCompressConfig(name string) *Compress {
	return &Compress{
		Name: name,
		cfg:  cfg,
	}
}

// NewLocationConfig new location config
func (cfg *Config) NewLocationConfig(name string) *Location {
	return &Location{
		Name: name,
		cfg:  cfg,
	}
}

// NewServerConfig new server config
func (cfg *Config) NewServerConfig(name string) *Server {
	return &Server{
		Name: name,
		cfg:  cfg,
	}
}

// NewUpstreamConfig new upstream config
func (cfg *Config) NewUpstreamConfig(name string) *Upstream {
	return &Upstream{
		Name: name,
		cfg:  cfg,
	}
}

// NewAdminConfig new upstream config
func (cfg *Config) NewAdminConfig() *Admin {
	return &Admin{
		cfg: cfg,
	}
}

// NewCertConfig new cert config
func (cfg *Config) NewCertConfig(name string) *Cert {
	return &Cert{
		Name: name,
		cfg:  cfg,
	}
}
