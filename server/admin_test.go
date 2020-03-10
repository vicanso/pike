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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vicanso/elton"
	"github.com/vicanso/hes"
	"github.com/vicanso/pike/config"
)

func TestNewAdminValidateMiddlewares(t *testing.T) {
	assert := assert.New(t)
	adminConfig := &config.Admin{
		EnabledInternetAccess: false,
		User:                  "tree.xie",
		Password:              "password",
	}
	handlers := newAdminValidateMiddlewares(adminConfig)
	assert.Equal(2, len(handlers))

	t.Run("intranet access", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(elton.HeaderXRealIP, "1.1.1.1")
		c := elton.NewContext(nil, req)
		c.Next = func() error {
			return nil
		}
		err := handlers[0](c)
		assert.NotNil(err)
		he, ok := err.(*hes.Error)
		assert.True(ok)
		assert.Equal(403, he.StatusCode)

		req.Header.Del(elton.HeaderXRealIP)
		req.RemoteAddr = "192.168.1.1:3000"
		c = elton.NewContext(nil, req)
		done := false
		c.Next = func() error {
			done = true
			return nil
		}
		err = handlers[0](c)
		assert.Nil(err)
		assert.True(done)
	})

	t.Run("basic auth", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		resp := httptest.NewRecorder()
		c := elton.NewContext(resp, req)
		c.Next = func() error {
			return nil
		}
		err := handlers[1](c)
		assert.NotNil(err)
		he, ok := err.(*hes.Error)
		assert.True(ok)
		assert.Equal(http.StatusUnauthorized, he.StatusCode)

		// 错误的账号或密码
		req.Header.Set("Authorization", "basic dHJlZS54aWU6cGFzcw==")
		err = handlers[1](c)
		assert.NotNil(err)
		he, ok = err.(*hes.Error)
		assert.True(ok)
		assert.Equal(http.StatusUnauthorized, he.StatusCode)

		// 正确的账号密码
		req.Header.Set("Authorization", "basic dHJlZS54aWU6cGFzc3dvcmQ=")
		done := false
		c.Next = func() error {
			done = true
			return nil
		}
		err = handlers[1](c)
		assert.Nil(err)
		assert.True(done)
	})
}

