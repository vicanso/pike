package customMiddleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vicanso/pike/cache"

	"github.com/labstack/echo"
	"github.com/vicanso/pike/vars"
)

func TestIdentifier(t *testing.T) {
	client := &cache.Client{
		Path: "/tmp/test.cache",
	}
	err := client.Init()
	if err != nil {
		t.Fatalf("cache init fail, %v", err)
	}
	defer client.Close()

	t.Run("pass", func(t *testing.T) {
		fn := Identifier(client)(func(c echo.Context) error {
			status := c.Get(vars.Status).(int)
			if status != cache.Pass {
				t.Fatalf("the status of post request should be pass")
			}
			return nil
		})
		e := echo.New()
		req := httptest.NewRequest(echo.POST, "/users/me", nil)
		c := e.NewContext(req, nil)
		fn(c)
	})

	t.Run("fetching", func(t *testing.T) {
		fn := Identifier(client)(func(c echo.Context) error {
			status := c.Get(vars.Status).(int)
			if status != cache.Fetching {
				t.Fatalf("the status of the first get request should be fetching")
			}
			return nil
		})
		e := echo.New()
		req := httptest.NewRequest(echo.GET, "/users/me", nil)
		c := e.NewContext(req, nil)
		fn(c)

	})

	t.Run("hit for pass", func(t *testing.T) {
		fn := Identifier(client)(func(c echo.Context) error {
			status := c.Get(vars.Status).(int)
			if status != cache.HitForPass {
				t.Fatalf("the status of the request should be hit for pass")
			}
			return nil
		})
		e := echo.New()
		req := httptest.NewRequest(echo.GET, "/users/me", nil)
		c := e.NewContext(req, nil)
		go func() {
			// 延时执行
			time.Sleep(10 * time.Millisecond)
			client.HitForPass([]byte("GET example.com /users/me"), 600)
		}()
		fn(c)

	})
}
