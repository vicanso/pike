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

import "time"

// Server server config
type Server struct {
	Name              string        `yaml:"-" json:"name,omitempty"`
	Cache             string        `yaml:"cache,omitempty" json:"cache,omitempty"`
	Compress          string        `yaml:"compress,omitempty" json:"compress,omitempty"`
	Locations         []string      `yaml:"locations,omitempty" json:"locations,omitempty"`
	ETag              bool          `yaml:"eTag,omitempty" json:"eTag,omitempty"`
	Addr              string        `yaml:"addr,omitempty" json:"addr,omitempty"`
	Concurrency       uint32        `yaml:"concurrency,omitempty" json:"concurrency,omitempty"`
	ReadTimeout       time.Duration `yaml:"readTimeout,omitempty" json:"readTimeout,omitempty"`
	ReadHeaderTimeout time.Duration `yaml:"readHeaderTimeout,omitempty" json:"readHeaderTimeout,omitempty"`
	WriteTimeout      time.Duration `yaml:"writeTimeout,omitempty" json:"writeTimeout,omitempty"`
	IdleTimeout       time.Duration `yaml:"idleTimeout,omitempty" json:"idleTimeout,omitempty"`
	MaxHeaderBytes    int           `yaml:"maxHeaderBytes,omitempty" json:"maxHeaderBytes,omitempty"`
}

// Servers server list
type Servers []*Server

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

// Get get server config from server list
func (servers Servers) Get(name string) (s *Server) {
	for _, item := range servers {
		if item.Name == name {
			s = item
		}
	}
	return
}

// GetServers get all server config
func GetServers() (servers Servers, err error) {
	keys, err := listKeysExcludePrefix(defaultServerPath)
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
