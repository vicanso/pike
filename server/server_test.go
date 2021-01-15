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
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vicanso/elton"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
)

func TestGetSetCacheStatus(t *testing.T) {
	assert := assert.New(t)
	c := elton.NewContext(nil, nil)
	assert.Equal(cache.StatusUnknown, getCacheStatus(c))
	setCacheStatus(c, cache.StatusHit)
	assert.Equal(cache.StatusHit, getCacheStatus(c))
}

func TestGetSetHTTPResp(t *testing.T) {
	assert := assert.New(t)
	c := elton.NewContext(nil, nil)
	assert.Nil(getHTTPResp(c))
	httpResp := &cache.HTTPResponse{}
	setHTTPResp(c, httpResp)
	assert.Equal(httpResp, getHTTPResp(c))
}

func TestGetSetHTTPRespAge(t *testing.T) {
	assert := assert.New(t)
	c := elton.NewContext(nil, nil)
	assert.Equal(0, getHTTPRespAge(c))
	age := 10
	setHTTPRespAge(c, age)
	assert.Equal(age, getHTTPRespAge(c))
}

func TestGetSetHTTPCacheMaxAge(t *testing.T) {
	assert := assert.New(t)
	c := elton.NewContext(nil, nil)
	assert.Equal(0, getHTTPCacheMaxAge(c))
	age := 10
	setHTTPCacheMaxAge(c, age)
	assert.Equal(10, getHTTPCacheMaxAge(c))
}

func TestServer(t *testing.T) {
	assert := assert.New(t)

	locations := []string{
		"location-test",
	}
	cache := "cache-test"
	compress := "compress-test"
	filter := regexp.MustCompile(`text`)
	s := NewServer(ServerOption{
		Locations:                 locations,
		Cache:                     cache,
		Compress:                  compress,
		CompressContentTypeFilter: filter,
	})
	defer s.Close()

	assert.Equal(cache, s.GetCache())
	assert.Equal(locations, s.GetLocations())
	compressSrv, compressMinLength, compressContentTypeFilter := s.GetCompress()
	assert.Equal(compress, compressSrv)
	assert.Equal(defaultCompressMinLength, compressMinLength)
	assert.Equal(filter, compressContentTypeFilter)

	err := s.Start(true)
	assert.True(s.listening)
	assert.Nil(err)
	assert.NotEmpty(s.GetListenAddr())

	minLength := 101
	s.Update(ServerOption{
		CompressMinLength: minLength,
	})
	_, compressMinLength, _ = s.GetCompress()

	assert.Equal(minLength, compressMinLength)
}

func TestServers(t *testing.T) {
	assert := assert.New(t)
	locations := []string{
		"location-test",
	}
	cache := "cache-test"
	compress := "compress-test"
	filter := regexp.MustCompile(`text`)
	ss := NewServers([]ServerOption{
		{
			Locations:                 locations,
			Cache:                     cache,
			Compress:                  compress,
			CompressContentTypeFilter: filter,
		},
	})
	err := ss.Start()
	assert.Nil(err)
	defer ss.Close()
	s := ss.Get("")
	assert.NotNil(s)
	compresName, _, _ := s.GetCompress()
	assert.Equal(compress, compresName)

	newCompress := "compress-new"
	ss.Reset([]ServerOption{
		{
			Locations:                 locations,
			Cache:                     cache,
			Compress:                  newCompress,
			CompressContentTypeFilter: filter,
		},
	})
	compresName, _, _ = s.GetCompress()
	assert.Equal(newCompress, compresName)
}

func TestConvertConfig(t *testing.T) {
	assert := assert.New(t)
	addr := ":3015"
	locations := []string{
		"location-test",
	}
	cache := "cache-test"
	compress := "compress-test"
	minLength := 1000
	filter := `text|json`
	configs := []config.ServerConfig{
		{
			Addr:                      addr,
			Locations:                 locations,
			Cache:                     cache,
			Compress:                  compress,
			CompressMinLength:         "1kb",
			CompressContentTypeFilter: filter,
		},
	}
	opts := convertConfig(configs)
	assert.Equal(1, len(opts))
	assert.Equal(addr, opts[0].Addr)
	assert.Equal(locations, opts[0].Locations)
	assert.Equal(cache, opts[0].Cache)
	assert.Equal(compress, opts[0].Compress)
	assert.Equal(minLength, opts[0].CompressMinLength)
	assert.Equal(filter, opts[0].CompressContentTypeFilter.String())
}

func TestDefaultServers(t *testing.T) {
	assert := assert.New(t)

	defer Close()

	assert.Nil(Get(""))

	locations := []string{
		"location-test",
	}
	cache := "cache-test"
	compress := "compress-test"
	Reset([]config.ServerConfig{
		{
			Locations:                 locations,
			Cache:                     cache,
			Compress:                  compress,
			CompressContentTypeFilter: "text",
		},
	})
	err := Start()
	assert.Nil(err)
	s := Get("")
	assert.NotNil(s)
}
