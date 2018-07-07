package middleware

import (
	"net/http"
	"testing"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/pike"
)

func TestHeaderSetter(t *testing.T) {
	headerSetterConfig := HeaderSetterConfig{}
	t.Run("set header(response not set)", func(t *testing.T) {
		fn := HeaderSetter(headerSetterConfig)
		c := pike.NewContext(nil)
		err := fn(c, func() error {
			return nil
		})
		if err != ErrResponseNotSet {
			t.Fatalf("response not set should return error")
		}
	})

	t.Run("set header", func(t *testing.T) {
		fn := HeaderSetter(headerSetterConfig)
		c := pike.NewContext(nil)
		resp := &cache.Response{
			Header: http.Header{
				"Token": []string{
					"ABCD",
				},
				"Date": []string{
					"Sat, 28 Apr 2018 02:59:16 GMT",
				},
			},
		}
		c.Resp = resp
		err := fn(c, func() error {
			return nil
		})
		if err != nil {
			t.Fatalf("set header fail, %v", err)
		}
		header := c.Response.Header()
		if len(header) != 1 {
			t.Fatalf("header length should be 1")
		}
		if header["Token"][0] != "ABCD" {
			t.Fatalf("header token field should be ABCD")
		}
	})
}
