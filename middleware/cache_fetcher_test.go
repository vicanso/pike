package custommiddleware

import (
	"testing"

	"github.com/mitchellh/go-server-timing"

	"github.com/vicanso/pike/vars"

	"github.com/labstack/echo"
	"github.com/vicanso/pike/cache"
)

func TestCacheFetcher(t *testing.T) {
	client := &cache.Client{
		Path: "/tmp/test.cache",
	}
	err := client.Init()
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
			resp := c.Get(vars.Response).(*cache.Response)
			if resp.TTL != 300 {
				t.Fatalf("fetch cache fail")
			}
			return nil
		})
		e := echo.New()
		c := e.NewContext(nil, nil)
		timing := &servertiming.Header{}
		c.Set(vars.Status, cache.Cacheable)
		c.Set(vars.Identity, identity)
		c.Set(vars.Timing, timing)
		c.Set(vars.RID, "a")
		fn(c)
	})

	t.Run("fectch with no status", func(t *testing.T) {
		fn := CacheFetcher(config, client)(func(c echo.Context) error {
			return nil
		})
		e := echo.New()
		c := e.NewContext(nil, nil)
		err := fn(c)
		if err != vars.ErrRequestStatusNotSet {
			t.Fatalf("fetch with no status should return error")
		}
	})

	t.Run("fetch is not cacheable", func(t *testing.T) {
		fn := CacheFetcher(config, client)(func(c echo.Context) error {
			resp := c.Get(vars.Response)
			if resp != nil {
				t.Fatalf("fetch is not cacheable fail")
			}
			return nil
		})
		e := echo.New()
		c := e.NewContext(nil, nil)
		c.Set(vars.Status, cache.Pass)
		c.Set(vars.RID, "a")
		fn(c)
	})

	t.Run("fetch cacheable but no identity", func(t *testing.T) {
		fn := CacheFetcher(config, client)(func(c echo.Context) error {
			return nil
		})
		e := echo.New()
		c := e.NewContext(nil, nil)
		c.Set(vars.Status, cache.Cacheable)
		c.Set(vars.RID, "a")
		err := fn(c)
		if err != vars.ErrIdentityStatusNotSet {
			t.Fatalf("fetch cacheable but not identity should return error")
		}
	})
}
