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

func BenchmarkGzip(b *testing.B) {
	buf := []byte("onaonreau0912jo3nfonawou09u219fnznahgi9123hfdlanh234o1no2u1ouf9dnqwhjo1zdfn`1m12m3lmgMANGN128DNQDNRJYFNSU28FYH32N1HDND1JXN")
	for i := 0; i < b.N; i++ {
		Gzip(buf, 0)
	}
}
