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

	assert := assert.New(t)
	s := &Server{
		Name: "tiny",
	}
	defer s.Delete()

	err := s.Fetch()
	assert.Nil(err)
	assert.Empty(s.Concurrency)
	assert.Empty(s.EnableServerTiming)
	assert.Empty(s.Addr)
	assert.Empty(s.ReadTimeout)
	assert.Empty(s.ReadHeaderTimeout)
	assert.Empty(s.WriteTimeout)
	assert.Empty(s.IdleTimeout)
	assert.Empty(s.MaxHeaderBytes)

	addr := ":7000"
	concurrency := 1000
	enableServerTiming := true
	readTimeout := 1 * time.Second
	readHeaderTimeout := 2 * time.Second
	writeTimeout := 3 * time.Second
	ideleTimeout := 4 * time.Second
	maxHeaderBytes := 10

	s.Addr = addr
	s.Concurrency = concurrency
	s.EnableServerTiming = enableServerTiming
	s.ReadTimeout = readTimeout
	s.ReadHeaderTimeout = readHeaderTimeout
	s.WriteTimeout = writeTimeout
	s.IdleTimeout = ideleTimeout
	s.MaxHeaderBytes = maxHeaderBytes
	err = s.Save()
	assert.Nil(err)

	ns := &Server{
		Name: s.Name,
	}
	err = ns.Fetch()
	assert.Nil(err)
	assert.Equal(addr, ns.Addr)
	assert.Equal(concurrency, ns.Concurrency)
	assert.Equal(enableServerTiming, ns.EnableServerTiming)
	assert.Equal(readTimeout, ns.ReadTimeout)
	assert.Equal(readHeaderTimeout, ns.ReadHeaderTimeout)
	assert.Equal(writeTimeout, ns.WriteTimeout)
	assert.Equal(ideleTimeout, ns.IdleTimeout)
	assert.Equal(maxHeaderBytes, ns.MaxHeaderBytes)

	servers, err := GetServers()
	assert.Nil(err)
	assert.Equal(1, len(servers))

	ns = servers.Get(s.Name)
	assert.Equal(s, ns)
}
