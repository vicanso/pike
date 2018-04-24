package customMiddleware

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
	err := client.Init()
	if err != nil {
		t.Fatalf("cache init fail, %v", err)
	}
	defer client.Close()
	t.Run("cache fetch", func(t *testing.T) {
		identity := []byte("ABCD")
		client.SaveResponse(identity, &cache.Response{
			TTL: 300,
		})

		fn := CacheFetcher(client)(func(c echo.Context) error {
			resp := c.Get(vars.Response).(*cache.Response)
			if resp.TTL != 300 {
				t.Fatalf("fetch cache fail")
			}
			return nil
		})
		e := echo.New()
		c := e.NewContext(nil, nil)
		c.Set(vars.Status, cache.Cacheable)
		c.Set(vars.Identity, identity)
		fn(c)
	})
}
