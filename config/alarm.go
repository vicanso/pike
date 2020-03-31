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

package config

// Alarm alarm config
type Alarm struct {
	cfg      *Config
	Name     string `yaml:"-" json:"name,omitempty" valid:"xName"`
	URI      string `yaml:"uri" json:"uri,omitempty" valid:"url"`
	Template string `yaml:"template,omitempty" json:"template,omitempty" valid:"json"`
}

// Alarms alarm configs
type Alarms []*Alarm

// Fetch fetch alarm config
func (a *Alarm) Fetch() (err error) {
	return a.cfg.fetchConfig(a, AlarmsCategory, a.Name)
}

// Save save alarm config
func (a *Alarm) Save() (err error) {
	return a.cfg.saveConfig(a, AlarmsCategory, a.Name)
}

// Delete delete alarm config
func (a *Alarm) Delete() (err error) {
	return a.cfg.deleteConfig(AlarmsCategory, a.Name)
}

// Get get alarm config from alarm list
func (alarms Alarms) Get(name string) (a *Alarm) {
	for _, item := range alarms {
		if item.Name == name {
			a = item
		}
	}
	return
}
