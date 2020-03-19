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

// The http server of pike

package server

import (
	"crypto/tls"
	"encoding/base64"
	"net/http"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/vicanso/elton"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/log"
	"github.com/vicanso/pike/upstream"
	"go.uber.org/zap"
)

const (
	serverStatusNotRunning = iota // nolint
	serverStatusRunning
	serverStatusStop
)

// Server http server
type Server struct {
	name           string
	influxSrv      *InfluxSrv
	message        string
	concurrency    uint32
	eTag           bool
	maxConcurrency uint32
	status         int32
	server         *http.Server
	e              *elton.Elton
	locations      config.Locations
	upstreams      *upstream.Upstreams
	dispatcher     *cache.Dispatcher
	compress       *config.Compress
	cfg            *config.Config
}

// ServerOptions server options
type ServerOptions struct {
	influxSrv  *InfluxSrv
	server     *config.Server
	locations  config.Locations
	upstreams  *upstream.Upstreams
	dispatcher *cache.Dispatcher
	compress   *config.Compress
	cfg        *config.Config
}

// Instance pike server instance
type Instance struct {
	InfluxSrv          *InfluxSrv
	Config             *config.Config
	EnabledAdminServer bool
	servers            *sync.Map
	upstreams          *upstream.Upstreams
}

// Fetch fetch config for instance
func (ins *Instance) Fetch() (err error) {
	cfg := ins.Config
	serversConfig, err := cfg.GetServers()
	if err != nil {
		return
	}
	if ins.EnabledAdminServer {
		serversConfig = append(serversConfig, &config.Server{
			Name: "admin",
			Addr: ":3015",
		})
	}

	influxdbConfig, err := cfg.GetInfluxdb()
	if err != nil {
		return
	}
	var influxSrv *InfluxSrv
	if influxdbConfig != nil && influxdbConfig.Enabled {
		influxSrv, err = NewInfluxSrv(influxdbConfig)
	}
	// 初始化influxdb失败只输出日志
	if err != nil {
		log.Default().Error("create influxdb service fail",
			zap.Error(err),
		)
	}
	if ins.InfluxSrv != nil {
		ins.InfluxSrv.Flush()
	}
	ins.InfluxSrv = influxSrv

	cachesConfig, err := cfg.GetCaches()
	if err != nil {
		return
	}
	dispatchers := cache.NewDispatchers(cachesConfig)

	locationsConfig, err := cfg.GetLocations()
	if err != nil {
		return
	}
	locationsConfig.Sort()

	upstreamsConfig, err := cfg.GetUpstreams()
	if err != nil {
		return
	}
	upstreams := upstream.NewUpstreams(upstreamsConfig)

	compressesConfig, err := cfg.GetCompresses()
	if err != nil {
		return
	}
	servers := ins.servers
	if servers == nil {
		servers = new(sync.Map)
	}
	for _, conf := range serversConfig {
		locations := locationsConfig.Filter(conf.Locations...)
		dispatcher := dispatchers.Get(conf.Cache)
		compress := compressesConfig.Get(conf.Compress)

		data, ok := servers.Load(conf.Name)
		// 如果已存在，仅更新信息
		if ok {
			srv := data.(*Server)
			srv.influxSrv = influxSrv
			srv.Update(conf, locations, upstreams, dispatcher, compress)
		} else {
			opts := &ServerOptions{
				influxSrv:  influxSrv,
				server:     conf,
				locations:  locations,
				upstreams:  upstreams,
				dispatcher: dispatcher,
				compress:   compress,
				cfg:        cfg,
			}
			srv := NewServer(opts)
			servers.Store(conf.Name, srv)
		}
	}
	oldUpstreams := ins.upstreams
	// 如果已有upstreams存在，则将原有upstream销毁
	if oldUpstreams != nil {
		oldUpstreams.Destroy()
	}
	ins.servers = servers
	ins.upstreams = upstreams
	return
}

// Restart restart all server
func (ins *Instance) Restart() {
	ins.servers.Range(func(k, v interface{}) bool {
		srv, ok := v.(*Server)
		if ok {
			go func() {
				_ = srv.ListenAndServe()
			}()
		}
		return true
	})
}

