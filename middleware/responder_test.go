package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vicanso/cod"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/df"
)

func TestNewResponder(t *testing.T) {
	fn := NewResponder()

	t.Run("response body has been set", func(t *testing.T) {
		c := cod.NewContext(nil, nil)
		c.Next = func() error {
			return nil
		}
		c.BodyBuffer = bytes.NewBufferString("abc")
		err := fn(c)
		if err != nil {
			t.Fatalf("fail, %v", err)
		}
	})

	t.Run("no http cache", func(t *testing.T) {
		c := cod.NewContext(nil, nil)
		c.Next = func() error {
			return nil
		}
		err := fn(c)
		if err != nil {
			t.Fatalf("fail, %v", err)
		}
	})

	t.Run("invalid cache", func(t *testing.T) {
		c := cod.NewContext(nil, nil)
		c.Set(df.Cache, "1")
		c.Next = func() error {
			return nil
		}
		err := fn(c)
		if err != errCacheInvalid {
			t.Fatalf("invalid cache should return error")
		}
	})

	h := make(http.Header)
	responseIDKey := "X-Response-ID"
	responseID := "1234"

	h.Set(responseIDKey, responseID)
	buf := []byte("abcd")
	gzipBody, _ := cache.Gzip(buf)
	// mock brotli data
	brBody := []byte("abcd")
	hc := &cache.HTTPCache{
		CreatedAt:  time.Now().Unix() - 10,
		Headers:    h,
		Status:     cache.Cacheable,
		StatusCode: 200,
		GzipBody:   bytes.NewBuffer(gzipBody),
		BrBody:     bytes.NewBuffer(brBody),
	}

	t.Run("brotli cache", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		resp := httptest.NewRecorder()
		req.Header.Set(cod.HeaderAcceptEncoding, "gzip, deflate, br")
		c := cod.NewContext(resp, req)
		c.Set(df.Cache, hc)
		c.Next = func() error {
			return nil
		}
		err := fn(c)
		if err != nil {
			t.Fatalf("brotli cache fail, %v", err)
		}
		if !bytes.Equal(brBody, c.BodyBuffer.Bytes()) ||
			c.StatusCode != 200 ||
			c.GetHeader(df.HeaderAge) == "" ||
			c.GetHeader(responseIDKey) != responseID ||
			c.GetHeader(cod.HeaderContentEncoding) != "br" {
			t.Fatalf("brotli cache fail")
		}
	})

	t.Run("gzip cache", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		resp := httptest.NewRecorder()
		req.Header.Set(cod.HeaderAcceptEncoding, "gzip, deflate")
		c := cod.NewContext(resp, req)
		c.Set(df.Cache, hc)
		c.Next = func() error {
			return nil
		}
		err := fn(c)
		if err != nil {
			t.Fatalf("gzip cache fail, %v", err)
		}
		if !bytes.Equal(gzipBody, c.BodyBuffer.Bytes()) ||
			c.StatusCode != 200 ||
			c.GetHeader(df.HeaderAge) == "" ||
			c.GetHeader(responseIDKey) != responseID ||
			c.GetHeader(cod.HeaderContentEncoding) != "gzip" {
			t.Fatalf("gzip cache fail")
		}
	})

	t.Run("gunzip cache", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		resp := httptest.NewRecorder()
		c := cod.NewContext(resp, req)
		c.Set(df.Cache, hc)
		c.Next = func() error {
			return nil
		}
		err := fn(c)
		if err != nil {
			t.Fatalf("gunzip cache fail, %v", err)
		}
		if !bytes.Equal(buf, c.BodyBuffer.Bytes()) ||
			c.StatusCode != 200 ||
			c.GetHeader(df.HeaderAge) == "" ||
			c.GetHeader(responseIDKey) != responseID ||
			c.GetHeader(cod.HeaderContentEncoding) != "" {
			t.Fatalf("gunzip cache fail")
		}
	})

	t.Run("raw body cache", func(t *testing.T) {
		hc.GzipBody = nil
		hc.BrBody = nil
		hc.Body = bytes.NewBuffer(buf)
		req := httptest.NewRequest("GET", "/", nil)
		resp := httptest.NewRecorder()
		c := cod.NewContext(resp, req)
		c.Set(df.Cache, hc)
		c.Next = func() error {
			return nil
		}
		err := fn(c)
		if err != nil {
			t.Fatalf("raw body cache fail, %v", err)
		}
		if !bytes.Equal(buf, c.BodyBuffer.Bytes()) ||
			c.StatusCode != 200 ||
			c.GetHeader(df.HeaderAge) == "" ||
			c.GetHeader(responseIDKey) != responseID ||
			c.GetHeader(cod.HeaderContentEncoding) != "" {
			t.Fatalf("raw body cache fail")
		}
	})
}
