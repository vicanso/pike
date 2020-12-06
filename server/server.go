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
	"net"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/vicanso/elton"
	"github.com/vicanso/elton/middleware"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/log"
	"github.com/vicanso/pike/util"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

type (
	// server pike server
	server struct {
		mutex                     *sync.RWMutex
		logFormat                 string
		listening                 bool
		listenAddr                string
		addr                      string
		locations                 []string
		cache                     string
		compress                  string
		compressMinLength         int
		compressContentTypeFilter *regexp.Regexp
		processing                atomic.Int32
		ln                        net.Listener
		e                         *elton.Elton
	}
	// servers pike server list
	servers struct {
		m *sync.Map
	}
	ServerOption struct {
		// 访问日志格式化
		LogFormat string
		// 监听地址
		Addr string
		// 使用的location列表
		Locations []string
		// 使用的缓存
		Cache string
		// 使用的压缩服务
		Compress string
		// 压缩最小尺寸
		CompressMinLength int
		// 压缩数据类型
		CompressContentTypeFilter *regexp.Regexp
	}
)

const (
	// statusKey 保存该请求对应的status: fetching, pass 等
	statusKey = "_status"
	// httpRespKey 保存请求对应的响应数据
	httpRespKey = "_httpResp"
	// httpRespAgeKey 保存缓存响应的age
	httpRespAgeKey = "_httpRespAge"
	// httpCacheMaxAgeKey 缓存有效期
	httpCacheMaxAgeKey = "_httpCacheMaxAge"
)

const defaultCompressMinLength = 1024

var defaultServers = NewServers(nil)

const (
	headerAge         = "Age"
	headerCacheStatus = "X-Status"
)

var (
	ErrInvalidResponse = util.NewError("Invalid response", http.StatusServiceUnavailable)

	ErrCacheDispatcherNotFound = util.NewError("Available cache dispatcher not found", http.StatusServiceUnavailable)

	ErrLocationNotFound = util.NewError("Available location not found", http.StatusServiceUnavailable)

	ErrUpstreamNotFound = util.NewError("Available upstream not found", http.StatusBadGateway)
)

func getCacheStatus(c *elton.Context) cache.Status {
	return cache.Status(c.GetInt(statusKey))
}
func setCacheStatus(c *elton.Context, cacheStatus cache.Status) {
	c.Set(statusKey, int(cacheStatus))
}

func getHTTPResp(c *elton.Context) *cache.HTTPResponse {
	value, exists := c.Get(httpRespKey)
	if !exists {
		return nil
	}
	resp, ok := value.(*cache.HTTPResponse)
	if !ok {
		return nil
	}
	return resp
}
func setHTTPResp(c *elton.Context, resp *cache.HTTPResponse) {
	c.Set(httpRespKey, resp)
}

func setHTTPRespAge(c *elton.Context, age int) {
	c.Set(httpRespAgeKey, age)
}
func getHTTPRespAge(c *elton.Context) int {
	return c.GetInt(httpRespAgeKey)
}

func setHTTPCacheMaxAge(c *elton.Context, age int) {
	c.Set(httpCacheMaxAgeKey, age)
}
func getHTTPCacheMaxAge(c *elton.Context) int {
	return c.GetInt(httpCacheMaxAgeKey)
}

// NewServer create a new server
func NewServer(opt ServerOption) *server {
	minLength := opt.CompressMinLength
	// 如果未设置最少压缩长度，则设置为1KB
	if minLength == 0 {
		minLength = defaultCompressMinLength
	}
	return &server{
		mutex:                     &sync.RWMutex{},
		logFormat:                 opt.LogFormat,
		addr:                      opt.Addr,
		locations:                 opt.Locations,
		cache:                     opt.Cache,
		compress:                  opt.Compress,
		compressMinLength:         minLength,
		compressContentTypeFilter: opt.CompressContentTypeFilter,
	}
}

// NewServers create new server list
func NewServers(opts []ServerOption) *servers {
	m := &sync.Map{}
	for _, opt := range opts {
		m.Store(opt.Addr, NewServer(opt))
	}
	return &servers{
		m: m,
	}
}

// Start start all server
func (ss *servers) Start() (err error) {
	ss.m.Range(func(key, value interface{}) bool {
		s, ok := value.(*server)
		if ok {
			err := s.Start(true)
			if err != nil {
				log.Default().Error("server start fail",
					zap.String("addr", s.addr),
					zap.Error(err),
				)
			}
		}
		return true
	})
	return nil
}

// Reset reset server list
func (ss *servers) Reset(opts []ServerOption) {
	// 删除不再存在的server
	result := util.MapDelete(ss.m, func(key string) bool {
		exists := false
		for _, opt := range opts {
			if opt.Addr == key {
				exists = true
				break
			}
		}
		return !exists
	})
	for _, item := range result {
		s, _ := item.(*server)
		if s != nil {
			// 由于close需要等待，因此切换时，使用goroutine来关闭
			go func() {
				err := s.Close()
				if err != nil {
					log.Default().Error("close server fail",
						zap.String("addr", s.addr),
						zap.Error(err),
					)
				}
			}()
		}
	}
	for _, opt := range opts {
		value, ok := ss.m.Load(opt.Addr)
		// 如果该服务存在，则修改属性
		if ok {
			s, _ := value.(*server)
			if s != nil {
				s.Update(opt)
			}
		} else {
			ss.m.Store(opt.Addr, NewServer(opt))
		}
	}
}

