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
// cache config

package config

import "github.com/robfig/cron/v3"

// Cache cache config
type Cache struct {
	cfg         *Config
	Name        string `yaml:"-" json:"name,omitempty" valid:"xName"`
	Zone        int    `yaml:"zone,omitempty" json:"zone,omitempty" valid:"numeric,range(1|10000)"`
	Size        int    `yaml:"size,omitempty" json:"size,omitempty" valid:"numeric,range(1|10000)"`
	HitForPass  int    `yaml:"hitForPass,omitempty" json:"hitForPass,omitempty" valid:"numeric,range(1|3600)"`
	PurgedAt    string `yaml:"purgedAt,omitempty" json:"purgedAt,omitempty" valid:"-"`
	Description string `yaml:"description,omitempty" json:"description,omitempty" valid:"-"`
}

// Caches cache configs
type Caches []*Cache

// Fetch fetch cache config
func (c *Cache) Fetch() (err error) {
	return c.cfg.fetchConfig(c, CachesCategory, c.Name)
}

// Save save ccache config
func (c *Cache) Save() (err error) {
	if c.PurgedAt != "" {
		_, err = cron.New().AddFunc(c.PurgedAt, nil)
		if err != nil {
			return
		}
	}
	return c.cfg.saveConfig(c, CachesCategory, c.Name)
}

// Delete delete cache config
func (c *Cache) Delete() (err error) {
	return c.cfg.deleteConfig(CachesCategory, c.Name)
}

// Get get cache config from cache list
func (caches Caches) Get(name string) (c *Cache) {
	for _, item := range caches {
		if item.Name == name {
			c = item
		}
	}
	return
}
