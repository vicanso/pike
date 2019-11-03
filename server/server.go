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
	"net/http"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/vicanso/elton"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/upstream"
)

const (
	serverStatusNotRunning = iota
	serverStatusRunning
	serverStatusStop
)

// Server http server
type Server struct {
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
}

// Instance pike server instance
type Instance struct {
	servers   *sync.Map
	upstreams *upstream.Upstreams
}

// NewInstance new a pike server instance
func NewInstance() (ins *Instance, err error) {
	// TODO 对于未有任何配置信息时，增加默认的server
	serversConfig, err := config.GetServers()
	if err != nil {
		return
	}

	cachesConfig, err := config.GetCaches()
	if err != nil {
		return
	}
	dispatchers := cache.NewDispatchers(cachesConfig)

	locationsConfig, err := config.GetLocations()
	if err != nil {
		return
	}
	locationsConfig.Sort()

	upstreamsConfig, err := config.GetUpstreams()
	if err != nil {
		return
	}
	upstreams := upstream.NewUpstreams(upstreamsConfig)

	compressesConfig, err := config.GetCompresses()
	if err != nil {
		return
	}

	servers := new(sync.Map)
	for _, conf := range serversConfig {
		result := locationsConfig.Filter(conf.Locations...)
		dispatcher := dispatchers.Get(conf.Cache)
		srv := NewServer(conf, result, upstreams, dispatcher, compressesConfig.Get(conf.Compress))
		servers.Store(conf.Name, srv)
	}
	ins = &Instance{
		servers:   servers,
		upstreams: upstreams,
	}
	return
}

// Start start all server
func (ins *Instance) Start() {
	ins.servers.Range(func(k, v interface{}) bool {
		srv, ok := v.(*Server)
		if ok {
			go srv.ListenAndServe()
		}
		return true
	})
}

// NewServer new a server
func NewServer(conf *config.Server, locations config.Locations, upstreams *upstream.Upstreams, dispatcher *cache.Dispatcher, compress *config.Compress) *Server {
	server := &http.Server{
		Addr:              conf.Addr,
		ReadTimeout:       conf.ReadTimeout,
		ReadHeaderTimeout: conf.ReadHeaderTimeout,
		WriteTimeout:      conf.WriteTimeout,
		IdleTimeout:       conf.IdleTimeout,
		MaxHeaderBytes:    conf.MaxHeaderBytes,
	}
	srv := &Server{
		maxConcurrency: conf.Concurrency,
		eTag:           conf.ETag,
		server:         server,
		locations:      locations,
		upstreams:      upstreams,
		dispatcher:     dispatcher,
		compress:       compress,
	}
	srv.ToggleElton()
	server.Handler = srv
	return srv
}

// ListenAndServe call http server's listen and serve
func (s *Server) ListenAndServe() error {
	if s.status == serverStatusRunning {
		return nil
	}
	s.status = serverStatusRunning
	err := s.server.ListenAndServe()
	if err != nil {
		s.message = err.Error()
	}
	s.status = serverStatusStop
	return err
}

// ToggleElton toggle elton
func (s *Server) ToggleElton() *elton.Elton {
	e := NewElton(&EltonConfig{
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
