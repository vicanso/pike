package util

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkGetIdentity(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/users/me?cache=no-cache&id=1", nil)
	for i := 0; i < b.N; i++ {
		_ = GetIdentity(req)
	}
}

func BenchmarkGenerateGetIdentity(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/users/me?cache=no-cache&id=1", nil)
	fn := GenerateGetIdentity("method host uri")
	for i := 0; i < b.N; i++ {
		_ = fn(req)
	}
}
