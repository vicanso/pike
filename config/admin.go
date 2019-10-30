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

// Admin config of pike

package config

// Admin admin config
type Admin struct {
	Prefix   string `yaml:"prefix,omitempty" json:"prefix,omitempty"`
	User     string `yaml:"user,omitempty" json:"user,omitempty"`
	Password string `yaml:"password,omitempty" json:"password,omitempty"`
}

// Fetch fetch admin config
func (admin *Admin) Fetch() (err error) {
	err = fetchConfig(admin, defaultAdminKey)
	if err != nil {
		return
	}
	if admin.Prefix == "" {
		admin.Prefix = defaultAdminPrefix
	}
	return
}

// Save save admin config
func (admin *Admin) Save() (err error) {
	err = saveConfig(admin, defaultAdminKey)
	return
}

// Delete delete admin config
func (admin *Admin) Delete() (err error) {
	return deleteConfig(defaultAdminKey)
}
