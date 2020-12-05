// MIT License

// Copyright (c) 2020 Tree Xie

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package config

import (
	"math/rand"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	assert := assert.New(t)

	// location配置的upstream不存在
	c := &PikeConfig{
		Locations: []LocationConfig{
			{
				Name:     "location-test",
				Upstream: "upstream-test",
			},
		},
	}
	err := c.Validate()
	assert.Equal(ErrUpstreamNotFound, err)

	c = &PikeConfig{
		Servers: []ServerConfig{
			{
				Addr: ":3015",
				Locations: []string{
					"location-test",
				},
				Cache: "cache-test",
			},
		},
	}
	err = c.Validate()
	assert.Equal(ErrLocationNotFound, err)

	c = &PikeConfig{
		Upstreams: []UpstreamConfig{
			{
				Name: "upstream-test",
				Servers: []UpstreamServerConfig{
					{
						Addr: "http://127.0.0.1:3015",
					},
				},
			},
		},
		Locations: []LocationConfig{
			{
				Name:     "location-test",
				Upstream: "upstream-test",
			},
		},
		Servers: []ServerConfig{
			{
				Addr: ":3015",
				Locations: []string{
					"location-test",
				},
				Cache: "cache-test",
			},
		},
	}
	err = c.Validate()
	assert.Equal(ErrCacheNotFound, err)

	c = &PikeConfig{
		Caches: []CacheConfig{
			{
				Name:       "cache-test",
				Size:       100,
				HitForPass: "1m",
			},
		},
		Upstreams: []UpstreamConfig{
			{
				Name: "upstream-test",
				Servers: []UpstreamServerConfig{
					{
						Addr: "http://127.0.0.1:3015",
					},
				},
			},
		},
		Locations: []LocationConfig{
			{
				Name:     "location-test",
				Upstream: "upstream-test",
			},
		},
		Servers: []ServerConfig{
			{
				Addr: ":3015",
				Locations: []string{
					"location-test",
				},
				Cache:    "cache-test",
				Compress: "compress-test",
			},
		},
	}
	err = c.Validate()
	assert.Equal(ErrCompressNotFound, err)

	c = &PikeConfig{
		Caches: []CacheConfig{
			{
				Name:       "cache-test",
				Size:       100,
				HitForPass: "1m",
			},
		},
		Compresses: []CompressConfig{
			{
				Name: "compress-test",
			},
		},
		Upstreams: []UpstreamConfig{
			{
				Name:        "upstream-test",
				HealthCheck: "/ping",
				Policy:      "first",
				Servers: []UpstreamServerConfig{
					{
						Addr: "http://127.0.0.1:3015",
					},
				},
			},
		},
		Locations: []LocationConfig{
			{
				Name:     "location-test",
				Upstream: "upstream-test",
				Prefixes: []string{
					"/api",
				},
				Rewrites: []string{
					"/api/:/$1",
				},
				QueryStrings: []string{
					"id:1",
				},
				Hosts: []string{
					"test.com",
				},
				RespHeaders: []string{
					"X-Resp-Id:1",
				},
				ReqHeaders: []string{
					"X-Req-Id:2",
				},
			},
		},
		Servers: []ServerConfig{
			{
				Addr: ":3015",
				Locations: []string{
					"location-test",
				},
				Cache:                     "cache-test",
				Compress:                  "compress-test",
				CompressMinLength:         "1kb",
				CompressContentTypeFilter: "text|json",
			},
		},
	}
	err = c.Validate()
	assert.Nil(err)
}

func TestInitDefaultClient(t *testing.T) {
	assert := assert.New(t)

	file := strconv.Itoa(int(rand.Int31()))

	err := InitDefaultClient("etcd://127.0.0.1:2379/" + file)
	assert.Nil(err)

	err = InitDefaultClient(os.TempDir() + "/" + file)
	assert.Nil(err)
	err = Close()
	assert.Nil(err)
}

func TestReadWriteConfig(t *testing.T) {
	assert := assert.New(t)

	file := strconv.Itoa(int(rand.Int31()))
	err := InitDefaultClient(os.TempDir() + "/" + file)
	assert.Nil(err)

	c := &PikeConfig{
		Compresses: []CompressConfig{
			{
				Name: "compress-test",
			},
		},
	}
	err = Write(c)
	assert.Nil(err)

	currentConfig, err := Read()
	assert.Nil(err)
	assert.Equal(c.Compresses[0].Name, currentConfig.Compresses[0].Name)
}
