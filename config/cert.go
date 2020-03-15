// Copyright 2020 tree xie
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
// cert config

package config

type Cert struct {
	cfg  *Config
	Name string `yaml:"-" json:"name,omitempty" valid:"xName"`
	Key  []byte `yaml:"key,omitempty" json:"key,omitempty"`
	Cert []byte `yaml:"cert,omitempty" json:"cert,omitempty"`
}

type Certs []*Cert

// Fetch fetch cert config
func (c *Cert) Fetch() (err error) {
	return c.cfg.fetchConfig(c, CertCategory, c.Name)
}

// Save save cert config
func (c *Cert) Save() (err error) {
	return c.cfg.saveConfig(c, CertCategory, c.Name)
}

// Delete delete cert config
func (c *Cert) Delete() (err error) {
	return c.cfg.deleteConfig(CertCategory, c.Name)
}

// Get get cert config from cert list
func (certs Certs) Get(name string) (c *Cert) {
	for _, item := range certs {
		if item.Name == name {
			c = item
		}
	}
	return
}
