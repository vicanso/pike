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
	Addr   string `yaml:"addr,omitempty" json:"addr,omitempty" valid:"url"`
	Weight int    `yaml:"weight,omitempty" json:"weight,omitempty" valid:"-"`
	Backup bool   `yaml:"backup,omitempty" json:"backup,omitempty" valid:"-"`
}

// Upstream upstream config
type Upstream struct {
	cfg         *Config
	HealthCheck string           `yaml:"healthCheck,omitempty" json:"healthCheck,omitempty" valid:"xURLPath,optional"`
	Policy      string           `yaml:"policy,omitempty" json:"policy,omitempty" valid:"-"`
	Name        string           `yaml:"-" json:"name,omitempty" valid:"xName"`
	Servers     []UpstreamServer `yaml:"servers,omitempty" json:"servers,omitempty" valid:"xServers"`
	Description string           `yaml:"description,omitempty" json:"description,omitempty" valid:"-"`
}

// Upstreams upstream config list
type Upstreams []*Upstream

// Fetch fetch upstream config
func (u *Upstream) Fetch() (err error) {
	return u.cfg.fetchConfig(u, UpstreamsCategory, u.Name)
}

// Save save upstream config
func (u *Upstream) Save() (err error) {
	return u.cfg.saveConfig(u, UpstreamsCategory, u.Name)
}

// Delete delete upsteram config
func (u *Upstream) Delete() (err error) {
	return u.cfg.deleteConfig(UpstreamsCategory, u.Name)
}

// SetClient set client
func (u *Upstream) SetClient(cfg *Config) {
	u.cfg = cfg
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
