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

// Upstream相关处理函数，根据策略选择适当的upstream server

package upstream

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/vicanso/elton"
	"github.com/vicanso/elton/middleware"
	"github.com/vicanso/hes"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/log"
	"github.com/vicanso/pike/util"
	us "github.com/vicanso/upstream"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
)

type (
	UpstreamServerConfig struct {
		// 服务地址，如 http://127.0.0.1:8080
		Addr string
		// 是否备用
		Backup bool
	}
	UpstreamServerStatus struct {
		Addr    string
		Healthy bool
	}
	UpstreamServerOption struct {
		Name        string
		HealthCheck string
		Policy      string
		// 是否启用h2c(http/2 over tcp)
		EnableH2C bool
		// 设置可接受的编码
		AcceptEncoding string
		// OnStatus on status
		OnStatus OnStatus
		Servers  []UpstreamServerConfig
	}
	upstreamServer struct {
		servers      []UpstreamServerConfig
		Proxy        elton.Handler
		HTTPUpstream *us.HTTP
		Option       *UpstreamServerOption
	}
	upstreamServers struct {
		m *sync.Map
	}
	StatusInfo struct {
		Name   string
		URL    string
		Status string
	}
	// OnStatus on status listener
	OnStatus func(StatusInfo)
)

var defaultUpstreamServers = NewUpstreamServers(nil)
var (
	ErrUpstreamNotFound = &hes.Error{
		StatusCode: http.StatusServiceUnavailable,
		Message:    "Upstream Not Found",
	}
)

