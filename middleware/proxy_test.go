package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/h2non/gock"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/pike"
)

func TestGetCacheAge(t *testing.T) {
	t.Run("set cookie", func(t *testing.T) {
		header := make(http.Header)
		header["Cache-Control"] = []string{
			"public, max-age=60",
		}
		header["Set-Cookie"] = []string{
			"jt=abcd",
		}
		if getCacheAge(header) != 0 {
			t.Fatalf("max age of set cookie response should be 0")
		}
	})
	t.Run("no cache control", func(t *testing.T) {
		header := make(http.Header)
		if getCacheAge(header) != 0 {
			t.Fatalf("max age of no cache control response should be 0")
		}
	})
	t.Run("no cache", func(t *testing.T) {
		header := make(http.Header)
		header["Cache-Control"] = []string{
			"no-cache",
		}
		if getCacheAge(header) != 0 {
			t.Fatalf("max age of no cache response should be 0")
		}
	})
	t.Run("no store", func(t *testing.T) {
		header := make(http.Header)
		header["Cache-Control"] = []string{
			"no-store",
		}
		if getCacheAge(header) != 0 {
			t.Fatalf("max age of no store response should be 0")
		}
	})
	t.Run("private", func(t *testing.T) {
		header := make(http.Header)
		header["Cache-Control"] = []string{
			"private, max-age=60",
		}
		if getCacheAge(header) != 0 {
			t.Fatalf("max age of private response should be 0")
		}
	})
	t.Run("s-maxage", func(t *testing.T) {
		header := make(http.Header)
		header["Cache-Control"] = []string{
			"s-maxage=60, max-age=10",
		}
		if getCacheAge(header) != 60 {
			t.Fatalf("response cache should get from s-maxage")
		}
	})

	t.Run("max-age", func(t *testing.T) {
		header := make(http.Header)
		header["Cache-Control"] = []string{
			"max-age=10",
		}
		if getCacheAge(header) != 10 {
			t.Fatalf("response cache should get from max-age")
		}
	})
}

func TestGenETag(t *testing.T) {
	eTag := genETag([]byte(""))
	if eTag != "\"0-2jmj7l5rSw0yVb_vlWAYkK_YBwk=\"" {
		t.Fatalf("get empty data etag fail")
	}
	buf := []byte("测试使用的响应数据")
	eTag = genETag(buf)
	if eTag != "\"1b-gQEzXLxF7NjFZ-x0-GK1Pg8NBZA=\"" {
		t.Fatalf("get etag fail")
	}
}

