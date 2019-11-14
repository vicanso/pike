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
	description := "location description"
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
		Prefixs:        prefixs,
		Rewrites:       rewrites,
		Hosts:          hosts,
		ResponseHeader: responseHeader,
		RequestHeader:  requestHeader,
		Description:    description,
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
	assert.Equal(prefixs, l.Prefixs)
	assert.Equal(rewrites, l.Rewrites)
	assert.Equal(hosts, l.Hosts)
	assert.Equal(responseHeader, l.ResponseHeader)
	assert.Equal(requestHeader, l.RequestHeader)
	assert.Equal(description, l.Description)

	locations, err := GetLocations()
	assert.Nil(err)
	assert.Equal(1, len(locations))

	nl := locations.Get(l.Name)
	assert.Equal(l, nl)
}

func TestMatch(t *testing.T) {
	assert := assert.New(t)
	l := Location{
		Hosts: []string{
			"aslant.site",
		},
		Prefixs: []string{
			"/api",
		},
	}
	assert.False(l.Match("tiny.aslant.site", "/"))
	assert.False(l.Match("aslant.site", "/"))
	assert.True(l.Match("aslant.site", "/api"))
}

func newLocations() Locations {
	return Locations{
		&Location{
			Name:    "api",
			Prefixs: []string{"/api"},
		},
		&Location{
			Name:  "aslant",
			Hosts: []string{"aslant.site"},
		},
		&Location{
			Name:    "tiny",
			Prefixs: []string{"/api"},
			Hosts:   []string{"tiny.aslant.site"},
		},
	}
}
func TestSort(t *testing.T) {
	assert := assert.New(t)
	ls := newLocations()
	ls.Sort()
	assert.NotEmpty(ls[0].Hosts)
	assert.NotEmpty(ls[0].Prefixs)
	assert.Empty(ls[1].Hosts)
	assert.NotEmpty(ls[1].Prefixs)
	assert.NotEmpty(ls[2].Hosts)
	assert.Empty(ls[2].Prefixs)
}

func TestGetMatch(t *testing.T) {
	assert := assert.New(t)
	ls := newLocations()
	result := ls.GetMatch("aslant.site", "/")
	assert.Equal("aslant.site", result.Hosts[0])

	result = ls.GetMatch("tiny.aslant.site", "/api")
	assert.Equal("/api", result.Prefixs[0])
}

func TestFilter(t *testing.T) {
	assert := assert.New(t)
	ls := newLocations()
	result := ls.Filter("api", "tiny")
	assert.Equal(2, len(result))
}