// newTransport new a transport for http
func newTransport(h2c bool) http.RoundTripper {
	if h2c {
		return &http2.Transport{
			// 允许使用http的方式
			AllowHTTP: true,
			// tls的dial覆盖
			DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
		}
	}
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2: true,
		MaxIdleConns:      500,
		// 调整默认的每个host的最大连接因为缓存服务与backend可能会突发性的大量调用
		MaxIdleConnsPerHost:   50,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

// newTargetPicker create a target pick function
func newTargetPicker(uh *us.HTTP) middleware.ProxyTargetPicker {
	return func(c *elton.Context) (*url.URL, middleware.ProxyDone, error) {
		httpUpstream, done := uh.Next()
		if httpUpstream == nil {
			return nil, nil, ErrUpstreamNotFound
		}
		var proxyDone middleware.ProxyDone
		// 返回了done（如最少连接数的策略）
		if done != nil {
			proxyDone = func(_ *elton.Context) {
				done()
			}
		}
		return httpUpstream.URL, proxyDone, nil
	}
}

// newProxyMid new a proxy middleware
func newProxyMid(opt UpstreamServerOption, uh *us.HTTP) elton.Handler {
	return middleware.NewProxy(middleware.ProxyConfig{
		Transport:    newTransport(opt.EnableH2C),
		TargetPicker: newTargetPicker(uh),
	})
}

// NewUpstreamServer new an upstream server
func NewUpstreamServer(opt UpstreamServerOption) *upstreamServer {
	uh := &us.HTTP{
		Policy: opt.Policy,
		Ping:   opt.HealthCheck,
	}
	for _, server := range opt.Servers {
		// 添加失败的则忽略(地址配置有误则会添加失败)
		if server.Backup {
			_ = uh.AddBackup(server.Addr)
		} else {
			_ = uh.Add(server.Addr)
		}
	}
	// 如果有添加on status事件
	if opt.OnStatus != nil {
		uh.OnStatus(func(status int32, upstream *us.HTTPUpstream) {
			opt.OnStatus(StatusInfo{
				Name:   opt.Name,
				URL:    upstream.URL.String(),
				Status: us.ConvertStatusToString(status),
			})
		})
	}
	// 先执行一次health check，获取当前可用服务列表
	uh.DoHealthCheck()
	// 后续需要定时检测upstream是否可用
	go uh.StartHealthCheck()
	return &upstreamServer{
		servers:      opt.Servers,
		HTTPUpstream: uh,
		Option:       &opt,
		Proxy:        newProxyMid(opt, uh),
	}
}

// NewUpstreamServers new upstream servers
func NewUpstreamServers(opts []UpstreamServerOption) *upstreamServers {
	m := &sync.Map{}
	for _, opt := range opts {
		m.Store(opt.Name, NewUpstreamServer(opt))
	}
	return &upstreamServers{
		m: m,
	}
}

// Reset reset the upstream servers, remove not exists upstream servers and create new upstream server. If the upstream server is exists, then destroy the old one and add the new one.
func (us *upstreamServers) Reset(opts []UpstreamServerOption) {
	servers := util.MapDelete(us.m, func(key string) bool {
		// 如果不存在的，则删除
		exists := false
		for _, opt := range opts {
			if opt.Name == key {
				exists = true
				break
			}
		}
		return !exists
	})
	for _, item := range servers {
		server, _ := item.(*upstreamServer)
		if server != nil {
			server.Destroy()
		}
	}
	for _, opt := range opts {
		server := NewUpstreamServer(opt)
		currentServer := us.Get(opt.Name)
		// 先添加再删除
		us.m.Store(opt.Name, server)
		// 判断原来是否已存在此upstream server
		// 如果存在，则删除
		if currentServer != nil {
			currentServer.Destroy()
		}

	}
}

// Get get upstream server by name
func (us *upstreamServers) Get(name string) *upstreamServer {
	value, ok := us.m.Load(name)
	if !ok {
		return nil
	}
	server, ok := value.(*upstreamServer)
	if !ok {
		return nil
	}
	return server
}

// Destroy destory the upstream server
func (u *upstreamServer) Destroy() {
	// 停止定时检测
	u.HTTPUpstream.StopHealthCheck()
}

// GetServerStatusList get sever status list
func (u *upstreamServer) GetServerStatusList() []UpstreamServerStatus {
	statusList := make([]UpstreamServerStatus, 0)
	availableServers := u.HTTPUpstream.GetAvailableUpstreamList()
	for _, item := range u.servers {
		healthy := false
		for _, availableServer := range availableServers {
			if availableServer.URL.String() == item.Addr {
				healthy = true
			}
		}
		statusList = append(statusList, UpstreamServerStatus{
			Addr:    item.Addr,
			Healthy: healthy,
		})
	}

	return statusList
}

// Get get upstream server by name
func Get(name string) *upstreamServer {
	return defaultUpstreamServers.Get(name)
}

func onStatus(si StatusInfo) {
	log.Default().Info("upstream status change",
		zap.String("name", si.Name),
		zap.String("status", si.Status),
		zap.String("addr", si.URL),
	)
}

func convertConfigs(configs []config.UpstreamConfig, fn OnStatus) []UpstreamServerOption {
	opts := make([]UpstreamServerOption, 0)
	for _, item := range configs {
		servers := make([]UpstreamServerConfig, 0)
		for _, server := range item.Servers {
			servers = append(servers, UpstreamServerConfig{
				Addr:   server.Addr,
				Backup: server.Backup,
			})
		}
		opts = append(opts, UpstreamServerOption{
			Name:           item.Name,
			HealthCheck:    item.HealthCheck,
			Policy:         item.Policy,
			EnableH2C:      item.EnableH2C,
			AcceptEncoding: item.AcceptEncoding,
			Servers:        servers,
			OnStatus:       fn,
		})
	}
	return opts
}

// Reset reset the upstream server
func Reset(configs []config.UpstreamConfig) {
	ResetWithOnStats(configs, onStatus)
}

// ResetWithOnStats reset with on stats
func ResetWithOnStats(configs []config.UpstreamConfig, fn OnStatus) {
	defaultUpstreamServers.Reset(convertConfigs(configs, fn))
}
