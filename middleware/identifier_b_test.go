package middleware

import (
	"net/http"
	"testing"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/pike"
	"github.com/vicanso/pike/util"
)

func BenchmarkIdentifier(b *testing.B) {
	req := &http.Request{
		Method:     "GET",
		Host:       "aslant.site",
		RequestURI: "/users/me",
	}
	c := pike.NewContext(req)
	client := &cache.Client{
		Path: "/tmp/test.cache",
	}

	err := client.Init()
	if err != nil {
		panic(err)
	}
	defer client.Close()
	key := util.GetIdentity(req)
	client.GetRequestStatus(key)
	client.Cacheable(key, 100)

	fn := Identifier(IdentifierConfig{}, client)
	for j := 0; j < 20; j++ {
		go func() {
			for i := 0; i < b.N; i++ {
				fn(c, pike.NoopNext)
			}
		}()
	}

	for i := 0; i < b.N; i++ {
		fn(c, pike.NoopNext)
	}
}
