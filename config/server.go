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

// server config

package config

import (
	"time"

	"github.com/vicanso/pike/util"
)

// Server server config
type Server struct {
	Name              string        `yaml:"-" json:"name,omitempty" valid:"xName"`
	Cache             string        `yaml:"cache,omitempty" json:"cache,omitempty" valid:"xName"`
	Compress          string        `yaml:"compress,omitempty" json:"compress,omitempty" valid:"xName"`
	Locations         []string      `yaml:"locations,omitempty" json:"locations,omitempty" valid:"xNames"`
	ETag              bool          `yaml:"eTag,omitempty" json:"eTag,omitempty" valid:"-"`
	Addr              string        `yaml:"addr,omitempty" json:"addr,omitempty" valid:"ascii,runelength(1|50)"`
	Concurrency       uint32        `yaml:"concurrency,omitempty" json:"concurrency,omitempty" valid:"-"`
	ReadTimeout       time.Duration `yaml:"readTimeout,omitempty" json:"readTimeout,omitempty" valid:"-"`
	ReadHeaderTimeout time.Duration `yaml:"readHeaderTimeout,omitempty" json:"readHeaderTimeout,omitempty" valid:"-"`
	WriteTimeout      time.Duration `yaml:"writeTimeout,omitempty" json:"writeTimeout,omitempty" valid:"-"`
	IdleTimeout       time.Duration `yaml:"idleTimeout,omitempty" json:"idleTimeout,omitempty" valid:"-"`
	MaxHeaderBytes    int           `yaml:"maxHeaderBytes,omitempty" json:"maxHeaderBytes,omitempty" valid:"-"`
	Description       string        `yaml:"description,omitempty" json:"description,omitempty" valid:"-"`
}

// Servers server list
type Servers []*Server

// Fetch fetch server config
func (s *Server) Fetch() (err error) {
	return fetchConfig(s, ServersCategory, s.Name)
}

// Save save server config
func (s *Server) Save() (err error) {
	return saveConfig(s, ServersCategory, s.Name)
}

// Delete delete server config
func (s *Server) Delete() (err error) {
	return deleteConfig(ServersCategory, s.Name)
}

// Get get server config from server list
func (servers Servers) Get(name string) (s *Server) {
	for _, item := range servers {
		if item.Name == name {
			s = item
		}
	}
	return
}

// Exists check the category of config is exists
func (servers Servers) Exists(category, name string) bool {
	for _, item := range servers {
		switch category {
		case CachesCategory:
			if item.Cache == name {
				return true
			}
		case CompressesCategory:
			if item.Compress == name {
				return true
			}
		case LocationsCategory:
			if util.ContainesString(item.Locations, name) {
				return true
			}
		}
	}
	return false
}

// GetServers get all server config
func GetServers() (servers Servers, err error) {
	keys, err := listKeysExcludePrefix(ServersCategory)
	if err != nil {
		return
	}
	servers = make(Servers, 0, len(keys))
	for _, key := range keys {
		s := &Server{
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
