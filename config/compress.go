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

// compress config

package config

// Compress compress config
type Compress struct {
	Name      string `yaml:"-" json:"name,omitempty"`
	Level     int    `yaml:"level,omitempty" json:"level,omitempty"`
	MinLength int    `yaml:"minLength,omitempty" json:"minLength,omitempty"`
	Filter    string `yaml:"filter,omitempty" json:"filter,omitempty"`
}

// Compresses compress config list
type Compresses []*Compress

// Fetch fetch compress config
func (c *Compress) Fetch() (err error) {
	return fetchConfig(c, defaultCompressPath, c.Name)
}

// Save save compress config
func (c *Compress) Save() (err error) {
	return saveConfig(c, defaultCompressPath, c.Name)
}

// Delete delete compress config
func (c *Compress) Delete() (err error) {
	return deleteConfig(defaultCompressPath, c.Name)
}

// Get get compress config from compress list
func (compresses Compresses) Get(name string) (c *Compress) {
	for _, item := range compresses {
		if item.Name == name {
			c = item
		}
	}
	return
}

// GetCompresses get all compress config
func GetCompresses() (compresses Compresses, err error) {
	keys, err := listKeysExcludePrefix(defaultCompressPath)
	if err != nil {
		return
	}
	compresses = make(Compresses, 0, len(keys))
	for _, key := range keys {
		c := &Compress{
			Name: key,
		}
		err = c.Fetch()
		if err != nil {
			return
		}
		compresses = append(compresses, c)
	}
	return
}
