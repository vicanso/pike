package middleware

import (
	"net/http"
	"testing"
	"time"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/pike"
	"github.com/vicanso/pike/util"
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

	conf := IdentifierConfig{
		Format: "method host uri",
	}

	fn := Identifier(conf, client)

	r := &http.Request{}

	r.Method = http.MethodGet
	r.Host = "127.0.0.1"
	r.RequestURI = "/users/me?cache-control=no-cache"
	t.Run("fetching", func(t *testing.T) {
		c := pike.NewContext(&http.Request{})
		err = fn(c, func() error {
			return nil
		})
		if err != nil {
			t.Fatalf("identifier middleware fail, %v", err)
		}
		if c.Status != cache.Pass {
			t.Fatalf("the request should be pass(Not GET OR HEAD)")
		}
		c.Request = r
		err = fn(c, func() error {
			return nil
		})
		if err != nil {
			t.Fatalf("identifier middleware fail, %v", err)
		}
		if c.Status != cache.Fetching {
			t.Fatalf("the request of get should be fetching")
		}
		if string(c.Identity) != "GET 127.0.0.1 /users/me?cache-control=no-cache" {
			t.Fatalf("the identity of request is wrong")
		}
	})

	t.Run("waiting status", func(t *testing.T) {
		c := pike.NewContext(r)
		go func() {
			time.Sleep(10 * time.Millisecond)
			client.UpdateRequestStatus(util.GetIdentity(r), cache.HitForPass, 300)
		}()
		err = fn(c, func() error {
			return nil
		})
		if err != nil {
			t.Fatalf("wait for status fail, %v", err)
		}
		if c.Status != cache.HitForPass {
			t.Fatalf("the wait for status should be hit for pass")
		}
	})
}
