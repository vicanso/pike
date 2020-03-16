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

import (
	"crypto/tls"
	"encoding/base64"
)

type Cert struct {
	cfg  *Config
	Name string `yaml:"-" json:"name,omitempty" valid:"xName"`
	Key  string `yaml:"key,omitempty" json:"key,omitempty" valid:"base64"`
	Cert string `yaml:"cert,omitempty" json:"cert,omitempty" valid:"base64"`
}

type Certs []*Cert

// Fetch fetch cert config
func (c *Cert) Fetch() (err error) {
	return c.cfg.fetchConfig(c, CertsCategory, c.Name)
}

// Save save cert config
func (c *Cert) Save() (err error) {
	key, err := base64.StdEncoding.DecodeString(c.Key)
	if err != nil {
		return
	}
	cert, err := base64.StdEncoding.DecodeString(c.Cert)
	if err != nil {
		return
	}
	_, err = tls.X509KeyPair(cert, key)
	if err != nil {
		return
	}
	return c.cfg.saveConfig(c, CertsCategory, c.Name)
}

// Delete delete cert config
func (c *Cert) Delete() (err error) {
	return c.cfg.deleteConfig(CertsCategory, c.Name)
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
