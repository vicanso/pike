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
	"github.com/vicanso/elton"
	bodyparser "github.com/vicanso/elton-body-parser"
	errorhandler "github.com/vicanso/elton-error-handler"
	etag "github.com/vicanso/elton-etag"
	fresh "github.com/vicanso/elton-fresh"
	responder "github.com/vicanso/elton-responder"
	"github.com/vicanso/hes"
	"github.com/vicanso/pike/config"
)

// NewAdmin new an admin elton istance
func NewAdmin(adminPath string, eltonConfig *EltonConfig) *elton.Elton {
	e := elton.New()

	e.Use(fresh.NewDefault())
	e.Use(etag.NewDefault())

	e.Use(responder.NewDefault())
	e.Use(errorhandler.NewDefault())
	e.Use(bodyparser.NewDefault())

	g := elton.NewGroup(adminPath)

	g.GET("/configs/:category", func(c *elton.Context) (err error) {
		var data interface{}
		category := c.Param("category")
		switch category {
		case config.CachesCategory:
			data, err = config.GetCaches()
		case config.CompressCategory:
			data, err = config.GetCompresses()
		case config.LocationsCategory:
			data, err = config.GetLocations()
		case config.ServersCategory:
			data, err = config.GetServers()
		case config.UpstreamsCategory:
			data, err = config.GetUpstreams()
		default:
			err = hes.New(category + " is not support")
		}
		if err != nil {
			return
		}
		res := make(map[string]interface{})
		res[category] = data
		c.Body = res
		return
	})

	// 添加与更新使用相同处理
	g.POST("/configs/:category", func(c *elton.Context) (err error) {
		category := c.Param("category")
		switch category {
		case config.CachesCategory:
			cacheConfig := config.Cache{}
			err = doValidate(&cacheConfig, c.RequestBody)
			if err != nil {
				return
			}
			err = cacheConfig.Save()
			if err != nil {
				return
			}
		case config.CompressCategory:
			compressConfig := config.Compress{}
			err = doValidate(&compressConfig, c.RequestBody)
			if err != nil {
				return
			}
			err = compressConfig.Save()
			if err != nil {
				return
			}
		case config.LocationsCategory:
			locationConfig := config.Location{}
			err = doValidate(&locationConfig, c.RequestBody)
			if err != nil {
				return
			}
			err = locationConfig.Save()
			if err != nil {
				return
			}
		case config.ServersCategory:
			serverConfig := config.Server{}
			err = doValidate(&serverConfig, c.RequestBody)
			if err != nil {
				return
			}
			err = serverConfig.Save()
			if err != nil {
				return
			}
		case config.UpstreamsCategory:
			upstreamConfig := config.Upstream{}
			err = doValidate(&upstreamConfig, c.RequestBody)
			if err != nil {
				return
			}
			err = upstreamConfig.Save()
			if err != nil {
				return
			}
		default:
			err = hes.New(category + " is not support")
		}
		if err != nil {
			return
		}
		c.NoContent()
		return
	})

	e.AddGroup(g)
	return e
}
