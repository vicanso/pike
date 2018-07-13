package middleware

import (
	"net/http"
	"os"
	"testing"

	"github.com/vicanso/pike/performance"
	"github.com/vicanso/pike/pike"
)

func TestInitialization(t *testing.T) {
	conf := InitializationConfig{
		Header: []string{
			"X-Token:ABCD",
			"X-Server:${SERVER}",
		},
		RequestHeader: []string{
			"X-Server:${SERVER}",
		},
		Concurrency: 1,
	}
	err := os.Setenv("SERVER", "puma")
	if err != nil {
		t.Fatalf("set env fail, %v", err)
	}
	fn := Initialization(conf)
	r := &http.Request{
		Header: make(http.Header),
	}
	c := pike.NewContext(r)
	err = fn(c, func() error {
		return nil
	})
	if err != nil {
		t.Fatalf("init middleware fail, %v", err)
	}
	if c.Response.Header().Get("X-Token") != "ABCD" {
		t.Fatalf("init middleware set header fail")
	}
	if c.Response.Header().Get("X-Server") != "puma" {
		t.Fatalf("init middleware set header fail")
	}
	if c.Request.Header.Get("X-Server") != "puma" {
		t.Fatalf("init middleware set request header fail")
	}
	performance.IncreaseConcurrency()
	err = fn(c, func() error {
		return nil
	})
	if err != ErrTooManyRequest {
		t.Fatalf("init middleware should throw too many request error")
	}
}
