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

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfluxdbConfig(t *testing.T) {
	assert := assert.New(t)
	cfg := NewTestConfig()
	defer func() {
		_ = (&Influxdb{
			cfg: cfg,
		}).Delete()
	}()

	influxdb, err := cfg.GetInfluxdb()
	assert.Nil(err)
	assert.Empty(influxdb.URI)
	assert.Empty(influxdb.Bucket)
	assert.Empty(influxdb.Org)
	assert.Empty(influxdb.Token)
	assert.Empty(influxdb.BatchSize)
	assert.False(influxdb.Enabled)

	uri := "http://127.0.0.1:9999"
	bucket := "bucket"
	org := "org"
	token := "token"
	batchSize := uint(100)
	flushInterval := uint(5000)
	enabled := true
	influxdb.URI = uri
	influxdb.Bucket = bucket
	influxdb.Org = org
	influxdb.Token = token
	influxdb.BatchSize = batchSize
	influxdb.FlushInterval = flushInterval
	influxdb.Enabled = enabled
	err = influxdb.Save()
	assert.Nil(err)

	influxdb = &Influxdb{
		cfg: cfg,
	}
	err = influxdb.Fetch()
	assert.Nil(err)
	assert.Equal(uri, influxdb.URI)
	assert.Equal(bucket, influxdb.Bucket)
	assert.Equal(org, influxdb.Org)
	assert.Equal(token, influxdb.Token)
	assert.Equal(batchSize, influxdb.BatchSize)
	assert.Equal(flushInterval, influxdb.FlushInterval)
	assert.True(influxdb.Enabled)
}