// Start start all server
func (ins *Instance) Start() (err error) {
	err = ins.Fetch()
	if err != nil {
		return
	}
	// restart 根据当前配置重新启动server
	ins.Restart()
	return
}

// NewServer new a server
func NewServer(opts *ServerOptions) *Server {
	conf := opts.server
	srv := &Server{
		name:      conf.Name,
		cfg:       opts.cfg,
		influxSrv: opts.influxSrv,
	}
	var tlsConfig *tls.Config
	if len(conf.Certs) != 0 {
		tlsConfig = &tls.Config{}
		tlsConfig.Certificates = make([]tls.Certificate, 0)
		for _, name := range conf.Certs {
			c := opts.cfg.NewCertConfig(name)
			err := c.Fetch()
			if err != nil {
				continue
			}
			key, err := base64.StdEncoding.DecodeString(c.Key)
			if err != nil {
				continue
			}
			cert, err := base64.StdEncoding.DecodeString(c.Cert)
			if err != nil {
				continue
			}
			certificate, err := tls.X509KeyPair(cert, key)
			if err != nil {
				continue
			}
			tlsConfig.Certificates = append(tlsConfig.Certificates, certificate)
		}
	}

	server := &http.Server{
		Addr:              conf.Addr,
		ReadTimeout:       conf.ReadTimeout,
		ReadHeaderTimeout: conf.ReadHeaderTimeout,
		WriteTimeout:      conf.WriteTimeout,
		IdleTimeout:       conf.IdleTimeout,
		MaxHeaderBytes:    conf.MaxHeaderBytes,
		Handler:           srv,
	}

	if tlsConfig != nil {
		server.TLSConfig = tlsConfig.Clone()
	}
	srv.server = server

	srv.Update(opts.server, opts.locations, opts.upstreams, opts.dispatcher, opts.compress)
	return srv
}

// Update update server config
func (s *Server) Update(conf *config.Server, locations config.Locations, upstreams *upstream.Upstreams, dispatcher *cache.Dispatcher, compress *config.Compress) {
	s.maxConcurrency = conf.Concurrency
	s.eTag = conf.ETag
	s.locations = locations
	s.dispatcher = dispatcher
	s.compress = compress
	s.upstreams = upstreams
	s.toggleElton()
}

// ListenAndServe call http server's listen and serve
func (s *Server) ListenAndServe() (err error) {
	if s.GetStatus() == serverStatusRunning {
		return nil
	}
	s.SetStatus(serverStatusRunning)

	log.Default().Info("server listening",
		zap.String("addr", s.server.Addr),
	)
	if s.server.TLSConfig != nil {
		err = s.server.ListenAndServeTLS("", "")
	} else {
		err = s.server.ListenAndServe()
	}
	if err != nil {
		s.message = err.Error()
	}
	s.SetStatus(serverStatusStop)
	return
}

// toggleElton toggle elton
func (s *Server) toggleElton() *elton.Elton {
	e := NewElton(&EltonOptions{
		name:           s.name,
		influxSrv:      s.influxSrv,
		cfg:            s.cfg,
		eTag:           s.eTag,
		maxConcurrency: s.concurrency,
		locations:      s.locations,
		upstreams:      s.upstreams,
		dispatcher:     s.dispatcher,
		compress:       s.compress,
	})
	return s.SetElton(e)
}

// GetElton get elton from server
func (s *Server) GetElton() *elton.Elton {
	return (*elton.Elton)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&s.e))))
}

// SetElton set elton to server
func (s *Server) SetElton(e *elton.Elton) (oldElton *elton.Elton) {
	oldPoint := atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&s.e)), unsafe.Pointer(e))
	if oldPoint == nil {
		return
	}
	oldElton = (*elton.Elton)(oldPoint)
	return
}

// GetStatus get status of server
func (s *Server) GetStatus() int32 {
	return atomic.LoadInt32(&s.status)
}

// SetStatus set status of server
func (s *Server) SetStatus(status int32) {
	atomic.StoreInt32(&s.status, status)
}

// ServeHTTP serve http
func (s *Server) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	s.GetElton().ServeHTTP(resp, req)
}
