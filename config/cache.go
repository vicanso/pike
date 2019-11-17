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

// Cache cache config
type Cache struct {
	Name        string `yaml:"-" json:"name,omitempty" valid:"xName"`
	Zone        int    `yaml:"zone,omitempty" json:"zone,omitempty" valid:"numeric,range(1|10000)"`
	Size        int    `yaml:"size,omitempty" json:"size,omitempty" valid:"numeric,range(1|10000)"`
	HitForPass  int    `yaml:"hitForPass,omitempty" json:"hitForPass,omitempty" valid:"numeric,range(1|3600)"`
	Description string `yaml:"description,omitempty" json:"description,omitempty" valid:"-"`
}

// Caches cache configs
type Caches []*Cache

// Fetch fetch cache config
func (c *Cache) Fetch() (err error) {
	return fetchConfig(c, CachesCategory, c.Name)
}

// Save save ccache config
func (c *Cache) Save() (err error) {
	return saveConfig(c, CachesCategory, c.Name)
}

// Delete delete compress config
func (c *Cache) Delete() (err error) {
	return deleteConfig(CachesCategory, c.Name)
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

// GetCaches get all config config
func GetCaches() (caches Caches, err error) {
	keys, err := listKeysExcludePrefix(CachesCategory)
	if err != nil {
		return
	}
	caches = make(Caches, 0, len(keys))
	for _, key := range keys {
		c := &Cache{
			Name: key,
		}
		err = c.Fetch()
		if err != nil {
			return
		}
		caches = append(caches, c)
	}
	return
}
