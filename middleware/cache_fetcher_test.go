package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/pike"
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
		identity := []byte("GET aslant.site /cache")
		var ttl uint16 = 300
		client.SaveResponse(identity, &cache.Response{
			TTL: ttl,
		})
		fn := CacheFetcher(config, client)

		r := httptest.NewRequest(http.MethodGet, "http://aslant.site/cache", nil)
		c := pike.NewContext(r)
		c.Status = cache.Cacheable
		c.Identity = identity
		err := fn(c, func() error {
			return nil
		})
		if err != nil {
			t.Fatalf("cache fetch fail, %v", err)
		}
		if c.Resp.TTL != ttl {
			t.Fatalf("get cache fail")
		}
	})

	t.Run("fetch with no status", func(t *testing.T) {
		fn := CacheFetcher(config, client)
		c := pike.NewContext(nil)
		err := fn(c, func() error {
			return nil
		})
		if err != ErrRequestStatusNotSet {
			t.Fatalf("fetch with no status should return error")
		}
	})

	t.Run("pass cache fetch", func(t *testing.T) {
		fn := CacheFetcher(config, client)
		c := pike.NewContext(nil)
		c.Status = cache.Pass
		err := fn(c, func() error {
			return nil
		})
		if err != nil {
			t.Fatalf("pass fetch fail, %v", err)
		}
		if c.Resp != nil {
			t.Fatalf("the response should be nil")
		}
	})

	t.Run("fetch cacheable but no identity", func(t *testing.T) {
		fn := CacheFetcher(config, client)
		c := pike.NewContext(nil)
		c.Status = cache.Cacheable
		err := fn(c, func() error {
			return nil
		})
		if err != ErrIdentityNotSet {
			t.Fatalf("fetch cacheable but not identity should return error")
		}
	})
}
