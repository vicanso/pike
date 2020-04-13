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
	"encoding/base64"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/vicanso/elton"
	compress "github.com/vicanso/elton-compress"
	"github.com/vicanso/elton/middleware"
	"github.com/vicanso/hes"
	intranetip "github.com/vicanso/intranet-ip"
	"github.com/vicanso/pike/application"
	"github.com/vicanso/pike/config"
)

func newAdminValidateMiddlewares(adminConfig *config.Admin) []elton.Handler {
	handlers := make([]elton.Handler, 0)
	// 不允许外网访问
	if !adminConfig.EnabledInternetAccess {
		fn := func(c *elton.Context) (err error) {
			// 会获取客户的访问IP（获取到非内网IP为止，如果都没有，则remote addr)
			ip := c.ClientIP()
			if !intranetip.Is(net.ParseIP(ip)) {
				err = hes.NewWithStatusCode("Not allow to access", http.StatusForbidden)
				return
			}
			return c.Next()
		}
		handlers = append(handlers, fn)
	}
	user := adminConfig.User
	password := adminConfig.Password
	// 如果配置了认证
	if user != "" && password != "" {
		fn := middleware.NewBasicAuth(middleware.BasicAuthConfig{
			Validate: func(account, pwd string, c *elton.Context) (bool, error) {
				if account == user && pwd == password {
					return true, nil
				}
				return false, nil
			},
		})

		handlers = append(handlers, fn)
	}
	return handlers
}
func newGetConfigHandler(cfg *config.Config) elton.Handler {
	return func(c *elton.Context) (err error) {
		var data interface{}
		res := make(map[string]interface{})
		arr := strings.Split(c.Param("category"), ",")
		for _, category := range arr {
			switch category {
			case config.CachesCategory:
				data, err = cfg.GetCaches()
			case config.CompressesCategory:
				data, err = cfg.GetCompresses()
			case config.LocationsCategory:
				data, err = cfg.GetLocations()
			case config.ServersCategory:
				data, err = cfg.GetServers()
			case config.UpstreamsCategory:
				data, err = cfg.GetUpstreams()
			case config.AdminCategory:
				data, err = cfg.GetAdmin()
			case config.CertsCategory:
				data, err = cfg.GetCerts()
			case config.InfluxdbCategory:
				data, err = cfg.GetInfluxdb()
			case config.AlarmsCategory:
				data, err = cfg.GetAlarms()
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
	}
}

func newCreateOrUpdateConfigHandler(cfg *config.Config) elton.Handler {
	return func(c *elton.Context) (err error) {
		category := c.Param("category")
		var iconfig config.IConfig
		switch category {
		case config.CachesCategory:
			iconfig = cfg.NewCacheConfig("")
		case config.CompressesCategory:
			iconfig = cfg.NewCompressConfig("")
		case config.LocationsCategory:
			iconfig = cfg.NewLocationConfig("")
		case config.ServersCategory:
			iconfig = cfg.NewServerConfig("")
		case config.UpstreamsCategory:
			iconfig = cfg.NewUpstreamConfig("")
		case config.AdminCategory:
			iconfig = cfg.NewAdminConfig()
		case config.CertsCategory:
			iconfig = cfg.NewCertConfig("")
		case config.InfluxdbCategory:
			iconfig = cfg.NewInfluxdbConfig()
		case config.AlarmsCategory:
			iconfig = cfg.NewAlarmConfig("")
		default:
			err = hes.New(category + " is not support")
			return
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
	}
}

func newDeleteConfigHandler(cfg *config.Config) elton.Handler {
	return func(c *elton.Context) (err error) {
		serverConfigs, err := cfg.GetServers()
		if err != nil {
			return
		}
		locations, err := cfg.GetLocations()
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
			iconfig = cfg.NewCacheConfig(name)
		case config.CompressesCategory:
			shouldBeCheckedByServer = true
			iconfig = cfg.NewCompressConfig(name)
		case config.LocationsCategory:
			shouldBeCheckedByServer = true
			iconfig = cfg.NewLocationConfig(name)
		case config.ServersCategory:
			iconfig = cfg.NewServerConfig(name)
		case config.UpstreamsCategory:
			shouldBeCheckedByLocation = true
			iconfig = cfg.NewUpstreamConfig(name)
		case config.CertsCategory:
			shouldBeCheckedByServer = true
			iconfig = cfg.NewCertConfig(name)
		case config.InfluxdbCategory:
			iconfig = cfg.NewInfluxdbConfig()
		case config.AlarmsCategory:
			iconfig = cfg.NewAlarmConfig(name)
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
	}
}

// NewAdmin new an admin elton istance
func NewAdmin(opts *ServerOptions) (string, *elton.Elton) {
	cfg := opts.cfg
	adminConfig, _ := cfg.GetAdmin()
	if adminConfig == nil {
		return "", nil
	}
	adminPath := defaultAdminPath
	if adminConfig.Prefix != "" {
		adminPath = adminConfig.Prefix
	}

	e := elton.New()

	if adminConfig != nil {
		e.Use(newAdminValidateMiddlewares(adminConfig)...)
	}

	e.Use(compress.NewDefault())

	e.Use(middleware.NewDefaultFresh())
	e.Use(middleware.NewDefaultETag())

	e.Use(middleware.NewDefaultError())
	e.Use(middleware.NewDefaultResponder())
	e.Use(middleware.NewDefaultBodyParser())

	g := elton.NewGroup(adminPath)

	// 按分类获取配置
	g.GET("/configs/{category}", newGetConfigHandler(cfg))
	// 添加与更新使用相同处理
	g.POST("/configs/{category}", newCreateOrUpdateConfigHandler(cfg))
	// 删除配置
	g.DELETE("/configs/{category}/{name}", newDeleteConfigHandler(cfg))

	files := new(assetFiles)

	g.GET("/", func(c *elton.Context) (err error) {
		file := "index.html"
		buf, err := files.Get(file)
		if err != nil {
			return
		}
		c.SetContentTypeByExt(file)
		c.BodyBuffer = bytes.NewBuffer(buf)
		return
	})

	icons := []string{
		"favicon.ico",
		"logo.png",
	}
	handleIcon := func(icon string) {
		g.GET("/"+icon, func(c *elton.Context) (err error) {
			buf, err := application.DefaultAsset().Find(icon)
			if err != nil {
				return
			}
			c.SetContentTypeByExt(icon)
			c.Body = buf
			return
		})
	}
	for _, icon := range icons {
		handleIcon(icon)
	}

	g.GET("/static/*", middleware.NewStaticServe(files, middleware.StaticServeConfig{
		Path: "/static",
		// 客户端缓存一年
		MaxAge: 365 * 24 * 3600,
		// 缓存服务器缓存一个小时
		SMaxAge:             60 * 60,
		DenyQueryString:     true,
		DisableLastModified: true,
	}))

	// 获取应用状态
	g.GET("/application", func(c *elton.Context) error {
		c.Body = application.Default()
		return nil
	})

	// 获取upstream的状态
	g.GET("/upstreams", func(c *elton.Context) error {
		c.Body = opts.upstreams.Status()
		return nil
	})

	// 上传
	g.POST("/upload", func(c *elton.Context) (err error) {
		file, fileHeader, err := c.Request.FormFile("file")
		if err != nil {
			return
		}
		buf, err := ioutil.ReadAll(file)
		if err != nil {
			return
		}
		c.Body = map[string]string{
			"contentType": fileHeader.Header.Get("Content-Type"),
			"data":        base64.StdEncoding.EncodeToString(buf),
		}
		return
	})

	e.AddGroup(g)
	return adminPath, e
}
