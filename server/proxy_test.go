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

package server

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http/httptest"
	"testing"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/upstream"

	"github.com/stretchr/testify/assert"
	"github.com/vicanso/elton"
	"github.com/vicanso/pike/util"
)

func newTestProxyServer() (net.Listener, config.Locations, *upstream.Upstreams, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:")

	e := elton.New()

	e.GET("/ping", func(c *elton.Context) error {
		c.BodyBuffer = bytes.NewBufferString("pong")
		return nil
	})
	e.GET("/", func(c *elton.Context) error {
		c.BodyBuffer = bytes.NewBufferString("hello world!")
		return nil
	})
	e.GET("/check", func(c *elton.Context) error {
		m := map[string]string{
			"Host":                      c.Request.Host,
			elton.HeaderAcceptEncoding:  c.GetRequestHeader(elton.HeaderAcceptEncoding),
			elton.HeaderIfModifiedSince: c.GetRequestHeader(elton.HeaderIfModifiedSince),
			elton.HeaderIfNoneMatch:     c.GetRequestHeader(elton.HeaderIfNoneMatch),
			"X-Request-ID":              c.GetRequestHeader("X-Request-ID"),
			"url":                       c.Request.URL.RequestURI(),
		}
		buf, _ := json.Marshal(m)
		c.BodyBuffer = bytes.NewBuffer(buf)
		return nil
	})

	go e.Serve(ln) // nolint

	upstreamName := "test-uptream"
	upstreamConfig := config.Upstream{
		HealthCheck: "/ping",
		Name:        upstreamName,
		Policy:      "leastconn",
		Servers: []config.UpstreamServer{
			config.UpstreamServer{
				Addr: "http://" + ln.Addr().String(),
			},
		},
	}
	upstreamsConfig := []*config.Upstream{
		&upstreamConfig,
	}
	upstreams := upstream.NewUpstreams(upstreamsConfig)
	defer upstreams.Destroy()

	locationName := "test"
	locations := []*config.Location{
		&config.Location{
			Name:     locationName,
			Upstream: upstreamName,
			Hosts: []string{
				"aslant.site",
			},
			Prefixs: []string{
				"/api",
			},
			Rewrites: []string{
				"/api/*:/$1",
			},
			ReqHeader: util.ConvertToHTTPHeader([]string{
				"X-Request-Id:123",
				"Host:tiny.aslant.site",
			}),
			ResHeader: util.ConvertToHTTPHeader([]string{
				"X-Response-Id:456",
			}),
		},
	}
	return ln, locations, upstreams, err
}

func TestNewProxyHandlers(t *testing.T) {
	assert := assert.New(t)
	ln, locations, upstreams, err := newTestProxyServer()
	assert.Nil(err)
	defer ln.Close()

	defer upstreams.Destroy()

	fns := newProxyHandlers(locations, upstreams)

	req := httptest.NewRequest("GET", "/", nil)
	resp := httptest.NewRecorder()
	c := elton.NewContext(resp, req)
	c.Next = func() error {
		return nil
	}
	err = fns[locations[0].Name](c)
	assert.Nil(err)
	assert.Equal(200, c.StatusCode)
	assert.Equal([]byte("hello world!"), c.BodyBuffer.Bytes())
}

func TestCreateProxyMiddleware(t *testing.T) {
	assert := assert.New(t)
	ln, locations, upstreams, err := newTestProxyServer()
	assert.Nil(err)
	defer ln.Close()

	defer upstreams.Destroy()
	fn := createProxyMiddleware(locations, upstreams)

	t.Run("fetch request", func(t *testing.T) {
		// fetch request
		req := httptest.NewRequest("GET", "/api/check", nil)
		req.Host = "aslant.site"
		req.Header.Set(elton.HeaderAcceptEncoding, "br, gzip")
		req.Header.Set(elton.HeaderIfModifiedSince, "date time")
		req.Header.Set(elton.HeaderIfNoneMatch, "etag")

		resp := httptest.NewRecorder()
		c := elton.NewContext(resp, req)
		c.Set(statusKey, cache.StatusFetching)
		c.Next = func() error {
			return nil
		}
		err = fn(c)
		assert.Nil(err)
		assert.Equal("456", c.GetHeader("X-Response-Id"))
		m := make(map[string]string)
		err = json.Unmarshal(c.BodyBuffer.Bytes(), &m)
		assert.Nil(err)
		assert.Equal("gzip", m[elton.HeaderAcceptEncoding])
		assert.Empty(m[elton.HeaderIfModifiedSince])
		assert.Empty(m[elton.HeaderIfNoneMatch])
		assert.Equal("/check", m["url"])
		assert.Equal("tiny.aslant.site", m["Host"])
	})

	t.Run("pass requset", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/check", nil)
		req.Host = "aslant.site"
		req.Header.Set(elton.HeaderAcceptEncoding, "br, gzip")
		req.Header.Set(elton.HeaderIfModifiedSince, "date time")
		req.Header.Set(elton.HeaderIfNoneMatch, "etag")
		// pass request
		resp := httptest.NewRecorder()
		c := elton.NewContext(resp, req)
		c.Set(statusKey, cache.StatusPassed)
		c.Next = func() error {
			return nil
		}
		err = fn(c)
		assert.Nil(err)
		assert.Equal("456", c.GetHeader("X-Response-Id"))
		m := make(map[string]string)
		err = json.Unmarshal(c.BodyBuffer.Bytes(), &m)

		assert.Nil(err)
		assert.Equal("br, gzip", m[elton.HeaderAcceptEncoding])
		assert.NotEmpty(m[elton.HeaderIfModifiedSince])
		assert.NotEmpty(m[elton.HeaderIfNoneMatch])
		assert.Equal("/check", m["url"])
		assert.Equal("tiny.aslant.site", m["Host"])
	})
}
