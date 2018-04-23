package customMiddleware

import (
	"net/http/httptest"
	"testing"

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

func TestProxyWithConfig(t *testing.T) {
	t.Run("proxy", func(t *testing.T) {
		fn := ProxyWithConfig(ProxyConfig{
			Rewrite: map[string]string{
				"/api/*": "/$1",
			},
		})(func(c echo.Context) error {
			body := c.Get(vars.Body).([]byte)
			if string(body) != "{\"name\":\"tree.xie\"}\n" {
				t.Fatalf("proxy fail")
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
			Hosts: []string{
				"(www.)?aslant.site",
			},
		}
		gock.New(backend).
			Get("/users/me").
			Reply(200).
			JSON(map[string]string{
				"name": "tree.xie",
			})
		d.AddAvailableBackend(backend)
		c.Set(vars.Director, d)
		fn(c)
	})
}
