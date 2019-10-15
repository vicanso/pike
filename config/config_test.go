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

func TestGetKey(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("/pike/test", getKey("test"))
	assert.Equal("/pike/test", getKey("/test"))
	assert.Equal("/pike/test/1", getKey("test", "1"))
}

func TestAdminConfig(t *testing.T) {
	assert := assert.New(t)
	defer func() {
		new(Admin).Delete()
	}()
	admin := new(Admin)
	err := admin.Fetch()
	assert.Nil(err)
	assert.Equal(defaultAdminPrefix, admin.Prefix)
	assert.Empty(admin.User)
	assert.Empty(admin.Password)

	user := "foo"
	password := "bar"
	prefix := "/user-admin"
	admin.User = user
	admin.Prefix = prefix
	admin.Password = password
	err = admin.Save()
	assert.Nil(err)

	admin = new(Admin)
	err = admin.Fetch()
	assert.Nil(err)
	assert.Equal(user, admin.User)
	assert.Equal(password, admin.Password)
	assert.Equal(prefix, admin.Prefix)
}

func TestServerConfig(t *testing.T) {

	assert := assert.New(t)
	s := &Server{}
	err := s.Fetch()
	assert.Equal(errServerNameIsNil, err)
	err = s.Save()
	assert.Equal(errServerNameIsNil, err)
	err = s.Delete()
	assert.Equal(errServerNameIsNil, err)

	s.Name = "tiny"
	defer s.Delete()

	err = s.Fetch()
	assert.Nil(err)
	assert.Empty(s.Concurrency)
	assert.Empty(s.EnableServerTiming)
	assert.Empty(s.Port)
	assert.Empty(s.ReadTimeout)
	assert.Empty(s.ReadHeaderTimeout)
	assert.Empty(s.WriteTimeout)
	assert.Empty(s.IdleTimeout)
	assert.Empty(s.MaxHeaderBytes)

	port := 7000
	concurrency := 1000
	enableServerTiming := true
	readTimeout := 1 * time.Second
	readHeaderTimeout := 2 * time.Second
	writeTimeout := 3 * time.Second
	ideleTimeout := 4 * time.Second
	maxHeaderBytes := 10

	s.Port = port
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
	assert.Equal(port, ns.Port)
	assert.Equal(concurrency, ns.Concurrency)
	assert.Equal(enableServerTiming, ns.EnableServerTiming)
	assert.Equal(readTimeout, ns.ReadTimeout)
	assert.Equal(readHeaderTimeout, ns.ReadHeaderTimeout)
	assert.Equal(writeTimeout, ns.WriteTimeout)
	assert.Equal(ideleTimeout, ns.IdleTimeout)
	assert.Equal(maxHeaderBytes, ns.MaxHeaderBytes)
}

func TestCompressConfig(t *testing.T) {
	assert := assert.New(t)
	c := &Compress{}
	err := c.Fetch()
	assert.Equal(errCompressNameIsNil, err)
	err = c.Save()
	assert.Equal(errCompressNameIsNil, err)
	err = c.Delete()
	assert.Equal(errCompressNameIsNil, err)

	c.Name = "tiny"
	defer c.Delete()

	err = c.Fetch()
	assert.Nil(err)
	assert.Empty(c.Level)
	assert.Empty(c.MinLength)
	assert.Empty(c.Filter)

	level := 9
	minLength := 1024
	filter := "abcd"
	c.Level = level
	c.MinLength = minLength
	c.Filter = filter
	err = c.Save()
	assert.Nil(err)

	nc := &Compress{
		Name: c.Name,
	}
	err = nc.Fetch()
	assert.Nil(err)
	assert.Equal(level, nc.Level)
	assert.Equal(minLength, nc.MinLength)
	assert.Equal(filter, c.Filter)
}

func TestCacheConfig(t *testing.T) {
	assert := assert.New(t)
	c := &Cache{}
	err := c.Fetch()
	assert.Equal(errCacheNameIsNil, err)
	err = c.Save()
	assert.Equal(errCacheNameIsNil, err)
	err = c.Delete()
	assert.Equal(errCacheNameIsNil, err)

	c.Name = "cache"
	defer c.Delete()

	err = c.Fetch()
	assert.Nil(err)
	assert.Empty(c.HitForPass)
	assert.Empty(c.Zone)
	assert.Empty(c.Size)

	hitForPass := 300
	zone := 1
	size := 10
	c.HitForPass = hitForPass
	c.Zone = zone
	c.Size = size
	err = c.Save()
	assert.Nil(err)

	nc := &Cache{
		Name: c.Name,
	}
	err = nc.Fetch()
	assert.Nil(err)
	assert.Equal(hitForPass, nc.HitForPass)
	assert.Equal(zone, nc.Zone)
	assert.Equal(size, nc.Size)
}

func TestUpstreamConfig(t *testing.T) {
	assert := assert.New(t)
	us := &Upstream{}
	err := us.Fetch()
	assert.Equal(errUpstreamNameIsNil, err)
	err = us.Save()
	assert.Equal(errUpstreamNameIsNil, err)
	err = us.Delete()
	assert.Equal(errUpstreamNameIsNil, err)

	us.Name = "upstream"
	defer us.Delete()

	err = us.Fetch()
	assert.Nil(err)
	assert.Empty(us.Servers)

	upstreamServer := UpstreamServer{
		Addr:   "127.0.0.1:7000",
		Weight: 10,
		Backup: true,
	}
	us.Servers = make([]UpstreamServer, 1)
	us.Servers[0] = upstreamServer
	err = us.Save()
	assert.Nil(err)

	nus := &Upstream{
		Name: us.Name,
	}
	err = nus.Fetch()
	assert.Nil(err)
	assert.Equal(1, len(nus.Servers))
	assert.Equal(upstreamServer, nus.Servers[0])
}
