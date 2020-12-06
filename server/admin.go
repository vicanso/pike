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

package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"time"

	"github.com/gobuffalo/packr/v2"
	"github.com/vicanso/elton"
	jwt "github.com/vicanso/elton-jwt"
	"github.com/vicanso/elton/middleware"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/log"
	"github.com/vicanso/pike/upstream"
	"github.com/vicanso/pike/util"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

type (
	AdminServerConfig struct {
		Addr     string
		User     string
		Password string
		Prefix   string
	}
	staticFile struct {
		box *packr.Box
	}
	loginParams struct {
		Account  string `json:"account,omitempty"`
		Password string `json:"password,omitempty"`
	}
	userInfo struct {
		Account string `json:"account,omitempty"`
	}

	// applicationInfo application info
	applicationInfo struct {
		GOARCH       string           `json:"goarch,omitempty"`
		GOOS         string           `json:"goos,omitempty"`
		GoVersion    string           `json:"goVersion,omitempty"`
		Version      string           `json:"version,omitempty"`
		BuildedAt    string           `json:"buildedAt,omitempty"`
		CommitID     string           `json:"commitID,omitempty"`
		Uptime       string           `json:"uptime,omitempty"`
		GoMaxProcs   int              `json:"goMaxProcs,omitempty"`
		RoutineCount int              `json:"routineCount,omitempty"`
		Processing   map[string]int32 `json:"processing,omitempty"`
	}
)

var (
	webBox   = packr.New("web", "../web")
	assetBox = packr.New("asset", "../asset")
)

var userNotLogin = util.NewError("Please login first", http.StatusUnauthorized)

var accountOrPasswordIsWrong = util.NewError("Account or password is wrong", http.StatusBadRequest)

var cacheKeyIsNil = util.NewError("The key of cache can't be null", http.StatusBadRequest)

var buildedAt time.Time
var commitID string
var startedAt = time.Now()

const jwtCookie = "pike"

// SetBuildInfo set build info
func SetBuildInfo(build, id string) {
	buildedAt, _ = time.Parse("20060102.150405", build)
	commitID = id
}

// Exists Test whether or not the given path exists
func (sf *staticFile) Exists(file string) bool {
	return sf.box.Has(file)
}

// Get Get data from file
func (sf *staticFile) Get(file string) ([]byte, error) {
	return sf.box.Find(file)
}

// Stat Get file's stat
func (sf *staticFile) Stat(file string) os.FileInfo {
	// 文件打包至程序中，因此无file info
	return nil
}

// NewReader Create a reader for file
func (sf *staticFile) NewReader(file string) (io.Reader, error) {
	buf, err := sf.Get(file)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buf), nil
}

func sendFile(c *elton.Context, file string) (err error) {
	data, err := webBox.Find(file)
	if err != nil {
		return
	}
	c.SetContentTypeByExt(file)
	c.CacheMaxAge(5 * time.Minute)
	c.BodyBuffer = bytes.NewBuffer(data)
	return
}

func getUserAccount(c *elton.Context) string {
	data := c.GetString(jwt.DefaultKey)
	params := userInfo{}
	_ = json.Unmarshal([]byte(data), &params)

	return params.Account
}

func newIsLoginHandler(user string) elton.Handler {
	return func(c *elton.Context) error {
		account := getUserAccount(c)
		if account == "" {
			return userNotLogin
		}
		return c.Next()
	}
}

func newLoginHandler(ttlToken *jwt.TTLToken, account, password string) elton.Handler {
	return func(c *elton.Context) (err error) {
		params := loginParams{}
		err = json.Unmarshal(c.RequestBody, &params)
		if err != nil {
			return
		}
		if params.Account != account || params.Password != password {
			err = accountOrPasswordIsWrong
			return
		}
		err = ttlToken.AddToCookie(c, &userInfo{
			Account: account,
		})
		if err != nil {
			return
		}

		c.Body = &userInfo{
			Account: account,
		}
		return
	}
}

func newUserMeHandler(user string) elton.Handler {
	return func(c *elton.Context) (err error) {
		account := "anonymous"
		if user != "" {
			account = getUserAccount(c)
		}
		c.Body = &userInfo{
			Account: account,
		}
		return nil
	}
}

func updateServerStatus(conf *config.PikeConfig) {
	// 数据需要复制
	upstreamServers := make([]config.UpstreamConfig, len(conf.Upstreams))
	for i, item := range conf.Upstreams {
		up := upstream.Get(item.Name)
		if len(item.Servers) == 0 || up == nil {
			upstreamServers[i] = item
			continue
		}
		p := &item
		statusList := up.GetServerStatusList()
		servers := make([]config.UpstreamServerConfig, len(p.Servers))
		// 填充upstream server的状态
		for j, server := range p.Servers {
			for _, status := range statusList {
				if server.Addr == status.Addr {
					server.Healthy = status.Healthy
				}
			}
			servers[j] = server
		}
		p.Servers = servers
		upstreamServers[i] = *p
	}
	conf.Upstreams = upstreamServers
}

// getConfig 获取config配置
func getConfig(c *elton.Context) (err error) {
	conf, err := config.Read()
	if err != nil {
		return
	}
	updateServerStatus(conf)
	c.Body = conf
	return nil
}

