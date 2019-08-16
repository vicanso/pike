package cache

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vicanso/elton"
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
		assert.Equal(HitForPass, hc2.Status, "status should be hit for pass")
		assert.Equal(HitForPass, hc3.Status, "status should be hit for pass")
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
			hc1.Cacheable(10, &elton.Context{
				Headers:    make(http.Header),
				BodyBuffer: bytes.NewBufferString("abcd"),
			})
			hc1.Done()
		}()
		hc2 := dsp.GetHTTPCache(k2)
		defer hc2.Done()
		hc3 := dsp.GetHTTPCache(k3)
		defer hc3.Done()
		assert.Equal(Cacheable, hc2.Status, "status should be cacheable")
		assert.Equal(Cacheable, hc3.Status, "status should be cacheable")
		assert.Equal(hc2, hc3, "these two caches should be same")
	})

	t.Run("expire cache", func(t *testing.T) {
		assert := assert.New(t)
		k := []byte("def")
		cache := dsp.getCache(k)

		key := byteSliceToString(k)
		lruCache := cache.lruCache
		v, ok := lruCache.Get(key)
		assert.True(ok)
		assert.NotEqual(int64(1), v.ExpiredAt)
		dsp.Expire(k)
		assert.Equal(int64(1), v.ExpiredAt)
	})

	t.Run("get all cache status", func(t *testing.T) {
		cacheList := dsp.GetCacheList()
		assert.NotEqual(t, 0, len(cacheList))
	})

	t.Run("different key create http cache", func(t *testing.T) {
		assert := assert.New(t)
		key1 := []byte("different1")
		key2 := []byte("different2")
		// 所有的缓存都命中同一个cache
		oneCacheDsp := NewDispatcher(Options{
			Size: 1,
		})

		hc1 := oneCacheDsp.GetHTTPCache(key1)
		go func() {
			time.Sleep(5 * time.Millisecond)
			// key1 有等待读的处理，此时key1可正常
			oneCacheDsp.GetHTTPCache(key2)
			hc1.HitForPass()
			hc1.Done()
		}()
		// 无法成功，等待中
		hc1Read := oneCacheDsp.GetHTTPCache(key1)
		assert.Equal(HitForPass, hc1Read.Status)
	})
}

func TestHitForPass(t *testing.T) {
	assert := assert.New(t)
	opts := Options{}
	hc := HTTPCache{
		opts: &opts,
	}
	hc.HitForPass()
	assert.NotEqual(0, hc.CreatedAt, "createdAt shouldn't be 0")
	assert.NotEqual(0, hc.ExpiredAt, "expiredAt shouldn't be 0")
	assert.Equal(HitForPass, hc.Status, "status should be hit for pass")
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
		header.Set(elton.HeaderContentType, "text/html")
		header.Set(elton.HeaderContentLength, "10")
		buf := make([]byte, 4096)
		maxAge := 10
		c := &elton.Context{
			Headers:    header,
			BodyBuffer: bytes.NewBuffer(buf),
			StatusCode: 200,
			Request:    httptest.NewRequest("GET", "/", nil),
		}
		hc := &HTTPCache{
			opts: &opts,
		}
		hc.Cacheable(maxAge, c)
		assert.Equal(Cacheable, hc.Status, "status should be cacheable")
		assert.NotNil(hc.Headers, "headers shouldn't be nil")
		assert.Nil(hc.Body, "original body should be nil")
		assert.NotNil(hc.GzipBody, "gzip body shouldn't be nil")
		assert.Equal("", hc.Headers.Get(elton.HeaderContentLength), "Content-Length should be empty")
	})

	t.Run("cacheable(gunzip success)", func(t *testing.T) {
		assert := assert.New(t)
		header := make(http.Header)
		header.Set(elton.HeaderContentType, "text/html")
		header.Set(elton.HeaderContentEncoding, "gzip")
		data := []byte("abcd")
		buf, _ := doGzip(data, 0)
		maxAge := 10
		c := &elton.Context{
			Headers:    header,
			BodyBuffer: bytes.NewBuffer(buf),
			StatusCode: 200,
		}
		hc := &HTTPCache{
			opts: &opts,
		}
		hc.Cacheable(maxAge, c)
		assert.Equal(Cacheable, hc.Status, "status should be cacheable")
		assert.Equal(data, hc.Body.Bytes(), "body should be the smae as original")
		assert.Nil(hc.GzipBody, "gzip data should be nil")
	})

	t.Run("gunzip fail should hit for pass", func(t *testing.T) {
		assert := assert.New(t)
		header := make(http.Header)
		header.Set(elton.HeaderContentType, "text/html")
		header.Set(elton.HeaderContentEncoding, "gzip")
		buf := make([]byte, 4096)
		maxAge := 10
		c := &elton.Context{
			Headers:    header,
			BodyBuffer: bytes.NewBuffer(buf),
			StatusCode: 200,
		}
		hc := &HTTPCache{
			opts: &opts,
		}
		hc.Cacheable(maxAge, c)
		assert.Equal(HitForPass, hc.Status, "status should be hit for pass")
	})

	t.Run("not support encoding should hit for pass", func(t *testing.T) {
		assert := assert.New(t)
		header := make(http.Header)
		header.Set(elton.HeaderContentType, "text/html")
		header.Set(elton.HeaderContentEncoding, "xx")
		buf := make([]byte, 4096)
		maxAge := 10
		c := &elton.Context{
			Headers:    header,
			BodyBuffer: bytes.NewBuffer(buf),
			StatusCode: 200,
		}
		hc := &HTTPCache{
			opts: &opts,
		}
		hc.Cacheable(maxAge, c)
		assert.Equal(HitForPass, hc.Status, "status should be hit for pass")
	})

	t.Run("response for accept encoding", func(t *testing.T) {
		assert := assert.New(t)
		rawData := bytes.NewBufferString("abcd")
		brBody := rawData
		gzipData, err := Gzip([]byte("abcd"))
		assert.Nil(err)
		gzipBody := bytes.NewBuffer(gzipData)
		hc := &HTTPCache{
			BrBody:   brBody,
			GzipBody: gzipBody,
		}

		buf, encoding, err := hc.Response("br")
		assert.Nil(err)
		assert.Equal("br", encoding, "should return br enconding")
		assert.Equal(brBody, buf, "should return br data")

		buf, encoding, err = hc.Response("gzip")
		assert.Nil(err)
		assert.Equal("gzip", encoding, "should return gzip enconding")
		assert.Equal(gzipBody, buf, "should return gzip data")

		buf, encoding, err = hc.Response("")
		assert.Nil(err)
		assert.Equal("", encoding, "should return empty enconding")
		assert.Equal(rawData, buf, "should return raw data")

		body := bytes.NewBufferString("a")
		hc.Body = body
		hc.GzipBody = nil
		buf, encoding, err = hc.Response("")
		assert.Nil(err)
		assert.Equal("", encoding, "should return empty enconding")
		assert.Equal(body, buf, "should return body data")
	})
}
