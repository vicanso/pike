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

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpstreamConfig(t *testing.T) {
	cfg := NewTestConfig()
	assert := assert.New(t)
	us := &Upstream{
		cfg:         cfg,
		Policy:      "first",
		HealthCheck: "/ping",
		Name:        "testupstream",
		Description: "upstream description",
	}
	defer func() {
		_ = us.Delete()
	}()

	err := us.Fetch()
	assert.Nil(err)
	assert.Empty(us.Servers)

	upstreamServer := UpstreamServer{
		Addr:   "127.0.0.1:7000",
		Weight: 10,
		Backup: true,
	}
	us.Servers = make([]UpstreamServer, 1)
	us.Servers[0] = upstreamServer
	err = us.Save()
	assert.Nil(err)

	nus := &Upstream{
		cfg:  cfg,
		Name: us.Name,
	}
	err = nus.Fetch()
	assert.Nil(err)
	assert.Equal(1, len(nus.Servers))
	assert.Equal(upstreamServer, nus.Servers[0])

	upstreams, err := cfg.GetUpstreams()
	assert.Nil(err)
	assert.Equal(1, len(upstreams))

	nus = upstreams.Get(us.Name)
	assert.Equal(us, nus)
}
