package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vicanso/pike/pike"

	"github.com/vicanso/pike/cache"
)

func TestFreshChecker(t *testing.T) {
	freshCheckerConfig := FreshCheckerConfig{}
	t.Run("no response", func(t *testing.T) {
		fn := FreshChecker(freshCheckerConfig)
		c := pike.NewContext(nil)
		err := fn(c, pike.NoopNext)
		if err != ErrResponseNotSet {
			t.Fatalf("no response should return error")
		}
	})

	t.Run("post request", func(t *testing.T) {
		fn := FreshChecker(freshCheckerConfig)
		req := httptest.NewRequest(http.MethodPost, "/users/me", nil)
		c := pike.NewContext(req)
		c.Resp = &cache.Response{
			StatusCode: http.StatusOK,
		}
		err := fn(c, pike.NoopNext)
		if err != nil {
			t.Fatalf("check post request fail, %v", err)
		}
		if c.Fresh {
			t.Fatalf("post request will not be fresh")
		}
	})

	t.Run("get request(502)", func(t *testing.T) {
		fn := FreshChecker(freshCheckerConfig)
		req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
		c := pike.NewContext(req)
		c.Resp = &cache.Response{
			StatusCode: http.StatusBadGateway,
		}
		err := fn(c, pike.NoopNext)
		if err != nil {
			t.Fatalf("check get request(502) fail, %v", err)
		}
		if c.Fresh {
			t.Fatalf("post request will not be fresh")
		}
	})

	t.Run("get reqeust(fresh)", func(t *testing.T) {
		fn := FreshChecker(freshCheckerConfig)
		req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
		c := pike.NewContext(req)
		c.Resp = &cache.Response{
			StatusCode: http.StatusOK,
		}
		c.Request.Header.Set(pike.HeaderIfNoneMatch, "ABCD")
		c.Response.Header().Set(pike.HeaderETag, "ABCD")
		err := fn(c, pike.NoopNext)
		if err != nil {
			t.Fatalf("check get reqeust(fresh) fail, %v", err)
		}
	})
}
