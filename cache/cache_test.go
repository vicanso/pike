package cache

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vicanso/cod"
)

func TestGetHTTPCache(t *testing.T) {
	assert := assert.New(t)
	dsp := NewDispatcher(Options{})
	key := []byte("abc")
	hc := dsp.GetHTTPCache(key)
	assert.NotNil(hc, "get http cache should not be nil")
}

func TestDispatcher(t *testing.T) {
	dsp := NewDispatcher(Options{})
	t.Run("get status(hit for pass)", func(t *testing.T) {
		assert := assert.New(t)
		k1 := []byte("abc")
		k2 := []byte("abc")
		k3 := []byte("abc")
		hc1 := dsp.GetHTTPCache(k1)
		go func() {
			time.Sleep(time.Millisecond)
			hc1.HitForPass()
			hc1.Done()
		}()
		hc2 := dsp.GetHTTPCache(k2)
		defer hc2.Done()
		hc3 := dsp.GetHTTPCache(k3)
		defer hc3.Done()
		assert.Equal(hc2.Status, HitForPass, "status should be hit for pass")
		assert.Equal(hc3.Status, HitForPass, "status should be hit for pass")
		assert.Equal(hc2, hc3, "these two caches should be same")
	})

	t.Run("get status(cacheable)", func(t *testing.T) {
		assert := assert.New(t)
		k1 := []byte("def")
		k2 := []byte("def")
		k3 := []byte("def")
		hc1 := dsp.GetHTTPCache(k1)
		go func() {
			time.Sleep(time.Millisecond)
			hc1.Cacheable(10, &cod.Context{
				Headers:    make(http.Header),
				BodyBuffer: bytes.NewBufferString("abcd"),
			})
			hc1.Done()
		}()
		hc2 := dsp.GetHTTPCache(k2)
		defer hc2.Done()
		hc3 := dsp.GetHTTPCache(k3)
		defer hc3.Done()
		assert.Equal(hc2.Status, Cacheable, "status should be cacheable")
		assert.Equal(hc3.Status, Cacheable, "status should be cacheable")
		assert.Equal(hc2, hc3, "these two caches should be same")
	})

	t.Run("get all cache status", func(t *testing.T) {
		cacheList := dsp.GetCacheList()
		assert.NotEqual(t, len(cacheList), 0)
	})
}

func TestHitForPass(t *testing.T) {
	assert := assert.New(t)
	opts := Options{}
	hc := HTTPCache{
		opts: &opts,
	}
	hc.HitForPass()
	assert.NotEqual(hc.CreatedAt, 0, "createdAt shouldn't be 0")
	assert.NotEqual(hc.ExpiredAt, 0, "expiredAt shouldn't be 0")
	assert.Equal(hc.Status, HitForPass, "status should be hit for pass")
}

func TestCacheable(t *testing.T) {
	opts := Options{
		Size:              10,
		ZoneSize:          10,
		TextFilter:        regexp.MustCompile("text|javascript|json"),
		CompressMinLength: 1024,
	}
	t.Run("cacheable", func(t *testing.T) {
		assert := assert.New(t)
		header := make(http.Header)
		header.Set(cod.HeaderContentType, "text/html")
		header.Set(cod.HeaderContentLength, "10")
		buf := make([]byte, 4096)
		maxAge := 10
		c := &cod.Context{
			Headers:    header,
			BodyBuffer: bytes.NewBuffer(buf),
			StatusCode: 200,
			Request:    httptest.NewRequest("GET", "/", nil),
		}
		hc := &HTTPCache{
			opts: &opts,
		}
		hc.Cacheable(maxAge, c)
		assert.Equal(hc.Status, Cacheable, "status should be cacheable")
		assert.NotNil(hc.Headers, "headers shouldn't be nil")
		assert.Nil(hc.Body, "original body should be nil")
		assert.NotNil(hc.GzipBody, "gzip body shouldn't be nil")
		assert.Equal(hc.Headers.Get(cod.HeaderContentLength), "", "Content-Length should be empty")
	})

	t.Run("cacheable(gunzip success)", func(t *testing.T) {
		assert := assert.New(t)
		header := make(http.Header)
		header.Set(cod.HeaderContentType, "text/html")
		header.Set(cod.HeaderContentEncoding, "gzip")
		data := []byte("abcd")
		buf, _ := doGzip(data, 0)
		maxAge := 10
		c := &cod.Context{
			Headers:    header,
			BodyBuffer: bytes.NewBuffer(buf),
			StatusCode: 200,
		}
		hc := &HTTPCache{
			opts: &opts,
		}
		hc.Cacheable(maxAge, c)
		assert.Equal(hc.Status, Cacheable, "status should be cacheable")
		assert.Equal(hc.Body.Bytes(), data, "body should be the smae as original")
		assert.Nil(hc.GzipBody, "gzip data should be nil")
	})

	t.Run("gunzip fail should hit for pass", func(t *testing.T) {
		assert := assert.New(t)
		header := make(http.Header)
		header.Set(cod.HeaderContentType, "text/html")
		header.Set(cod.HeaderContentEncoding, "gzip")
		buf := make([]byte, 4096)
		maxAge := 10
		c := &cod.Context{
			Headers:    header,
			BodyBuffer: bytes.NewBuffer(buf),
			StatusCode: 200,
		}
		hc := &HTTPCache{
			opts: &opts,
		}
		hc.Cacheable(maxAge, c)
		assert.Equal(hc.Status, HitForPass, "status should be hit for pass")
	})

	t.Run("not support encoding should hit for pass", func(t *testing.T) {
		assert := assert.New(t)
		header := make(http.Header)
		header.Set(cod.HeaderContentType, "text/html")
		header.Set(cod.HeaderContentEncoding, "xx")
		buf := make([]byte, 4096)
		maxAge := 10
		c := &cod.Context{
			Headers:    header,
			BodyBuffer: bytes.NewBuffer(buf),
			StatusCode: 200,
		}
		hc := &HTTPCache{
			opts: &opts,
		}
		hc.Cacheable(maxAge, c)
		assert.Equal(hc.Status, HitForPass, "status should be hit for pass")
	})
}
