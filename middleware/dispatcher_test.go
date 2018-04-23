package customMiddleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vicanso/pike/vars"

	"github.com/labstack/echo"
	"github.com/vicanso/pike/cache"
)

func TestDispatcher(t *testing.T) {
	client := &cache.Client{
		Path: "/tmp/test.cache",
	}
	err := client.Init()
	if err != nil {
		t.Fatalf("cache init fail, %v", err)
	}
	defer client.Close()
	t.Run("get cache age", func(t *testing.T) {
		if getCacheAge([]byte("")) != 0 {
			t.Fatalf("no cache-control header should be 0")
		}

		if getCacheAge([]byte("no-cache")) != 0 {
			t.Fatalf("no cache should be 0")
		}

		if getCacheAge([]byte("no-store")) != 0 {
			t.Fatalf("no store should be 0")
		}

		if getCacheAge([]byte("private")) != 0 {
			t.Fatalf("private cache should be 0")
		}

		if getCacheAge([]byte("max-age=10")) != 10 {
			t.Fatalf("get cache age from max-age fail")
		}

		if getCacheAge([]byte("max-age=10,s-maxage=1")) != 1 {
			t.Fatalf("get cache age from s-maxage fail")
		}
	})
	t.Run("dispatch response", func(t *testing.T) {
		fn := Dispatcher(client)(func(c echo.Context) error {
			return nil
		})
		e := echo.New()
		req := httptest.NewRequest(echo.POST, "/users/me", nil)
		resp := &httptest.ResponseRecorder{
			Body: new(bytes.Buffer),
		}
		c := e.NewContext(req, resp)
		c.Set(vars.Identity, []byte("abc"))
		c.Set(vars.Status, cache.Fetching)
		c.Set(vars.Code, 200)
		c.Set(vars.Body, []byte("ABCD"))
		c.Set(vars.Header, http.Header{
			"Token": []string{
				"A",
			},
		})
		fn(c)
		if resp.Code != 200 {
			t.Fatalf("the response code should be 200")
		}
		if resp.Header().Get("Token") != "A" {
			t.Fatalf("the response header of token should be A")
		}
		if string(resp.Body.Bytes()) != "ABCD" {
			t.Fatalf("the response body should be ABCD")
		}
	})
}
