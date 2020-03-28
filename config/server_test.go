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
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServerConfig(t *testing.T) {
	cfg := NewTestConfig()

	assert := assert.New(t)
	s := &Server{
		cfg:  cfg,
		Name: "tiny",
	}
	defer func() {
		_ = s.Delete()
	}()

	err := s.Fetch()
	assert.Nil(err)
	assert.Empty(s.Concurrency)
	assert.Empty(s.Addr)
	assert.Empty(s.ReadTimeout)
	assert.Empty(s.ReadHeaderTimeout)
	assert.Empty(s.WriteTimeout)
	assert.Empty(s.IdleTimeout)
	assert.Empty(s.MaxHeaderBytes)

	addr := ":7000"
	cache := "tiny"
	compress := "gzipBr9"
	description := "server description"
	locations := []string{
		"test",
		"ip2location",
	}
	certs := []string{
		"me.dev",
	}
	concurrency := uint32(1000)
	readTimeout := 1 * time.Second
	readHeaderTimeout := 2 * time.Second
	writeTimeout := 3 * time.Second
	ideleTimeout := 4 * time.Second
	maxHeaderBytes := 10

	s.Addr = addr
	s.Cache = cache
	s.Compress = compress
	s.Locations = locations
	s.Concurrency = concurrency
	s.ReadTimeout = readTimeout
	s.ReadHeaderTimeout = readHeaderTimeout
	s.WriteTimeout = writeTimeout
	s.IdleTimeout = ideleTimeout
	s.MaxHeaderBytes = maxHeaderBytes
	s.Description = description
	s.Certs = certs
	err = s.Save()
	assert.Nil(err)

	ns := &Server{
		cfg:  cfg,
		Name: s.Name,
	}
	err = ns.Fetch()
	assert.Nil(err)
	assert.Equal(addr, ns.Addr)
	assert.Equal(cache, ns.Cache)
	assert.Equal(compress, ns.Compress)
	assert.Equal(locations, ns.Locations)
	assert.Equal(concurrency, ns.Concurrency)
	assert.Equal(readTimeout, ns.ReadTimeout)
	assert.Equal(readHeaderTimeout, ns.ReadHeaderTimeout)
	assert.Equal(writeTimeout, ns.WriteTimeout)
	assert.Equal(ideleTimeout, ns.IdleTimeout)
	assert.Equal(maxHeaderBytes, ns.MaxHeaderBytes)
	assert.Equal(description, ns.Description)

	servers, err := cfg.GetServers()
	assert.Nil(err)
	assert.Equal(1, len(servers))

	ns = servers.Get(s.Name)
	assert.Equal(s, ns)

	assert.True(servers.Exists(CachesCategory, cache))
	assert.False(servers.Exists(CachesCategory, cache+"1"))

	assert.True(servers.Exists(CompressesCategory, compress))
	assert.False(servers.Exists(CompressesCategory, compress+"1"))

	assert.True(servers.Exists(LocationsCategory, locations[0]))
	assert.False(servers.Exists(LocationsCategory, locations[0]+"1"))

	assert.True(servers.Exists(CertsCategory, certs[0]))
}
