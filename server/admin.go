package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"os"
	"regexp"
	"runtime"

	"github.com/vicanso/hes"

	"github.com/gobuffalo/packr/v2"
	"github.com/vicanso/cod"
	basicauth "github.com/vicanso/cod-basic-auth"
	bodyparser "github.com/vicanso/cod-body-parser"
	compress "github.com/vicanso/cod-compress"
	errorhandler "github.com/vicanso/cod-error-handler"
	etag "github.com/vicanso/cod-etag"
	fresh "github.com/vicanso/cod-fresh"
	recover "github.com/vicanso/cod-recover"
	responder "github.com/vicanso/cod-responder"
	staticServe "github.com/vicanso/cod-static-serve"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/df"
	"github.com/vicanso/pike/stats"
	"github.com/vicanso/pike/upstream"
)

type (
	staticFile struct {
		box *packr.Box
	}
	// ApplicationInfo applicatin information
	ApplicationInfo struct {
		// version版本号
		Version string `json:"version"`
		// 程序启动时间
		StartedAt string `json:"startedAt"`
		// 构建时间
		BuildedAt string `json:"buildedAt"`
		// 当前版本的git commit id
		CommitID string `json:"commitId"`
		// 编译的go版本
		GoVersion string `json:"goVersion"`
	}
)

var (
	box = packr.New("asset", "../web/build")
)

func (sf *staticFile) Exists(file string) bool {
	return sf.box.Has(file)
}
func (sf *staticFile) Get(file string) ([]byte, error) {
	return sf.box.Find(file)
}
func (sf *staticFile) Stat(file string) os.FileInfo {
	return nil
}
func sendFile(c *cod.Context, file string) (err error) {
	buf, err := box.Find(file)
	if err != nil {
		return
	}
	c.SetContentTypeByExt(file)
	c.BodyBuffer = bytes.NewBuffer(buf)
	return
}

func doSha256(str string) string {
	hash := sha256.New()
	hash.Write([]byte(str))
	hashBytes := hash.Sum(nil)
	return base64.StdEncoding.EncodeToString(hashBytes)
}

