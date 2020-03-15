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
	cfg                   *Config
	Prefix                string `yaml:"prefix,omitempty" json:"prefix,omitempty" valid:"ascii"`
	User                  string `yaml:"user,omitempty" json:"user,omitempty" valid:"-"`
	Password              string `yaml:"password,omitempty" json:"password,omitempty" valid:"-"`
	EnabledInternetAccess bool   `yaml:"enabledInternetAccess,omitempty" json:"enabledInternetAccess,omitempty" valid:"-"`
	Description           string `yaml:"description,omitempty" json:"description,omitempty" valid:"-"`
}

// Fetch fetch admin config
func (admin *Admin) Fetch() (err error) {
	err = admin.cfg.fetchConfig(admin, AdminCategory)
	if err != nil {
		return
	}
	if admin.Prefix == "" {
		admin.Prefix = defaultAdminPrefix
	}
	// 如果未配置账号时，设置为默认允许外部访问，方便首次配置
	if admin.User == "" {
		admin.EnabledInternetAccess = true
	}
	return
}

// Save save admin config
func (admin *Admin) Save() (err error) {
	err = admin.cfg.saveConfig(admin, AdminCategory)
	return
}

// Delete delete admin config
func (admin *Admin) Delete() (err error) {
	return admin.cfg.deleteConfig(AdminCategory)
}
