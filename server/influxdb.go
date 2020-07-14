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

package server

import (
	"time"

	influxdb "github.com/influxdata/influxdb-client-go"
	influxdbAPI "github.com/influxdata/influxdb-client-go/api"
	"github.com/vicanso/pike/config"
)

type (
	InfluxSrv struct {
		client influxdb.Client
		writer influxdbAPI.WriteApi
	}
)

func NewInfluxSrv(cfg *config.Influxdb) (*InfluxSrv, error) {
	opts := influxdb.DefaultOptions()
	opts.SetBatchSize(cfg.BatchSize)
	if cfg.FlushInterval != 0 {
		opts.SetFlushInterval(cfg.FlushInterval)
	}

	c := influxdb.NewClientWithOptions(cfg.URI, cfg.Token, opts)
	writer := c.WriteApi(cfg.Org, cfg.Bucket)

	return &InfluxSrv{
		client: c,
		writer: writer,
	}, nil
}

// Write write metric to influxdb
func (srv *InfluxSrv) Write(measurement string, fields map[string]interface{}, tags map[string]string) {
	srv.writer.WritePoint(influxdb.NewPoint(measurement, tags, fields, time.Now()))
}

// Close flush the point to influxdb and close client
func (srv *InfluxSrv) Close() {
	srv.client.Close()
}
