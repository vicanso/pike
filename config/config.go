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

package config

import (
	"errors"
	"strings"

	"github.com/vicanso/pike/app"
	"github.com/vicanso/pike/log"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

type (
	// Client client interface
	Client interface {
		// Get get the data of key
		Get() (data []byte, err error)
		// Set set the data of key
		Set(data []byte) (err error)
		// Watch watch change
		Watch(OnChange)
		// Close close client
		Close() error
	}
	OnChange func()

	// PikeConfig pike config
	PikeConfig struct {
		// YAML 界面展示之用，不需要保存
		YAML string `json:"yaml,omitempty" yaml:"-"`
		// Version 程序版本
		Version    string           `json:"version,omitempty" yaml:"version,omitempty" `
		Admin      AdminConfig      `json:"admin,omitempty" yaml:"admin,omitempty" validate:"omitempty,dive"`
		Compresses []CompressConfig `json:"compresses,omitempty" yaml:"compresses,omitempty" validate:"omitempty,dive"`
		Caches     []CacheConfig    `json:"caches,omitempty" yaml:"caches,omitempty" validate:"omitempty,dive"`
		Upstreams  []UpstreamConfig `json:"upstreams,omitempty" yaml:"upstreams,omitempty" validate:"omitempty,dive"`
		Locations  []LocationConfig `json:"locations,omitempty" yaml:"locations,omitempty" validate:"omitempty,dive"`
		Servers    []ServerConfig   `json:"servers,omitempty" yaml:"servers,omitempty" validate:"omitempty,dive"`
	}
	// AdminConfig admin config
	AdminConfig struct {
		User     string `json:"user,omitempty" yaml:"user,omitempty" validate:"omitempty,min=3"`
		Password string `json:"password,omitempty" yaml:"password,omitempty" validate:"omitempty,min=6"`
		Remark   string `json:"remark,omitempty" yaml:"remark,omitempty"`
	}
	// CompressConfig compress config
	CompressConfig struct {
		Name   string          `json:"name,omitempty" yaml:"name,omitempty" validate:"required,xName"`
		Levels map[string]uint `json:"levels,omitempty" yaml:"levels,omitempty"`
		Remark string          `json:"remark,omitempty" yaml:"remark,omitempty"`
	}
	// CacheConfig cache config
	CacheConfig struct {
		Name       string `json:"name,omitempty" yaml:"name,omitempty" validate:"required,xName"`
		Size       int    `json:"size,omitempty" yaml:"size,omitempty" validate:"required,gt=0" `
		HitForPass string `json:"hitForPass,omitempty" yaml:"hitForPass,omitempty" validate:"required,xDuration"`
		Remark     string `json:"remark,omitempty" yaml:"remark,omitempty"`
	}
	// UpstreamServerConfig upstream server config
	UpstreamServerConfig struct {
		Addr   string `json:"addr,omitempty" yaml:"addr,omitempty" validate:"required,xAddr"`
		Backup bool   `json:"backup,omitempty" yaml:"backup,omitempty"`
		// Healthy 界面展示使用，不需要保存
		Healthy bool `json:"healthy,omitempty" yaml:"-"`
	}
	// UpstreamConfig upstream config
	UpstreamConfig struct {
		Name           string                 `json:"name,omitempty" yaml:"name,omitempty" validate:"required,xName"`
		HealthCheck    string                 `json:"healthCheck,omitempty" yaml:"healthCheck,omitempty" validate:"omitempty,xURLPath"`
		Policy         string                 `json:"policy,omitempty" yaml:"policy,omitempty" validate:"omitempty,xPolicy"`
		EnableH2C      bool                   `json:"enableH2C,omitempty" yaml:"enableH2C,omitempty"`
		AcceptEncoding string                 `json:"acceptEncoding,omitempty" yaml:"acceptEncoding,omitempty" validate:"omitempty,ascii"`
		Servers        []UpstreamServerConfig `json:"servers,omitempty" yaml:"servers,omitempty" validate:"required,dive"`
		Remark         string                 `json:"remark,omitempty" yaml:"remark,omitempty"`
	}
	// LocationConfig location config
	LocationConfig struct {
		Name         string   `json:"name,omitempty" yaml:"name,omitempty" validate:"required,xName"`
		Upstream     string   `json:"upstream,omitempty" yaml:"upstream,omitempty" validate:"required,xName"`
		Prefixes     []string `json:"prefixes,omitempty" yaml:"prefixes,omitempty" validate:"omitempty,dive,xURLPath"`
		Rewrites     []string `json:"rewrites,omitempty" yaml:"rewrites,omitempty" validate:"omitempty,dive,xDivide"`
		QueryStrings []string `json:"queryStrings,omitempty" yaml:"queryStrings,omitempty" validate:"omitempty,dive,xDivide"`
		RespHeaders  []string `json:"respHeaders,omitempty" yaml:"respHeaders,omitempty" validate:"omitempty,dive,xDivide"`
		ReqHeaders   []string `json:"reqHeaders,omitempty" yaml:"reqHeaders,omitempty" validate:"omitempty,dive,xDivide"`
		Hosts        []string `json:"hosts,omitempty" yaml:"hosts,omitempty" validate:"omitempty,dive,hostname"`
		ProxyTimeout string   `json:"proxyTimeout,omitempty" yaml:"proxyTimeout,omitempty" validate:"omitempty,xDuration"`
		Remark       string   `json:"remark,omitempty" yaml:"remark,omitempty"`
	}
	// ServerConfig server config
	ServerConfig struct {
		LogFormat string   `json:"logFormat,omitempty" yaml:"logFormat,omitempty"`
		Addr      string   `json:"addr,omitempty" yaml:"addr,omitempty" validate:"required,ascii"`
		Locations []string `json:"locations,omitempty" yaml:"locations,omitempty" validate:"required,dive,xName"`
		Cache     string   `json:"cache,omitempty" yaml:"cache,omitempty" validate:"required,xName"`
		Compress  string   `json:"compress,omitempty" yaml:"compress,omitempty" validate:"omitempty"`
		// 最小压缩长度
		CompressMinLength string `json:"compressMinLength,omitempty" yaml:"compressMinLength,omitempty" validate:"omitempty,xSize"`
		// 压缩数据类型
		CompressContentTypeFilter string `json:"compressContentTypeFilter,omitempty" yaml:"compressContentTypeFilter,omitempty" validate:"omitempty,xFilter"`
		Remark                    string `json:"remark,omitempty" yaml:"remark,omitempty"`
	}
)

var defaultClient Client

var (
	ErrUpstreamNotFound = errors.New("upstream of location not found")
	ErrLocationNotFound = errors.New("location of server not found")
	ErrCacheNotFound    = errors.New("cache of server not found")
	ErrCompressNotFound = errors.New("compress of server not found")
)

// InitDefaultClient init default client
func InitDefaultClient(url string) (err error) {
	if defaultClient != nil {
		// 如果关闭出错，仅输出日志
		e := defaultClient.Close()
		if e != nil {
			log.Default().Error("close config client fail",
				zap.Error(e),
			)
		}
	}
	defaultClient = nil
	if strings.HasPrefix(url, "etcd://") {
		c, err := NewEtcdClient(url)
		if err != nil {
			return err
		}
		defaultClient = c
		return nil
	}

	c, err := NewFileClient(url)
	if err != nil {
		return
	}
	defaultClient = c
	return
}

func (c *PikeConfig) Validate() error {
	err := defaultValidator.Struct(c)
	if err != nil {
		return err
	}
	// 判断location中设置的upstream是否存在
	for _, l := range c.Locations {
		found := false
		for _, upstream := range c.Upstreams {
			if l.Upstream == upstream.Name {
				found = true
			}
		}
		if !found {
			return ErrUpstreamNotFound
		}
	}
	// 校验server中的location, cache 以及 compress 是否正确设置
	for _, s := range c.Servers {
		for _, item := range s.Locations {
			notFound := true
			for _, l := range c.Locations {
				if item == l.Name {
					notFound = false
				}
			}
			if notFound {
				return ErrLocationNotFound
			}
		}

		foundCache := s.Cache == ""
		for _, cacheConfig := range c.Caches {
			if cacheConfig.Name == s.Cache {
				foundCache = true
			}
		}
		if !foundCache {
			return ErrCacheNotFound
		}

		foundCompress := s.Compress == ""
		for _, compressConfig := range c.Compresses {
			if compressConfig.Name == s.Compress {
				foundCompress = true
			}
		}
		if !foundCompress {
			return ErrCompressNotFound
		}
	}

	return nil
}

// GetAdminConfig get admin config
func (p *PikeConfig) GetAdminConfig() AdminConfig {
	return p.Admin
}

// Read read pike config
func Read() (config *PikeConfig, err error) {
	data, err := defaultClient.Get()
	if err != nil {
		return
	}
	config = &PikeConfig{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return
	}
	config.YAML = string(data)
	return
}

// Write write pike config
func Write(config *PikeConfig) (err error) {
	err = config.Validate()
	if err != nil {
		return
	}
	config.Version = app.GetVersion()
	data, err := yaml.Marshal(config)
	if err != nil {
		return
	}
	return defaultClient.Set(data)
}

// Close close the client
func Close() (err error) {
	if defaultClient != nil {
		err = defaultClient.Close()
	}
	defaultClient = nil
	return
}

// Watch watch the change
func Watch(onChange OnChange) {
	defaultClient.Watch(onChange)
}
