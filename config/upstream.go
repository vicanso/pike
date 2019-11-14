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

// upsteam server config

package config

// UpstreamServer upstream server
type UpstreamServer struct {
	Addr   string `yaml:"addr,omitempty" json:"addr,omitempty"`
	Weight int    `yaml:"weight,omitempty" json:"weight,omitempty"`
	Backup bool   `yaml:"backup,omitempty" json:"backup,omitempty"`
}

// Upstream upstream config
type Upstream struct {
	HealthCheck string           `yaml:"healthCheck,omitempty" json:"healthCheck,omitempty"`
	Policy      string           `yaml:"policy,omitempty" json:"policy,omitempty"`
	Name        string           `yaml:"-" json:"name,omitempty"`
	Servers     []UpstreamServer `yaml:"servers,omitempty" json:"servers,omitempty"`
	Description string           `yaml:"description,omitempty" json:"description,omitempty"`
}

// Upstreams upstream config list
type Upstreams []*Upstream

// Fetch fetch upstream config
func (u *Upstream) Fetch() (err error) {
	return fetchConfig(u, defaultUpstreamPath, u.Name)
}

// Save save upstream config
func (u *Upstream) Save() (err error) {
	return saveConfig(u, defaultUpstreamPath, u.Name)
}

// Delete delete upsteram config
func (u *Upstream) Delete() (err error) {
	return deleteConfig(defaultUpstreamPath, u.Name)
}

// Get get upstream config from upstream list
func (upstreams Upstreams) Get(name string) (u *Upstream) {
	for _, item := range upstreams {
		if item.Name == name {
			u = item
		}
	}
	return
}

// GetUpstreams get all upstream config
func GetUpstreams() (upstreams Upstreams, err error) {
	keys, err := listKeysExcludePrefix(defaultUpstreamPath)
	upstreams = make(Upstreams, 0, len(keys))
	for _, key := range keys {
		u := &Upstream{
			Name: key,
		}
		err = u.Fetch()
		if err != nil {
			return
		}
		upstreams = append(upstreams, u)
	}
	return
}