func TestConfigHandler(t *testing.T) {
	cfg := config.NewTestConfig()
	createOrUpdateConfig := newCreateOrUpdateConfigHandler(cfg)
	getConfigs := newGetConfigHandler(cfg)
	deleteConfig := newDeleteConfigHandler(cfg)
	assert := assert.New(t)
	newContext := func(category string, requestBody []byte) *elton.Context {
		c := elton.NewContext(nil, nil)
		c.RequestBody = requestBody
		c.Params = map[string]string{
			"category": category,
		}
		return c
	}
	t.Run("cache config", func(t *testing.T) {
		category := config.CachesCategory
		c := newContext(category, []byte(`{
			"name": "testCache",
			"zone": 1000,
			"size": 10,
			"hitForPass": 300
		}`))
		err := createOrUpdateConfig(c)
		assert.Nil(err)

		c = newContext(category, nil)
		err = getConfigs(c)
		assert.Nil(err)
		caches := c.Body.(map[string]interface{})[category].(config.Caches)
		assert.NotEmpty(caches)

		c = elton.NewContext(nil, nil)
		c.Params = map[string]string{
			"category": category,
			"name":     "testCache",
		}
		err = deleteConfig(c)
		assert.Nil(err)

		c = newContext(category, nil)
		err = getConfigs(c)
		assert.Nil(err)
		caches = c.Body.(map[string]interface{})[category].(config.Caches)
		assert.Empty(caches)
	})

	t.Run("compress config", func(t *testing.T) {
		category := config.CompressesCategory
		c := newContext(category, []byte(`{
			"name": "testCompress",
			"level": 9,
			"minLength": 1000,
			"filter": "text|json"
		}`))
		err := createOrUpdateConfig(c)
		assert.Nil(err)

		c = newContext(category, nil)
		err = getConfigs(c)
		assert.Nil(err)
		compresses := c.Body.(map[string]interface{})[category].(config.Compresses)
		assert.NotEmpty(compresses)

		c = elton.NewContext(nil, nil)
		c.Params = map[string]string{
			"category": category,
			"name":     "testCompress",
		}
		err = deleteConfig(c)
		assert.Nil(err)

		c = newContext(category, nil)
		err = getConfigs(c)
		assert.Nil(err)
		compresses = c.Body.(map[string]interface{})[category].(config.Compresses)
		assert.Empty(compresses)
	})

	t.Run("locations config", func(t *testing.T) {
		category := config.LocationsCategory
		c := newContext(category, []byte(`{
			"name": "testLocation",
			"upstream": "testUpstream",
			"prefixs": ["/api"],
			"rewrites": ["/api/*:/$1"],
			"hosts": ["aslant.site"],
			"requestHeader": ["X-Request-Id:456"],
			"responseHeader": ["X-Response-Id:123"]
		}`))
		err := createOrUpdateConfig(c)
		assert.Nil(err)

		c = newContext(category, nil)
		err = getConfigs(c)
		assert.Nil(err)
		locations := c.Body.(map[string]interface{})[category].(config.Locations)
		assert.NotEmpty(locations)

		c = elton.NewContext(nil, nil)
		c.Params = map[string]string{
			"category": category,
			"name":     "testLocation",
		}
		err = deleteConfig(c)
		assert.Nil(err)

		c = newContext(category, nil)
		err = getConfigs(c)
		assert.Nil(err)
		locations = c.Body.(map[string]interface{})[category].(config.Locations)
		assert.Empty(locations)
	})

	t.Run("servers config", func(t *testing.T) {
		category := config.ServersCategory
		c := newContext(category, []byte(`{
			"name": "testServer",
			"cache": "testCache",
			"compress": "testCompress",
			"locations": ["testLocation"],
			"eTag": true,
			"addr": "127.0.0.1:3000",
			"concurrency": 100,
			"readTimeout": 100000,
			"writeTimeout": 100000,
			"idleTimeout": 100000,
			"maxHeaderBytes": 1000
		}`))
		err := createOrUpdateConfig(c)
		assert.Nil(err)

		c = newContext(category, nil)
		err = getConfigs(c)
		assert.Nil(err)
		servers := c.Body.(map[string]interface{})[category].(config.Servers)
		assert.NotEmpty(servers)

		c = elton.NewContext(nil, nil)
		c.Params = map[string]string{
			"category": category,
			"name":     "testServer",
		}
		err = deleteConfig(c)
		assert.Nil(err)

		c = newContext(category, nil)
		err = getConfigs(c)
		assert.Nil(err)
		servers = c.Body.(map[string]interface{})[category].(config.Servers)
		assert.Empty(servers)
	})

	t.Run("upstream config", func(t *testing.T) {
		category := config.UpstreamsCategory
		c := newContext(category, []byte(`{
			"healthCheck": "/ping",
			"policy": "first",
			"name": "testUpstream",
			"servers": [
				{
					"addr": "http://127.0.0.1:3000",
					"backup": true
				}
			]
		}`))
		err := createOrUpdateConfig(c)
		assert.Nil(err)

		c = newContext(category, nil)
		err = getConfigs(c)
		assert.Nil(err)
		upstreams := c.Body.(map[string]interface{})[category].(config.Upstreams)
		assert.NotEmpty(upstreams)

		c = elton.NewContext(nil, nil)
		c.Params = map[string]string{
			"category": category,
			"name":     "testUpstream",
		}
		err = deleteConfig(c)
		assert.Nil(err)

		c = newContext(category, nil)
		err = getConfigs(c)
		assert.Nil(err)
		upstreams = c.Body.(map[string]interface{})[category].(config.Upstreams)
		assert.Empty(upstreams)
	})

	t.Run("admin config", func(t *testing.T) {
		category := config.AdminCategory
		c := newContext(category, []byte(`{
			"prefix": "/pike",
			"user": "user",
			"password": "pass",
			"enabledInternetAccess": true
		}`))
		err := createOrUpdateConfig(c)
		assert.Nil(err)

		c = newContext(category, nil)
		err = getConfigs(c)
		assert.Nil(err)
		admin := c.Body.(map[string]interface{})[category].(*config.Admin)
		assert.NotNil(admin)

		err = admin.Delete()
		assert.Nil(err)
	})
}

func TestNewAdminServer(t *testing.T) {
	// 仅简单测试初始化成功
	assert := assert.New(t)
	cfg := config.NewTestConfig()
	_, e := NewAdmin(cfg)
	assert.NotNil(e)
}
