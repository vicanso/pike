package custommiddleware

import (
	"testing"

	"github.com/vicanso/pike/vars"

	"github.com/labstack/echo"
	"github.com/vicanso/pike/cache"
)

func TestCacheFetcher(t *testing.T) {
	client := &cache.Client{
		Path: "/tmp/test.cache",
	}
	err := client.Init(0)
	if err != nil {
		t.Fatalf("cache init fail, %v", err)
	}
	defer client.Close()
	config := CacheFetcherConfig{}
	t.Run("cache fetch", func(t *testing.T) {
		identity := []byte("ABCD")
		client.SaveResponse(identity, &cache.Response{
			TTL: 300,
		})

		fn := CacheFetcher(config, client)(func(c echo.Context) error {
			pc := c.(*Context)
			resp := pc.resp
			if resp.TTL != 300 {
				t.Fatalf("fetch cache fail")
			}
			return nil
		})
		e := echo.New()
		pc := NewContext(e.NewContext(nil, nil))
		pc.status = cache.Cacheable
		pc.identity = identity
		fn(pc)
	})

	t.Run("fectch with no status", func(t *testing.T) {
		fn := CacheFetcher(config, client)(func(c echo.Context) error {
			return nil
		})
		e := echo.New()
		pc := NewContext(e.NewContext(nil, nil))
		err := fn(pc)
		if err != vars.ErrRequestStatusNotSet {
			t.Fatalf("fetch with no status should return error")
		}
	})

	t.Run("fetch is not cacheable", func(t *testing.T) {
		fn := CacheFetcher(config, client)(func(c echo.Context) error {
			pc := c.(*Context)
			resp := pc.resp
			if resp != nil {
				t.Fatalf("fetch is not cacheable fail")
			}
			return nil
		})
		e := echo.New()
		pc := NewContext(e.NewContext(nil, nil))
		pc.status = cache.Pass
		fn(pc)
	})

	t.Run("fetch cacheable but no identity", func(t *testing.T) {
		fn := CacheFetcher(config, client)(func(c echo.Context) error {
			return nil
		})
		e := echo.New()
		pc := NewContext(e.NewContext(nil, nil))
		pc.status = cache.Cacheable
		err := fn(pc)
		if err != vars.ErrIdentityNotSet {
			t.Fatalf("fetch cacheable but not identity should return error")
		}
	})
}
