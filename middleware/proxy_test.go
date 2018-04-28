package custommiddleware

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mitchellh/go-server-timing"

	"github.com/vicanso/pike/cache"

	"github.com/h2non/gock"

	"github.com/vicanso/pike/vars"

	"github.com/labstack/echo"
	"github.com/vicanso/pike/proxy"
)

type (
	closeNotifyRecorder struct {
		*httptest.ResponseRecorder
		closed chan bool
	}
)

func newCloseNotifyRecorder() *closeNotifyRecorder {
	return &closeNotifyRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
}

func (c *closeNotifyRecorder) close() {
	c.closed <- true
}

func (c *closeNotifyRecorder) CloseNotify() <-chan bool {
	return c.closed
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
		fn := Proxy(ProxyConfig{})(func(c echo.Context) error {
			if c.Get(vars.Response).(*cache.Response) != resp {
				t.Fatalf("proxy with cache fail")
			}
			return nil
		})
		e := echo.New()
		c := e.NewContext(nil, nil)
		c.Set(vars.Response, resp)
		fn(c)
	})

	t.Run("proxy", func(t *testing.T) {
		fn := Proxy(ProxyConfig{
			Rewrite: map[string]string{
				"/api/*": "/$1",
			},
		})(func(c echo.Context) error {
			resp := c.Get(vars.Response).(*cache.Response)
			if strings.TrimSpace(string(resp.Body)) != `{"name":"tree.xie"}` {
				t.Fatalf("get response from proxy fail")
			}
			return nil
		})
		e := echo.New()
		req := httptest.NewRequest(echo.GET, "http://aslant.site/api/users/me", nil)
		req.Header.Set(echo.HeaderIfModifiedSince, "Mon, 07 Nov 2016 07:51:11 GMT")
		req.Header.Set(vars.IfNoneMatch, `"16e36-540b1498e39c0"`)
		res := newCloseNotifyRecorder()
		c := e.NewContext(req, res)
		timing := &servertiming.Header{}
		c.Set(vars.Timing, timing)
		aslant := "aslant"
		backend := "http://127.0.0.1:5001"
		d := &proxy.Director{
			Name: aslant,
		}
		err := fn(c)
		if err != vars.ErrDirectorNotFound {
			t.Fatalf("should return director not found")
		}

		c.Set(vars.Director, d)
		d.Hosts = []string{
			"(www.)?aslant.site",
		}

		err = fn(c)
		if err != vars.ErrNoBackendAvaliable {
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
		fn(c)
	})

	t.Run("proxy response gzip", func(t *testing.T) {
		fn := Proxy(ProxyConfig{})(func(c echo.Context) error {
			resp := c.Get(vars.Response).(*cache.Response)
			if strings.TrimSpace(string(resp.GzipBody)) != `{"name":"tree.xie"}` {
				t.Fatalf("get gzip response from proxy fail")
			}
			return nil
		})
		e := echo.New()
		req := httptest.NewRequest(echo.GET, "http://aslant.site/api/users/me", nil)
		res := newCloseNotifyRecorder()
		c := e.NewContext(req, res)
		aslant := "aslant"
		backend := "http://127.0.0.1:5001"
		d := &proxy.Director{
			Name: aslant,
		}

		c.Set(vars.Director, d)

		gock.New(backend).
			Get("/users/me").
			Reply(200).
			SetHeader("Cache-Control", "max-age=10").
			SetHeader("Content-Encoding", "gzip").
			JSON(map[string]string{
				"name": "tree.xie",
			})
		d.AddAvailableBackend(backend)
		fn(c)
	})

	t.Run("proxy response br", func(t *testing.T) {
		fn := Proxy(ProxyConfig{})(func(c echo.Context) error {
			resp := c.Get(vars.Response).(*cache.Response)
			if strings.TrimSpace(string(resp.BrBody)) != `{"name":"tree.xie"}` {
				t.Fatalf("get gzip response from proxy fail")
			}
			return nil
		})
		e := echo.New()
		req := httptest.NewRequest(echo.GET, "http://aslant.site/api/users/me", nil)
		res := newCloseNotifyRecorder()
		c := e.NewContext(req, res)
		aslant := "aslant"
		backend := "http://127.0.0.1:5001"
		d := &proxy.Director{
			Name: aslant,
		}

		c.Set(vars.Director, d)

		gock.New(backend).
			Get("/users/me").
			Reply(200).
			SetHeader("Cache-Control", "max-age=10").
			SetHeader("Content-Encoding", "br").
			JSON(map[string]string{
				"name": "tree.xie",
			})
		d.AddAvailableBackend(backend)
		fn(c)
	})

	t.Run("proxy response unsupport encoding", func(t *testing.T) {
		fn := Proxy(ProxyConfig{})(func(c echo.Context) error {
			return nil
		})
		e := echo.New()
		req := httptest.NewRequest(echo.GET, "http://aslant.site/api/users/me", nil)
		res := newCloseNotifyRecorder()
		c := e.NewContext(req, res)
		aslant := "aslant"
		backend := "http://127.0.0.1:5001"
		d := &proxy.Director{
			Name: aslant,
		}

		c.Set(vars.Director, d)

		gock.New(backend).
			Get("/users/me").
			Reply(200).
			SetHeader("Cache-Control", "max-age=10").
			SetHeader("Content-Encoding", "unknown encoding").
			JSON(map[string]string{
				"name": "tree.xie",
			})
		d.AddAvailableBackend(backend)
		err := fn(c)
		if err != vars.ErrContentEncodingNotSupport {
			t.Fatalf("not support encoding should return error")
		}
	})
}
