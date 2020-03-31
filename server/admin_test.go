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

	t.Run("cert config", func(t *testing.T) {
		category := config.CertsCategory
		name := "me.dev"
		c := newContext(category, []byte(`{
			"name": "`+name+`",
			"key": "LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JSUV2Z0lCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQktnd2dnU2tBZ0VBQW9JQkFRREVuMlg1VTdnd2VLdEcKbDhFSDJ5VVNuVGFPSEVLcERkaDNPcUp2TnlDcjk5dVZ3WnhMaUx3ZEpOR1BzMmY1ZXZkeW1zZU1wdjVoRk1GTApxQlZwQjVSdWFOK3Q1UVNNWDZmcnpINURtR0JxYlFCcUpHTjV5OHJpUklUYjVCcGtuYWQyMWFkbEh1WjU2bFp2ClRKWjJUa1JZTDRUMmZxZUR1U2pPL05iWFp2SUVJS2M3R1pZT1BhYlRMeGhUOWVlNGVtUUFKMWpwMjZRRVRxYmsKd0JSWmFuWmxtWVlhS0I0aXI1Q2swb29Jc21jM291NGI0UFJibjRtdVpPd25QSWl4dk5VUFpYenVyakJHdjB4MQpXVmFVSUlabWt6T1VBU0JzM0ZScFRsMk93N1lhZFVTODF0OEVWUzRGQ1k0akFmM0ExZWFlNDlPVlV1SjVaTndtCmhGaEp0V3RUQWdNQkFBRUNnZ0VBUndGVlB4UFh1VkZxY09UT3BicWpDYlRTaGNGNDVUb0Z5UkRZcGhjZmFscm8KNW96emwyZDZuMyt6V2hTczRMQmllZldoU0k3cDREOHhpdFBaWDRPSU85TU5xK3UvbDczWGsxVFc2Q3czN1ZjTgp4a2I3MFhraC9GSklOR3lNaDNkVGladWdodWtBekZndS9LU0kxWkp3SmZTTExNVVNVNFJqVTFTRmRXWk0wZVloCnV6MndXS2xPR0hrMGZZZFFwQUx6ZldQcHJJZWwxbXF0U2hxRXErRzFxWmJhendZZCtLZmluZVJyQUM0ZEZxWWEKZUMxd3ZlRENjdzc2MG91L0E1THA2d2o3UW16enkyclFQMEsxOUpHUjd6dmVNbjBZRUhWNnRTK1QvdlIyZ0h4aQpkeUlkeUZrOGtyNG5mYTRvbHNnR0VST3ltWVYrK1VCRGovUW93R0Ura1FLQmdRREh3aWhVaW01bEZHYVpNL0prCmd5QmJYQS80UnpXNUJHWjdqRG0yRFhEc3NPYkpkTDBNSUEvNE5rdDdMTDFTWWh3aGFlQ1F5a3Zhak5HNFBzMU0KR3VjRTVkU2RzRm9rS0FOeW0yM2R4VGdhQmFOZ1ZRYTllQytMT0tRZHBROTBGaWhZaEJlUE9jWmxIdU50OUpWNwp3eGlRSVNJUyszMTQxdUtZZEZ5T3ZEWGI2d0tCZ1FENyt6bkJGWFg5SXhLbFVtWjJXc1Z3UzFCY0F3V1pWUjJlCk9OZytKU04xTHVrNk5WTTRMMXp3anViWGJYN1hNU2MxOEhLVkRRV1ZhVEFrN28rOEpjMEwxc1BJaGgzbGRQMmsKWDFoSk82K3ZIc0JTL09yQzNLVWkrOG9vTiswSHNZVm9pSEJXNmhsa05NU2NOVDRBNyt1RUpjd3U1WjFnWTZPcgpKK0NLWTAxY09RS0JnSHN1Q2lxZnRwV1VML1JYS1NpOEIwN3ZCVllIcTJRdEIzazJManhLSzVGNVFNZUh1aS9vCjhaQVJBeGl3clFwSlA2bUhIWmlMZHAwTmF5R2ZjSDkyczNDOHZSQ0VPQUhGdnVLRVlBcDZYQzhIdlFoaFJpZSsKSGl0T3dUMGFsTjN6NytzdGdVMnJ4ZUNEWEtGb1NtbW9FOVNFNmZza28rbkpNSy9zU1Vzbldsc0RBb0dCQUtZYQpNRE1RYzR1ZlVBNDhxQ0JDcTczZlY2U2Z0VlFqSUhnSkhycXdmcFFqalVoNm1GWUVHcTdVZEdUejM5WDRwOUZOCnBDcU92K3lDdjJMSkEyVFNRajBZb0V5UjVDazZtbXg5RVZTTkRMMVNkeEw5ZDc5bDlWRi9TdjZDQnpTNEY2b1YKcm9BTXB4cEFFbzZxSmlvMS9UbEtOVE9BMXVJUUxIYUp2ZUZibmtZNUFvR0JBS2ZUalhiRzRwTkdBYlFBalhBTgpRMjNFL0pMWmdZWlpEbmFNazhGRHBORnRwNHdFMEwvQXl0dEtmMUh1UXRTanhLMTBETjZOZmlqeUJGaVBuUGl3Ckk5bTA0NmVnWHRuR0FqaGdaUnd3eDM1YURyZDErank0UElIMVNYNXJXdWdxK0dyU01oZk81cjZJUW80V1dwa3UKeStqNDl1Z013WXJ4bU52MVc0Wm5uWHZICi0tLS0tRU5EIFBSSVZBVEUgS0VZLS0tLS0K",
			"cert": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUVnekNDQXV1Z0F3SUJBZ0lSQUlqZ1dQMWFlRk0xMFFsUU9FTGJTb1F3RFFZSktvWklodmNOQVFFTEJRQXcKZ1kweEhqQWNCZ05WQkFvVEZXMXJZMlZ5ZENCa1pYWmxiRzl3YldWdWRDQkRRVEV4TUM4R0ExVUVDd3dvZUdsbApjMmgxZW1odmRVQjRhV1Z6YUhWNmFHOTFjeTFOWVdOQ2IyOXJMVkJ5Ynk1c2IyTmhiREU0TURZR0ExVUVBd3d2CmJXdGpaWEowSUhocFpYTm9kWHBvYjNWQWVHbGxjMmgxZW1odmRYTXRUV0ZqUW05dmF5MVFjbTh1Ykc5allXd3cKSGhjTk1Ua3dOakF4TURBd01EQXdXaGNOTWpreE1qTXhNRGN5TkRNMldqQmNNU2N3SlFZRFZRUUtFeDV0YTJObApjblFnWkdWMlpXeHZjRzFsYm5RZ1kyVnlkR2xtYVdOaGRHVXhNVEF2QmdOVkJBc01LSGhwWlhOb2RYcG9iM1ZBCmVHbGxjMmgxZW1odmRYTXRUV0ZqUW05dmF5MVFjbTh1Ykc5allXd3dnZ0VpTUEwR0NTcUdTSWIzRFFFQkFRVUEKQTRJQkR3QXdnZ0VLQW9JQkFRREVuMlg1VTdnd2VLdEdsOEVIMnlVU25UYU9IRUtwRGRoM09xSnZOeUNyOTl1Vgp3WnhMaUx3ZEpOR1BzMmY1ZXZkeW1zZU1wdjVoRk1GTHFCVnBCNVJ1YU4rdDVRU01YNmZyekg1RG1HQnFiUUJxCkpHTjV5OHJpUklUYjVCcGtuYWQyMWFkbEh1WjU2bFp2VEpaMlRrUllMNFQyZnFlRHVTak8vTmJYWnZJRUlLYzcKR1pZT1BhYlRMeGhUOWVlNGVtUUFKMWpwMjZRRVRxYmt3QlJaYW5abG1ZWWFLQjRpcjVDazBvb0lzbWMzb3U0Ygo0UFJibjRtdVpPd25QSWl4dk5VUFpYenVyakJHdjB4MVdWYVVJSVpta3pPVUFTQnMzRlJwVGwyT3c3WWFkVVM4CjF0OEVWUzRGQ1k0akFmM0ExZWFlNDlPVlV1SjVaTndtaEZoSnRXdFRBZ01CQUFHamdZMHdnWW93RGdZRFZSMFAKQVFIL0JBUURBZ1dnTUJNR0ExVWRKUVFNTUFvR0NDc0dBUVVGQndNQk1Bd0dBMVVkRXdFQi93UUNNQUF3SHdZRApWUjBqQkJnd0ZvQVVJWVlsU1NvaTIxVEpjRm5wbE5LeXkxNTRuMDR3TkFZRFZSMFJCQzB3SzRJR2JXVXVaR1YyCmdnbHNiMk5oYkdodmMzU0hCSDhBQUFHSEVBQUFBQUFBQUFBQUFBQUFBQUFBQUFFd0RRWUpLb1pJaHZjTkFRRUwKQlFBRGdnR0JBRUlDa01DS2lsaU82YmFBR3dYc1BKZ294RVNRS3RURDFrR0FRNE5yWXNJUlpISjhuTGk5QXA1RQpsQzVuZlp6blNwa1NpNXhjQ2lrYUFJeWVHZmtON2hVUVZCUmFKUTBBbTN5WFpzWFhjL3greUFjUDNjNVhVMml4CkRiZmJTK2JHeEcwb3NXRDFiYzBRdU1Ibk83SFhIUmpEb0ZOU0VNSjlBdVdST1ZRTEdqNkhWVWFzUERPdmhid3EKWENwbWc4OXphVXlralloZVBxMjVWMlZzTjQ1UzQ0Vm1Qb1VUbzBxVUNnRC9SS0ZLSTVYdjJlWUk5U0FCTHdQTApoSnp2aEk1K2czU0NOeTlxYllsNUszYUFIdjIvRWtpSGNpYUpVc3hTcUsxcmU0NytOZVNRendsTmt1TnIxN3JHCmNRTDdyVnZIeUtpSksyNTJSa2dmcFpLaFdObWVIcHEwcVRaVndLaXp6RnhkRFZoYTdYU1RTWlQ2VW5mZmdpOXUKcjI2SXIyVndsSHluekhiR0UwT3ZMQVdmUldWL1Nhakt3SDZMNUd4S2ZvTzdtTE9OZVZyVndhcVlhUVo2MEdMMAp4eHlEZm01eGJ4anRvZ3NGT24zUUNBWHhaejE0YmhuTDRrRzFkNDFFWFI0WXpMZDVZYWhWYVB6WDZkcS9Oa2JGCjRVUnhyOEhPV0E9PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="
		}`))

		err := createOrUpdateConfig(c)
		assert.Nil(err)

		c = newContext(category, nil)
		err = getConfigs(c)
		assert.Nil(err)
		certs := c.Body.(map[string]interface{})[category].(config.Certs)
		assert.Equal(name, certs.Get(name).Name)

		c = elton.NewContext(nil, nil)
		c.Params = map[string]string{
			"category": category,
			"name":     name,
		}
		err = deleteConfig(c)
		assert.Nil(err)

		c = newContext(category, nil)
		err = getConfigs(c)
		assert.Nil(err)
		certs = c.Body.(map[string]interface{})[category].(config.Certs)
		assert.Empty(certs)
	})

	t.Run("influxdb config", func(t *testing.T) {
		category := config.InfluxdbCategory
		c := newContext(category, []byte(`{
			"uri": "http://192.168.1.8:9999",
			"bucket": "red",
			"org": "bigTree",
			"token": "EEIjBRUEBSxncDQ1ZN3lBQ_dVBW4gEntJrUhwLnqqoK9CSyFdC1mUM_RGXHJWO7eAiBkMhXFf0XiEFEFFNYDoA==",
			"batchSize": 100,
			"enabled": true
		}`))

		err := createOrUpdateConfig(c)
		assert.Nil(err)

		c = newContext(category, nil)
		err = getConfigs(c)
		assert.Nil(err)
		influxdb := c.Body.(map[string]interface{})[category].(*config.Influxdb)
		assert.NotNil(influxdb)

		c = elton.NewContext(nil, nil)
		c.Params = map[string]string{
			"category": category,
		}
		err = deleteConfig(c)
		assert.Nil(err)
	})

	t.Run("alarm config", func(t *testing.T) {
		name := "upstream"
		category := config.AlarmsCategory
		c := newContext(category, []byte(`{
			"name": "`+name+`",
			"uri": "http://127.0.0.1/send-alarm",
			"template": "{}"
		}`))

		err := createOrUpdateConfig(c)
		assert.Nil(err)

		c = newContext(category, nil)
		err = getConfigs(c)
		assert.Nil(err)
		alarms := c.Body.(map[string]interface{})[category].(config.Alarms)
		assert.Equal(name, alarms.Get(name).Name)

		c = elton.NewContext(nil, nil)
		c.Params = map[string]string{
			"category": category,
			"name":     name,
		}
		err = deleteConfig(c)
		assert.Nil(err)

		c = newContext(category, nil)
		err = getConfigs(c)
		assert.Nil(err)
		alarms = c.Body.(map[string]interface{})[category].(config.Alarms)
		assert.Empty(alarms)
	})
}

func TestNewAdminServer(t *testing.T) {
	// 仅简单测试初始化成功
	assert := assert.New(t)
	cfg := config.NewTestConfig()
	_, e := NewAdmin(&ServerOptions{
		cfg: cfg,
	})
	assert.NotNil(e)
}