func TestProxy(t *testing.T) {
	// 响应数据已从缓存中获取，next
	t.Run("proxy with cache", func(t *testing.T) {
		resp := &cache.Response{}
		fn := Proxy(ProxyConfig{})
		c := pike.NewContext(nil)
		c.Resp = resp
		err := fn(c, func() error {
			return nil
		})
		if err != nil {
			t.Fatalf("proxy with cache response fail, %v", err)
		}
	})

	t.Run("proxy", func(t *testing.T) {
		fn := Proxy(ProxyConfig{
			Rewrites: []string{
				"/api/*:/$1",
			},
		})
		req := httptest.NewRequest(http.MethodGet, "http://aslant.site/api/users/me", nil)
		req.Header.Set(pike.HeaderIfModifiedSince, "Mon, 07 Nov 2016 07:51:11 GMT")
		req.Header.Set(pike.HeaderIfNoneMatch, `"16e36-540b1498e39c0"`)
		c := pike.NewContext(req)
		aslant := "aslant"
		backend := "http://127.0.0.1:5001"
		d := &pike.Director{
			Name:         aslant,
			TargetURLMap: make(map[string]*url.URL),
		}
		err := fn(c, func() error {
			return nil
		})
		if err != ErrDirectorNotFound {
			t.Fatalf("should return director not found")
		}
		c.Director = d
		d.Hosts = []string{
			"(www.)?aslant.site",
		}
		err = fn(c, func() error {
			return nil
		})
		if err != ErrNoBackendAvaliable {
			t.Fatalf("should return no backend avaliable")
		}
		gock.New(backend).
			Get("/users/me").
			Reply(200).
			SetHeader("Cache-Control", "max-age=10").
			JSON(map[string]string{
				"name": "tree.xie",
			})
		d.AddAvailableBackend(backend)
		err = fn(c, func() error {
			return nil
		})
		if err != nil {
			t.Fatalf("proxy fail")
		}
		// 字符串最后有个换行符
		str := strings.Trim(string(c.Response.Bytes()), "\n")
		if str != `{"name":"tree.xie"}` {
			t.Fatalf("response is wrong")
		}
	})

	t.Run("director with rewrites", func(t *testing.T) {
		fn := Proxy(ProxyConfig{})
		req := httptest.NewRequest(http.MethodGet, "http://aslant.site/api/users/me", nil)
		req.Header.Set(pike.HeaderIfModifiedSince, "Mon, 07 Nov 2016 07:51:11 GMT")
		req.Header.Set(pike.HeaderIfNoneMatch, `"16e36-540b1498e39c0"`)
		c := pike.NewContext(req)
		aslant := "aslant"
		backend := "http://127.0.0.1:5001"
		d := &pike.Director{
			Name: aslant,
			Rewrites: []string{
				"/api/*:/_api/$1",
			},
			TargetURLMap: make(map[string]*url.URL),
		}
		d.GenRewriteRegexp()
		c.Director = d
		gock.New(backend).
			Get("/_api/users/me").
			Reply(200).
			SetHeader("Cache-Control", "max-age=10").
			JSON(map[string]string{
				"name": "tree.xie",
			})
		d.AddAvailableBackend(backend)
		err := fn(c, func() error {
			return nil
		})
		if err != nil {
			t.Fatalf("director with rewrites fail, %v", err)
		}
		str := strings.Trim(string(c.Response.Bytes()), "\n")
		if str != `{"name":"tree.xie"}` {
			t.Fatalf("response is wrong")
		}
	})

	t.Run("proxy response gzip", func(t *testing.T) {
		fn := Proxy(ProxyConfig{})
		req := httptest.NewRequest(http.MethodGet, "http://aslant.site/api/users/me", nil)
		c := pike.NewContext(req)
		aslant := "aslant"
		backend := "http://127.0.0.1:5001"
		d := &pike.Director{
			Name:         aslant,
			TargetURLMap: make(map[string]*url.URL),
		}
		c.Director = d

		gock.New(backend).
			Get("/users/me").
			Reply(200).
			SetHeader("Cache-Control", "max-age=10").
			SetHeader("Content-Encoding", "gzip").
			JSON(map[string]string{
				"name": "tree.xie",
			})
		d.AddAvailableBackend(backend)
		err := fn(c, func() error {
			return nil
		})
		if err != nil {
			t.Fatalf("proxy response gzip fail, %v", err)
		}
		str := strings.Trim(string(c.Response.Bytes()), "\n")
		if str != `{"name":"tree.xie"}` {
			t.Fatalf("response is wrong")
		}
	})

	t.Run("proxy response br", func(t *testing.T) {
		fn := Proxy(ProxyConfig{})
		req := httptest.NewRequest(http.MethodGet, "http://aslant.site/api/users/me", nil)
		c := pike.NewContext(req)
		aslant := "aslant"
		backend := "http://127.0.0.1:5001"
		d := &pike.Director{
			Name:         aslant,
			TargetURLMap: make(map[string]*url.URL),
		}
		c.Director = d

		gock.New(backend).
			Get("/users/me").
			Reply(200).
			SetHeader("Cache-Control", "max-age=10").
			SetHeader("Content-Encoding", "br").
			JSON(map[string]string{
				"name": "tree.xie",
			})
		d.AddAvailableBackend(backend)
		err := fn(c, func() error {
			return nil
		})
		if err != nil {
			t.Fatalf("proxy response br fail, %v", err)
		}
		str := strings.Trim(string(c.Response.Bytes()), "\n")
		if str != `{"name":"tree.xie"}` {
			t.Fatalf("response is wrong")
		}
	})

	t.Run("proxy response unsupport encoding", func(t *testing.T) {
		fn := Proxy(ProxyConfig{})
		req := httptest.NewRequest(http.MethodGet, "http://aslant.site/api/users/me", nil)
		c := pike.NewContext(req)
		aslant := "aslant"
		backend := "http://127.0.0.1:5001"
		d := &pike.Director{
			Name:         aslant,
			TargetURLMap: make(map[string]*url.URL),
		}
		c.Director = d

		gock.New(backend).
			Get("/users/me").
			Reply(200).
			SetHeader("Cache-Control", "max-age=10").
			SetHeader("Content-Encoding", "unknown encoding").
			JSON(map[string]string{
				"name": "tree.xie",
			})
		d.AddAvailableBackend(backend)
		err := fn(c, func() error {
			return nil
		})
		if err != ErrContentEncodingNotSupport {
			t.Fatalf("not support encoding should return error")
		}
	})
}
