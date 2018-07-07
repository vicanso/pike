package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vicanso/pike/pike"
)

func TestPing(t *testing.T) {
	pingConfig := PingConfig{
		URL: "/ping",
	}
	fn := Ping(pingConfig)
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	c := pike.NewContext(req)
	err := fn(c, pike.NoopNext)
	if err != nil {
		t.Fatalf("ping middleware fail, %v", err)
	}
	if c.Response.Status() != http.StatusOK {
		t.Fatalf("ping response status should be ok")
	}
	if string(c.Response.Bytes()) != "pong" {
		t.Fatalf("ping response should be pong")
	}

}
