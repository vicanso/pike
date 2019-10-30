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

	"github.com/stretchr/testify/assert"
)

func TestLocation(t *testing.T) {
	assert := assert.New(t)
	upstream := "testupstream"
	server := "testserver"
	cache := "testcache"
	prefixs := []string{
		"/api",
	}
	rewrites := []string{
		"/api:/$1",
	}
	hosts := []string{
		"aslant.site",
	}
	responseHeader := []string{
		"a:1",
		"b:2",
	}
	requestHeader := []string{
		"c:3",
		"d:4",
	}
	l := &Location{
		Name:           "testlocation",
		Upstream:       upstream,
		Server:         server,
		Cache:          cache,
		Prefixs:        prefixs,
		Rewrites:       rewrites,
		Hosts:          hosts,
		ResponseHeader: responseHeader,
		RequestHeader:  requestHeader,
	}
	defer l.Delete()

	err := l.Save()
	assert.Nil(err)

	l = &Location{
		Name: "testlocation",
	}
	err = l.Fetch()
	assert.Nil(err)
	assert.Equal(upstream, l.Upstream)
	assert.Equal(server, l.Server)
	assert.Equal(cache, l.Cache)
	assert.Equal(prefixs, l.Prefixs)
	assert.Equal(rewrites, l.Rewrites)
	assert.Equal(hosts, l.Hosts)
	assert.Equal(responseHeader, l.ResponseHeader)
	assert.Equal(requestHeader, l.RequestHeader)

	locations, err := GetLocations()
	assert.Nil(err)
	assert.Equal(1, len(locations))

	nl := locations.Get(l.Name)
	assert.Equal(l, nl)
}
