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

// location config

package config

import (
	"net/http"
	"sort"
	"strings"

	"github.com/vicanso/pike/util"
)

// Location location config
type Location struct {
	cfg            *Config
	Name           string      `yaml:"name,omitempty" json:"name,omitempty" valid:"xName"`
	Upstream       string      `yaml:"upstream,omitempty" json:"upstream,omitempty" valid:"xName"`
	Prefixs        []string    `yaml:"prefixs,omitempty" json:"prefixs,omitempty" valid:"xPrefixs,optional"`
	Rewrites       []string    `yaml:"rewrites,omitempty" json:"rewrites,omitempty" valid:"xRewrites,optional"`
	Hosts          []string    `yaml:"hosts,omitempty" json:"hosts,omitempty" valid:"xHosts,optional"`
	ResponseHeader []string    `yaml:"responseHeader,omitempty" json:"responseHeader,omitempty" valid:"xHeader,optional"`
	ResHeader      http.Header `yaml:"-" json:"-" valid:"-"`
	RequestHeader  []string    `yaml:"requestHeader,omitempty" json:"requestHeader,omitempty" valid:"xHeader,optional"`
	ReqHeader      http.Header `yaml:"-" json:"-" valid:"-"`
	Description    string      `yaml:"description,omitempty" json:"description,omitempty" valid:"-"`
}

// Locations locations
type Locations []*Location

// Fetch fetch location config
func (l *Location) Fetch() (err error) {
	err = l.cfg.fetchConfig(l, LocationsCategory, l.Name)
	if err != nil {
		return
	}
	l.ReqHeader = util.ConvertToHTTPHeader(l.RequestHeader)
	l.ResHeader = util.ConvertToHTTPHeader(l.ResponseHeader)
	return
}

// Save save location config
func (l *Location) Save() (err error) {
	return l.cfg.saveConfig(l, LocationsCategory, l.Name)
}

// Delete delete location config
func (l *Location) Delete() (err error) {
	return l.cfg.deleteConfig(LocationsCategory, l.Name)
}

// Match check location's hosts and prefixs match host/url
func (l *Location) Match(host, url string) bool {
	if len(l.Hosts) != 0 {
		found := false
		for _, item := range l.Hosts {
			if item == host {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if len(l.Prefixs) != 0 {
		found := false
		for _, item := range l.Prefixs {
			if strings.HasPrefix(url, item) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (l *Location) getPriority() int {
	priority := 8
	if len(l.Prefixs) != 0 {
		priority -= 4
	}
	if len(l.Hosts) != 0 {
		priority -= 2
	}
	return priority
}

// Sort sort locations
func (locations Locations) Sort() {
	sort.Slice(locations, func(i, j int) bool {
		return locations[i].getPriority() < locations[j].getPriority()
	})
}

// Get get location config from location list
func (locations Locations) Get(name string) (l *Location) {
	for _, item := range locations {
		if item.Name == name {
			l = item
		}
	}
	return
}

// GetMatch get match location
func (locations Locations) GetMatch(host, url string) (l *Location) {
	for _, item := range locations {
		if item.Match(host, url) {
			l = item
			break
		}
	}
	return
}

// Filter filter locations
func (locations Locations) Filter(filters ...string) (result Locations) {
	result = make(Locations, 0)
	for _, item := range locations {
		for _, name := range filters {
			if item.Name == name {
				result = append(result, item)
			}
		}
	}
	return
}

// ExistsUpstream check the upstream exists
func (locations Locations) ExistsUpstream(name string) bool {
	for _, item := range locations {
		if item.Upstream == name {
			return true
		}
	}
	return false
}
