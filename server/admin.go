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

	g := elton.NewGroup(adminPath)

	g.GET("/configs/:name", func(c *elton.Context) (err error) {
		var data interface{}
		name := c.Param("name")
		switch name {
		case "caches":
			data, err = config.GetCaches()
		case "compresses":
			data, err = config.GetCompresses()
		case "locations":
			data, err = config.GetLocations()
		case "servers":
			data, err = config.GetServers()
		case "upstreams":
			data, err = config.GetUpstreams()
		default:
			err = hes.New(name + " is not support")
		}
		if err != nil {
			return
		}
		res := make(map[string]interface{})
		res[name] = data
		c.Body = res
		return
	})

	e.AddGroup(g)
	return e
}
