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
	"context"
	"sync"
	"time"

	influxdb "github.com/influxdata/influxdb-client-go"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/log"
	"go.uber.org/zap"
)

type (
	InfluxSrv struct {
		client *influxdb.Client
		sync.Mutex
		BatchSize int
		Bucket    string
		Org       string
		metrics   []influxdb.Metric
	}
)

func NewInfluxSrv(cfg *config.Influxdb) (*InfluxSrv, error) {
	c, err := influxdb.New(cfg.URI, cfg.Token)
	if err != nil {
		return nil, err
	}
	batchSize := 100
	if cfg.BatchSize > 0 {
		batchSize = cfg.BatchSize
	}
	return &InfluxSrv{
		client:    c,
		BatchSize: batchSize,
		Bucket:    cfg.Bucket,
		Org:       cfg.Org,
	}, nil
}

// Write write metric to influxdb
func (srv *InfluxSrv) Write(measurement string, fields map[string]interface{}, tags map[string]string) {
	metric := influxdb.NewRowMetric(fields, measurement, tags, time.Now())
	srv.Lock()
	defer srv.Unlock()
	size := len(srv.metrics)
	if size == 0 {
		srv.metrics = make([]influxdb.Metric, 0, srv.BatchSize)
	}
	srv.metrics = append(srv.metrics, metric)
	if size+1 >= srv.BatchSize {
		metrics := srv.metrics
		go func() {
			srv.writeMetrics(metrics)
		}()
		srv.metrics = nil
	}
}

// Flush flush metric list
func (srv *InfluxSrv) Flush() {
	srv.Lock()
	defer srv.Unlock()
	metrics := srv.metrics
	if len(metrics) == 0 {
		return
	}
	go func() {
		srv.writeMetrics(metrics)
	}()
	srv.metrics = nil
}

// writeMetrics write metric list to influxdb
func (srv *InfluxSrv) writeMetrics(metrics []influxdb.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := srv.client.Write(
		ctx,
		srv.Bucket,
		srv.Org,
		metrics...,
	)
	if err != nil {
		log.Default().Error("influxdb write fail",
			zap.Error(err),
		)
	}
}
