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

package upstream

import (
	"net"
	"net/http"
	"testing"

	"github.com/vicanso/pike/config"

	"github.com/stretchr/testify/assert"
)

func TestUpstreams(t *testing.T) {
	assert := assert.New(t)
	l, err := net.Listen("tcp", "0.0.0.0:0")
	assert.Nil(err)
	defer func() {
		_ = l.Close()
	}()
	go func() {
		server := http.Server{}
		_ = server.Serve(l)
	}()
	addr := "http://" + l.Addr().String()

	name := "test"
	upstreamsConfig := config.Upstreams{
		&config.Upstream{
			Name: name,
			Servers: []config.UpstreamServer{
				config.UpstreamServer{
					Addr: addr,
				},
				config.UpstreamServer{
					Backup: true,
					Addr:   "http://127.0.0.1:80",
				},
			},
		},
	}
	upstreams := NewUpstreams(upstreamsConfig)
	us := upstreams.Get(name)
	assert.NotNil(us)
	for i := 0; i < 10; i++ {
		httpUpstream, _ := us.Next()
		assert.Equal(addr, httpUpstream.URL.String())
	}
	data := upstreams.Status()
	assert.Equal(1, len(data))

	upstreams.Destroy()
}