// NewAdminServer create an admin server
func NewAdminServer(opts Options) *cod.Cod {
	cfg := opts.BasicConfig
	insStats := opts.Stats
	director := opts.Director
	dsp := opts.Dispatcher
	directorConfig := opts.DirectorConfig

	d := cod.New()
	d.Use(func(c *cod.Context) error {
		c.NoCache()
		return c.Next()
	})
	d.Use(recover.New())
	d.Use(errorhandler.NewDefault())
	d.Use(fresh.NewDefault())
	d.Use(etag.NewDefault())
	d.Use(compress.NewDefault())
	d.Use(bodyparser.NewDefault())
	d.Use(responder.NewDefault())

	adminHandlerList := make([]cod.Handler, 0)

	adminConfig := cfg.Data.Admin
	adminUser := adminConfig.User
	adminPwd := adminConfig.Password
	// 设置 basic auth 认证
	if adminUser != "" && adminPwd != "" {
		adminHandlerList = append(adminHandlerList, basicauth.New(basicauth.Config{
			Validate: func(account, pwd string, _ *cod.Context) (bool, error) {
				if account == adminUser && (pwd == adminPwd ||
					doSha256(pwd) == adminPwd) {
					return true, nil
				}
				return false, nil
			},
		}))
	}

	g := cod.NewGroup("", adminHandlerList...)
	// 获取系统状态统计
	g.GET("/stats", func(c *cod.Context) error {
		c.Body = &struct {
			Stats *stats.Info `json:"stats,omitempty"`
		}{
			insStats.GetInfo(),
		}
		return nil
	})
	// 获取 upstream 列表
	g.GET("/upstreams", func(c *cod.Context) error {
		c.Body = &struct {
			Upstreams []upstream.Info `json:"upstreams,omitempty"`
		}{
			director.GetUpstreamInfos(),
		}
		return nil
	})

	// 获取单个upstream信息
	g.GET("/upstreams/:name", func(c *cod.Context) error {
		name := c.Param("name")
		infos := director.GetUpstreamInfos()
		for _, item := range infos {
			if item.Name == name {
				c.Body = item
			}
		}
		if c.Body == nil {
			return hes.New("upstream's name is invalid")
		}
		return nil
	})

	// 增加upstream
	g.POST("/upstreams", func(c *cod.Context) (err error) {
		backend := config.BackendConfig{}
		err = json.Unmarshal(c.RequestBody, &backend)
		if err != nil {
			err = hes.Wrap(err)
			return
		}
		if backend.Name == "" || len(backend.Backends) == 0 {
			err = hes.New("name and backends can't be nil")
			return
		}
		err = directorConfig.AddBackend(backend)
		if err != nil {
			return
		}
		err = directorConfig.WriteConfig()
		if err != nil {
			return
		}
		c.Created(backend)
		return nil
	})

	g.PATCH("/upstreams/:name", func(c *cod.Context) (err error) {
		backend := config.BackendConfig{}
		err = json.Unmarshal(c.RequestBody, &backend)
		if err != nil {
			err = hes.Wrap(err)
			return
		}
		backend.Name = c.Param("name")
		err = directorConfig.UpdateBackend(backend)
		if err != nil {
			return
		}
		err = directorConfig.WriteConfig()
		if err != nil {
			return
		}
		c.NoContent()
		return nil
	})

	// 删除upstream
	g.DELETE("/upstreams/:name", func(c *cod.Context) (err error) {
		directorConfig.RemoveBackend(c.Param("name"))
		err = directorConfig.WriteConfig()
		if err != nil {
			return
		}
		c.NoContent()
		return nil
	})

	// 获取缓存列表
	g.GET("/caches", func(c *cod.Context) error {
		c.Body = &struct {
			Caches []*cache.HTTPCacheInfo `json:"caches,omitempty"`
		}{
			dsp.GetCacheList(),
		}
		return nil
	})

	// 删除缓存
	g.DELETE("/caches", func(c *cod.Context) error {
		dsp.Expire([]byte(c.QueryParam("key")))
		c.NoContent()
		return nil
	})

	// 获取配置列表
	g.GET("/configs", func(c *cod.Context) (err error) {
		basicYaml, err := opts.BasicConfig.YAML()
		if err != nil {
			return
		}
		// 替换password
		reg := regexp.MustCompile(`password:[\s\S]+?\n`)
		basicYaml = reg.ReplaceAll(basicYaml, []byte("password: ***\n"))
		directorYaml, err := opts.DirectorConfig.YAML()
		if err != nil {
			return
		}
		c.Body = &struct {
			ApplicationInfo *ApplicationInfo `json:"applicationInfo,omitempty"`
			BasicYaml       string           `json:"basicYaml,omitempty"`
			DirectorYaml    string           `json:"directorYaml,omitempty"`
		}{
			&ApplicationInfo{
				Version:   df.Version,
				BuildedAt: df.BuildedAt,
				StartedAt: df.StartedAt,
				CommitID:  df.CommitID,
				GoVersion: runtime.Version(),
			},
			string(basicYaml),
			string(directorYaml),
		}
		return
	})

	// 获取基础配置信息
	g.GET("/configs/:name", func(c *cod.Context) (err error) {
		if c.Param("name") != "basic" {
			err = hes.New("Only support to get basic config")
			return
		}
		data := opts.BasicConfig.Data
		data.Admin.Password = "***"
		c.Body = data
		return
	})

	// 更新基础配置信息
	g.PATCH("/configs/basic", func(c *cod.Context) (err error) {
		data := config.BasicConfig{}
		err = json.Unmarshal(c.RequestBody, &data)
		if err != nil {
			err = hes.Wrap(err)
			return
		}
		opts.BasicConfig.Data = data
		err = opts.BasicConfig.WriteConfig()
		if err != nil {
			return
		}
		c.NoContent()
		return
	})

	sf := &staticFile{
		box: box,
	}
	// 静态文件

	g.GET("/", func(c *cod.Context) error {
		c.CacheMaxAge("10s")
		return sendFile(c, "index.html")
	})
	g.GET("/static/*file", staticServe.New(sf, staticServe.Config{
		Path: "/static",
		// 客户端缓存一年
		MaxAge: 365 * 24 * 3600,
		// 缓存服务器缓存一个小时
		SMaxAge:             60 * 60,
		DenyQueryString:     true,
		DisableLastModified: true,
	}))

	d.AddGroup(g)
	return d
}
