package custommiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vicanso/pike/cache"

	"github.com/vicanso/pike/vars"

	"github.com/labstack/echo"
)

func TestFreshChecker(t *testing.T) {
	freshCheckerConfig := FreshCheckerConfig{}
	t.Run("no response", func(t *testing.T) {
		fn := FreshChecker(freshCheckerConfig)(func(c echo.Context) error {
			return nil
		})
		e := echo.New()
		c := e.NewContext(nil, nil)
		err := fn(c)
		if err != vars.ErrResponseNotSet {
			t.Fatalf("no response should return error")
		}
	})

	t.Run("post request", func(t *testing.T) {
		fn := FreshChecker(freshCheckerConfig)(func(c echo.Context) error {
			fresh := c.Get(vars.Fresh).(bool)
			if fresh {
				t.Fatalf("post request will not be fresh")
			}
			return nil
		})
		e := echo.New()
		req := httptest.NewRequest(echo.POST, "/users/me", nil)
		c := e.NewContext(req, nil)
		c.Set(vars.Response, &cache.Response{
			StatusCode: http.StatusOK,
		})
		err := fn(c)
		if err != nil {
			t.Fatalf("check post request fail, %v", err)
		}
	})

	t.Run("get request(502)", func(t *testing.T) {
		fn := FreshChecker(freshCheckerConfig)(func(c echo.Context) error {
			fresh := c.Get(vars.Fresh).(bool)
			if fresh {
				t.Fatalf("get request(502) will not be fresh")
			}
			return nil
		})
		e := echo.New()
		req := httptest.NewRequest(echo.GET, "/users/me", nil)
		c := e.NewContext(req, nil)
		c.Set(vars.Response, &cache.Response{
			StatusCode: http.StatusBadGateway,
		})
		err := fn(c)
		if err != nil {
			t.Fatalf("check get request(502) fail, %v", err)
		}
	})

	t.Run("get reqeust(fresh)", func(t *testing.T) {
		fn := FreshChecker(freshCheckerConfig)(func(c echo.Context) error {
			fresh := c.Get(vars.Fresh).(bool)
			if !fresh {
				t.Fatalf("get reqeust(fresh) should be fresh")
			}
			return nil
		})
		e := echo.New()
		req := httptest.NewRequest(echo.GET, "/users/me", nil)
		resp := &httptest.ResponseRecorder{}
		c := e.NewContext(req, resp)
		c.Set(vars.Response, &cache.Response{
			StatusCode: http.StatusOK,
		})
		c.Request().Header.Set(vars.IfNoneMatch, "ABCD")
		c.Response().Header().Set(vars.ETag, "ABCD")
		err := fn(c)
		if err != nil {
			t.Fatalf("check get reqeust(fresh) fail, %v", err)
		}
	})
}
