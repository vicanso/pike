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

// Upstream server

package upstream

import (
	"github.com/vicanso/pike/config"

	us "github.com/vicanso/upstream"
)

type (
	// Upstreams upstream servers
	Upstreams struct {
		httpUps map[string]*us.HTTP
	}
	// UpStream upstream status
	UpStream struct {
		Name   string `json:"name,omitempty"`
		URL    string `json:"url,omitempty"`
		Status string `json:"status,omitempty"`
	}
	// OnStatus on status listener
	OnStatus func(UpStream)
)

// NewUpstreams create a new upstreams
func NewUpstreams(upstreamsConfig config.Upstreams) *Upstreams {
	upstreams := make(map[string]*us.HTTP)
	for _, stream := range upstreamsConfig {
		uh := &us.HTTP{
			Policy: stream.Policy,
		}
		if stream.HealthCheck != "" {
			uh.Ping = stream.HealthCheck
		}
		for _, server := range stream.Servers {
			addr := server.Addr
			// 如果添加失败，直接忽略
			if server.Backup {
				_ = uh.AddBackup(addr)
			} else {
				_ = uh.Add(addr)
			}
		}
		// 先执行一次health check，获取当前可用服务列表
		uh.DoHealthCheck()
		// 后续需要定时检测upstream是否可用
		go uh.StartHealthCheck()
		upstreams[stream.Name] = uh
	}

	return &Upstreams{
		httpUps: upstreams,
	}
}

// Get get http upstream
func (upstreams *Upstreams) Get(name string) *us.HTTP {
	return upstreams.httpUps[name]
}

// Destroy destroy all upstreams
func (upstreams *Upstreams) Destroy() {
	for _, item := range upstreams.httpUps {
		item.StopHealthCheck()
	}
}

// Status get upstreams status
func (upstreams *Upstreams) Status() map[string][]UpStream {
	data := make(map[string][]UpStream)
	for name, item := range upstreams.httpUps {
		ups := make([]UpStream, 0)
		for _, up := range item.GetUpstreamList() {
			ups = append(ups, UpStream{
				URL:    up.URL.String(),
				Status: up.StatusDesc(),
			})
		}
		data[name] = ups
	}
	return data
}

// OnStatus add event listener to watch upstream's status
func (upstreams *Upstreams) OnStatus(onStats OnStatus) {
	for name, item := range upstreams.httpUps {
		item.OnStatus(func(status int32, upstream *us.HTTPUpstream) {
			info := UpStream{
				Name:   name,
				URL:    upstream.URL.String(),
				Status: us.ConvertStatusToString(status),
			}
			onStats(info)
		})
	}
}
