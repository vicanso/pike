package util

import (
	"net/http"
	"testing"
)

func BenchmarkGetIdentity(b *testing.B) {
	req := &http.Request{
		Method:     "GET",
		Host:       "aslant.site",
		RequestURI: "/users/me",
	}
	for i := 0; i < b.N; i++ {
		_ = GetIdentity(req)
	}
}
