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

// influxdb config
package config

// Influxdb influxdb config
type Influxdb struct {
	cfg       *Config
	URI       string `yaml:"uri" json:"uri,omitempty" valid:"url"`
	Bucket    string `yaml:"bucket" json:"bucket,omitempty" valid:"runelength(1|100)"`
	Org       string `yaml:"org" json:"org,omitempty" valid:"runelength(1|100)"`
	Token     string `yaml:"token" json:"token,omitempty" valid:"ascii,runelength(1|200)"`
	BatchSize int    `yaml:"batchSize" json:"batchSize,omitempty" valid:"numeric,range(1|10000)"`
	Enabled   bool   `yaml:"enabled,omitempty" json:"enabled,omitempty" valid:"-"`
}

// Fetch fetch influxdb config
func (influx *Influxdb) Fetch() (err error) {
	err = influx.cfg.fetchConfig(influx, InfluxdbCategory)
	if err != nil {
		return
	}
	return
}

// Save save influxdb config
func (influx *Influxdb) Save() (err error) {
	return influx.cfg.saveConfig(influx, InfluxdbCategory)
}

// Delete delete influxdb config
func (influx *Influxdb) Delete() (err error) {
	return influx.cfg.deleteConfig(InfluxdbCategory)
}
