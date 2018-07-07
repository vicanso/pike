package middleware

import (
	"net/http"
	"testing"

	"github.com/vicanso/pike/performance"
	"github.com/vicanso/pike/pike"
)

func TestInitialization(t *testing.T) {
	conf := InitializationConfig{
		Header: []string{
			"X-Token:ABCD",
		},
		Concurrency: 1,
	}
	fn := Initialization(conf)
	r := &http.Request{}
	c := pike.NewContext(r)
	err := fn(c, func() error {
		return nil
	})
	if err != nil {
		t.Fatalf("init middleware fail, %v", err)
	}
	if c.Response.Header().Get("X-Token") != "ABCD" {
		t.Fatalf("init middleware set header fail")
	}
	performance.IncreaseConcurrency()
	err = fn(c, func() error {
		return nil
	})
	if err != ErrTooManyRequest {
		t.Fatalf("init middleware should throw too many request error")
	}
}