// saveConfig 保存config配置
func saveConfig(c *elton.Context) (err error) {
	conf := config.PikeConfig{}
	err = json.Unmarshal(c.RequestBody, &conf)
	if err != nil {
		return
	}
	err = config.Write(&conf)
	if err != nil {
		return
	}
	data, _ := yaml.Marshal(conf)
	conf.YAML = string(data)
	// 简单的等待1秒后再更新状态
	// 这样有可能检测到配置有更新，重新加载
	delay := c.QueryParam("delay")
	if delay != "" {
		v, _ := time.ParseDuration(delay)
		if v == 0 || v > 5*time.Second {
			v = time.Second
		}
		time.Sleep(v)
	}
	updateServerStatus(&conf)
	// 因为yaml部分要根据配置数据重新生成，因此重新读取返回
	c.Body = conf
	return
}

// getApplicationInfo 获取应用信息
func getApplicationInfo(c *elton.Context) (err error) {
	processing := make(map[string]int32)
	defaultServers.m.Range(func(key, value interface{}) bool {
		if key == nil || value == nil {
			return true
		}
		name, ok := key.(string)
		if !ok {
			return true
		}
		s, ok := value.(*server)
		if !ok {
			return true
		}
		processing[name] = s.processing.Load()
		return true
	})
	seconds := time.Duration(time.Since(startedAt).Seconds())
	d := time.Second * seconds
	version, _ := assetBox.Find("version")
	c.Body = &applicationInfo{
		GOARCH:       runtime.GOARCH,
		GOOS:         runtime.GOOS,
		GoVersion:    runtime.Version(),
		Version:      string(version),
		BuildedAt:    buildedAt.Format(time.RFC3339),
		CommitID:     commitID,
		Uptime:       d.String(),
		GoMaxProcs:   runtime.GOMAXPROCS(0),
		RoutineCount: runtime.NumGoroutine(),
		Processing:   processing,
	}
	return
}

// removeCache 删除缓存
func removeCache(c *elton.Context) (err error) {
	key := c.QueryParam("key")
	if key == "" {
		err = cacheKeyIsNil
		return
	}
	cache.RemoveHTTPCache(c.QueryParam("cache"), []byte(key))
	c.NoContent()
	return
}

// StartAdminServer start admin server
func StartAdminServer(config AdminServerConfig) (err error) {
	logger := log.Default()
	ttlToken := &jwt.TTLToken{
		TTL: 24 * time.Hour,
		// 密钥用于加密数据，需保密
		Secret:     []byte(config.Password),
		CookieName: jwtCookie,
	}

	// Passthrough为false，会校验token是否正确
	jwtNormal := jwt.NewJWT(jwt.Config{
		CookieName: jwtCookie,
		Decode:     ttlToken.Decode,
	})
	// 用于初始化创建token使用（此时可能token还没有或者已过期)
	jwtPassthrough := jwt.NewJWT(jwt.Config{
		CookieName:  jwtCookie,
		Decode:      ttlToken.Decode,
		Passthrough: true,
	})

	e := elton.New()
	sf := &staticFile{
		box: webBox,
	}

	e.Use(func(c *elton.Context) error {
		// 全局设置为不可缓存，后续可覆盖
		c.NoCache()

		// cors
		c.SetHeader("Access-Control-Allow-Credentials", "true")
		c.SetHeader("Access-Control-Allow-Origin", "http://127.0.0.1:3123")
		c.SetHeader("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
		c.SetHeader("Access-Control-Allow-Headers", "Content-Type, Accept")
		c.SetHeader("Access-Control-Max-Age", "86400")
		return c.Next()
	})

	e.Use(middleware.NewError(middleware.ErrorConfig{
		ResponseType: "json",
	}))
	e.Use(middleware.NewStats(middleware.StatsConfig{
		OnStats: func(info *middleware.StatsInfo, _ *elton.Context) {
			logger.Info("access log",
				zap.String("ip", info.IP),
				zap.String("method", info.Method),
				zap.String("uri", info.URI),
				zap.Int("status", info.Status),
				zap.String("consuming", info.Consuming.String()),
				zap.Int("bytes", info.Size),
			)
		},
	}))

	compressConfig := middleware.NewCompressConfig(new(middleware.GzipCompressor))
	compressConfig.Checker = regexp.MustCompile("text|javascript|json|wasm|font")
	e.Use(middleware.NewCompress(compressConfig))
	e.Use(middleware.NewDefaultBodyParser())
	e.Use(middleware.NewDefaultResponder())

	// 获取、更新配置
	var isLogin elton.Handler
	if config.User != "" {
		isLogin = elton.Compose(jwtNormal, newIsLoginHandler(config.User))
	} else {
		isLogin = func(c *elton.Context) error {
			return c.Next()
		}
	}
	e.GET("/config", isLogin, getConfig)
	e.PUT("/config", isLogin, saveConfig)

	// 登录
	e.POST("/login", jwtPassthrough, newLoginHandler(ttlToken, config.User, config.Password))
	// 用户信息
	e.GET("/me", jwtPassthrough, newUserMeHandler(config.User))

	e.GET("/application-info", getApplicationInfo)

	// 缓存
	e.DELETE("/cache", removeCache)

	e.GET("/ping", func(c *elton.Context) error {
		c.BodyBuffer = bytes.NewBufferString("pong")
		return nil
	})
	// 静态文件
	e.GET("/", func(c *elton.Context) error {
		return sendFile(c, "index.html")
	})
	e.GET("/*", middleware.NewStaticServe(sf, middleware.StaticServeConfig{
		// 客户端缓存一年
		MaxAge: 365 * 24 * time.Hour,
		// 缓存服务器缓存一个小时
		SMaxAge:             time.Hour,
		DisableLastModified: true,
	}))

	// cors设置
	e.OPTIONS("/*", func(c *elton.Context) error {
		c.NoContent()
		return nil
	})
	logger.Info("start admin server",
		zap.String("addr", config.Addr),
	)
	return e.ListenAndServe(config.Addr)
}
