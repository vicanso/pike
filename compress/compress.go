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

package compress

import (
	"compress/gzip"
	"errors"
	"sync"

	"github.com/andybalholm/brotli"
	"github.com/vicanso/pike/config"
	"go.uber.org/atomic"
)

const (
	EncodingGzip   = "gzip"
	EncodingBrotli = "br"
	EncodingLZ4    = "lz4"
	EncodingSnappy = "snz"
	EncodingZSTD   = "zst"
)

type (
	compressSrv struct {
		levels map[string]*atomic.Int32
	}
	CompressOption struct {
		Name   string
		Levels map[string]int
	}
	compressSrvs struct {
		m *sync.Map
	}
)

const BestCompression = "bestCompression"

var defaultCompressSrvList = NewServices([]CompressOption{
	{
		Name: BestCompression,
		Levels: map[string]int{
			// -1则会选择默认的压缩级别
			"br":   -1,
			"gzip": gzip.BestCompression,
		},
	},
})
var defaultCompressSrv = NewService()
var notSupportedEncoding = errors.New("not supported encoding")

// NewServices new compress services
func NewServices(opts []CompressOption) *compressSrvs {
	cs := &compressSrvs{
		m: &sync.Map{},
	}
	for _, opt := range opts {
		srv := NewService()
		srv.SetLevels(opt.Levels)
		cs.m.Store(opt.Name, srv)
	}
	return cs
}

// NewService new compress service
func NewService() *compressSrv {
	// 配置压缩级别，只设置了gzip与br
	levels := map[string]*atomic.Int32{
		EncodingGzip:   atomic.NewInt32(gzip.DefaultCompression),
		EncodingBrotli: atomic.NewInt32(brotli.DefaultCompression),
	}
	return &compressSrv{
		levels: levels,
	}
}

// Get get service by name
func (cs *compressSrvs) Get(name string) *compressSrv {
	value, ok := cs.m.Load(name)
	if !ok {
		return defaultCompressSrv
	}
	srv, ok := value.(*compressSrv)
	if !ok {
		return defaultCompressSrv
	}
	return srv
}

// Reset reset the services
func (cs *compressSrvs) Reset(opts []CompressOption) {
	// 此处不删除存在的压缩服务，因为compress实例并不占多少内存
	// 也避免配置了bestCompression后删除
	for _, opt := range opts {
		srv := NewService()
		srv.SetLevels(opt.Levels)
		cs.m.Store(opt.Name, srv)
	}
}

func convertConfigs(configs []config.CompressConfig) []CompressOption {
	opts := make([]CompressOption, 0)
	for _, item := range configs {
		levels := make(map[string]int)
		for key, value := range item.Levels {
			levels[key] = int(value)
		}
		opts = append(opts, CompressOption{
			Name:   item.Name,
			Levels: levels,
		})
	}
	return opts
}

// Reset reset default compress services
func Reset(configs []config.CompressConfig) {
	defaultCompressSrvList.Reset(convertConfigs(configs))
}

// Get get default compress service
func Get(name string) *compressSrv {
	return defaultCompressSrvList.Get(name)
}

// GetLevel get compress level
func (srv *compressSrv) GetLevel(encoding string) int {
	levelValue, ok := srv.levels[encoding]
	if !ok {
		return 0
	}
	return int(levelValue.Load())
}

// SetLevels set compres levels
func (srv *compressSrv) SetLevels(levels map[string]int) {
	for name, value := range levels {
		levelValue, ok := srv.levels[name]
		if ok {
			levelValue.Store(int32(value))
		}
	}
}

// Decompress decompress data
func (srv *compressSrv) Decompress(encoding string, data []byte) ([]byte, error) {
	switch encoding {
	case EncodingGzip:
		return srv.Gunzip(data)
	case EncodingBrotli:
		return srv.BrotliDecode(data)
	case EncodingLZ4:
		return srv.LZ4Decode(data)
	case EncodingSnappy:
		return srv.SnappyDecode(data)
	case EncodingZSTD:
		return srv.ZSTDDecode(data)
	case "":
		return data, nil
	}
	return nil, notSupportedEncoding
}

// Gzip compress data by gzip
func (srv *compressSrv) Gzip(data []byte) ([]byte, error) {
	level := srv.GetLevel(EncodingGzip)
	return doGzip(data, level)
}

// Gunzip decompress data by gzip
func (srv *compressSrv) Gunzip(data []byte) ([]byte, error) {
	return doGunzip(data)
}

// Brotli compress data by br
func (srv *compressSrv) Brotli(data []byte) ([]byte, error) {
	level := srv.GetLevel(EncodingBrotli)
	return doBrotli(data, level)
}

// BrotliDecode decompress data by brotli
func (srv *compressSrv) BrotliDecode(data []byte) ([]byte, error) {
	return doBrotliDecode(data)
}

// LZ4Decode decompress data by lz4
func (srv *compressSrv) LZ4Decode(data []byte) ([]byte, error) {
	return doLZ4Decode(data)
}

// SnappyDecode decompress data by snappy
func (srv *compressSrv) SnappyDecode(data []byte) ([]byte, error) {
	return doSnappyDecode(data)
}

// ZSTDDecode decompress data by zstd
func (srv *compressSrv) ZSTDDecode(data []byte) ([]byte, error) {
	return doZSTDDecode(data)
}