// Close close the server list
func (ss *servers) Close() error {
	ss.m.Range(func(_, value interface{}) bool {
		s, _ := value.(*server)
		if s != nil {
			err := s.Close()
			log.Default().Error("close server fail",
				zap.String("addr", s.addr),
				zap.Error(err),
			)
		}
		return true
	})
	return nil
}

// Get get server for server list
func (ss *servers) Get(name string) *server {
	value, ok := ss.m.Load(name)
	if !ok {
		return nil
	}
	s, ok := value.(*server)
	if !ok {
		return nil
	}
	return s
}

// Update 更新配置
func (s *server) Update(opt ServerOption) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.locations = opt.Locations
	s.cache = opt.Cache
	s.compress = opt.Compress
	s.compressMinLength = opt.CompressMinLength
	s.compressContentTypeFilter = opt.CompressContentTypeFilter
}

// GetCache get the cache of server
func (s *server) GetCache() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.cache
}

// GetLocations get the locations of server
func (s *server) GetLocations() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.locations
}

// GetCompress get the compress option of server
func (s *server) GetCompress() (name string, minLength int, filter *regexp.Regexp) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.compress, s.compressMinLength, s.compressContentTypeFilter
}

// Start start the server
func (s *server) Start(useGoRoutine bool) (err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	// 如监听中，则直接返回
	if s.listening {
		return
	}

	logger := log.Default()

	// TODO 如果发生panic，停止处理新请求，程序退出
	e := elton.New()
	if s.logFormat != "" {
		e.Use(middleware.NewLogger(middleware.LoggerConfig{
			OnLog: func(str string, _ *elton.Context) {
				logger.Info(str)
			},
			Format: s.logFormat,
		}))
	}
	e.Use(func(c *elton.Context) error {
		s.processing.Add(1)
		defer s.processing.Dec()
		return c.Next()
	})
	// TODO 考虑是否自定义出错中间件，对于系统的error(category: "pike")触发告警
	e.Use(middleware.NewDefaultError())
	e.Use(middleware.NewDefaultFresh())
	e.Use(NewResponder())
	e.Use(NewCache(s))
	e.Use(NewProxy(s))
	e.GET("/ping", func(c *elton.Context) error {
		c.BodyBuffer = bytes.NewBufferString("pong")
		return nil
	})
	e.ALL("/*", func(c *elton.Context) error {
		return nil
	})
	// TODO 一般使用时，pike的前置还有nginx或haproxy，
	// 因此与客户端的各类超时由前置反向代理处理，
	// 后续确认是否需要增加更多的参数设置，
	// 如ReadTimeout ReadHeaderTimeout等，
	srv := &http.Server{
		Handler: e,
	}
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return
	}
	s.listening = true
	s.e = e
	s.ln = ln
	s.listenAddr = ln.Addr().String()
	if !useGoRoutine {
		return srv.Serve(ln)
	}
	go func() {
		err := srv.Serve(ln)
		log.Default().Error("server serve fail",
			zap.String("addr", s.addr),
			zap.Error(err),
		)
	}()
	return nil
}

// Close close the server
func (s *server) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if !s.listening {
		return nil
	}
	s.listening = false
	err := s.e.GracefulClose(10 * time.Second)
	if err != nil {
		return err
	}
	return s.ln.Close()
}

// GetAddr get listen addr of server
func (s *server) GetListenAddr() string {
	return s.listenAddr
}

func convertConfig(configs []config.ServerConfig) []ServerOption {
	opts := make([]ServerOption, 0)
	for _, item := range configs {
		minLength, _ := humanize.ParseBytes(item.CompressMinLength)
		var reg *regexp.Regexp
		// 如果有配置则生成
		if item.CompressContentTypeFilter != "" {
			reg, _ = regexp.Compile(item.CompressContentTypeFilter)
		}
		opts = append(opts, ServerOption{
			LogFormat:                 item.LogFormat,
			Addr:                      item.Addr,
			Locations:                 item.Locations,
			Cache:                     item.Cache,
			Compress:                  item.Compress,
			CompressMinLength:         int(minLength),
			CompressContentTypeFilter: reg,
		})
	}
	return opts
}

// Reset reset the default server list
func Reset(configs []config.ServerConfig) {
	defaultServers.Reset(convertConfig(configs))
}

// Get get server from default server list
func Get(name string) *server {
	return defaultServers.Get(name)
}

// Start start the default server list
func Start() error {
	return defaultServers.Start()
}

// CLose close the default server list
func Close() error {
	return defaultServers.Close()
}
