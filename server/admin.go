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
	"net"
	"strings"

	"github.com/vicanso/elton"
	basicauth "github.com/vicanso/elton-basic-auth"
	bodyparser "github.com/vicanso/elton-body-parser"
	errorhandler "github.com/vicanso/elton-error-handler"
	etag "github.com/vicanso/elton-etag"
	fresh "github.com/vicanso/elton-fresh"
	responder "github.com/vicanso/elton-responder"
	"github.com/vicanso/hes"
	intranetip "github.com/vicanso/intranet-ip"
	"github.com/vicanso/pike/config"
)

// NewAdmin new an admin elton istance
func NewAdmin(adminPath string, eltonConfig *EltonConfig) *elton.Elton {
	e := elton.New()

	adminConfig := eltonConfig.adminConfig

	if adminConfig != nil {
		// 不允许外网访问
		if !adminConfig.EnabledInternetAccess {
			e.Use(func(c *elton.Context) (err error) {
				// 会获取客户的访问IP（获取到非内网IP为止，如果都没有，则remote addr)
				ip := c.ClientIP()
				if !intranetip.Is(net.ParseIP(ip)) {
					err = hes.New("Not allow to access")
					return
				}
				return c.Next()
			})
		}
		user := adminConfig.User
		password := adminConfig.Password
		// 如果配置了认证
		if user != "" && password != "" {
			e.Use(basicauth.New(basicauth.Config{
				Validate: func(account, pwd string, c *elton.Context) (bool, error) {
					if account == user && pwd == password {
						return true, nil
					}
					return false, nil
				},
			}))
		}
	}

	e.Use(fresh.NewDefault())
	e.Use(etag.NewDefault())

	e.Use(responder.NewDefault())
	e.Use(errorhandler.NewDefault())
	e.Use(bodyparser.NewDefault())

	g := elton.NewGroup(adminPath)

	g.GET("/configs/:category", func(c *elton.Context) (err error) {
		var data interface{}
		res := make(map[string]interface{})
		arr := strings.Split(c.Param("category"), ",")
		for _, category := range arr {
			switch category {
			case config.CachesCategory:
				data, err = config.GetCaches()
			case config.CompressesCategory:
				data, err = config.GetCompresses()
			case config.LocationsCategory:
				data, err = config.GetLocations()
			case config.ServersCategory:
				data, err = config.GetServers()
			case config.UpstreamsCategory:
				data, err = config.GetUpstreams()
			case config.AdminCategory:
				data, err = config.GetAdmin()
			default:
				err = hes.New(category + " is not support")
			}
			if err != nil {
				return
			}
			res[category] = data
		}
		c.Body = res
		return
	})

	// 添加与更新使用相同处理
	g.POST("/configs/:category", func(c *elton.Context) (err error) {
		category := c.Param("category")
		var iconfig config.IConfig
		switch category {
		case config.CachesCategory:
			iconfig = new(config.Cache)
		case config.CompressesCategory:
			iconfig = new(config.Compress)
		case config.LocationsCategory:
			iconfig = new(config.Location)
		case config.ServersCategory:
			iconfig = new(config.Server)
		case config.UpstreamsCategory:
			iconfig = new(config.Upstream)
		case config.AdminCategory:
			iconfig = new(config.Admin)
		default:
			err = hes.New(category + " is not support")
		}

		err = doValidate(iconfig, c.RequestBody)
		if err != nil {
			return
		}
		err = iconfig.Save()
		if err != nil {
			return
		}

		if err != nil {
			return
		}
		c.NoContent()
		return
	})

	g.DELETE("/configs/:category/:name", func(c *elton.Context) (err error) {
		serverConfigs, err := config.GetServers()
		if err != nil {
			return
		}
		locations, err := config.GetLocations()
		if err != nil {
			return
		}

		category := c.Param("category")
		name := c.Param("name")
		var iconfig config.IConfig
		shouldBeCheckedByServer := false
		shouldBeCheckedByLocation := false
		switch category {
		case config.CachesCategory:
			shouldBeCheckedByServer = true
			iconfig = &config.Cache{
				Name: name,
			}
		case config.CompressesCategory:
			shouldBeCheckedByServer = true
			iconfig = &config.Compress{
				Name: name,
			}
		case config.LocationsCategory:
			shouldBeCheckedByServer = true
			iconfig = &config.Location{
				Name: name,
			}
		case config.ServersCategory:
			iconfig = &config.Server{
				Name: name,
			}
		case config.UpstreamsCategory:
			shouldBeCheckedByLocation = true
			iconfig = &config.Upstream{
				Name: name,
			}
		default:
			err = hes.New(category + " is not support")
			return
		}
		// 判断是否在现有server配置中有使用
		if shouldBeCheckedByServer && serverConfigs.Exists(category, name) {
			err = hes.New(name + " of " + category + " is used by server, it can't be delelted")
			return
		}
		// 判断是否有location在使用该upstream
		if shouldBeCheckedByLocation && locations.ExistsUpstream(name) {
			err = hes.New(name + " of " + category + " is used by location, it can't be delelted")
			return
		}

		err = iconfig.Delete()
		if err != nil {
			return
		}
		c.NoContent()
		return
	})

	e.AddGroup(g)
	return e
}
