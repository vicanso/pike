package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vicanso/elton"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/df"
)

func TestNewResponder(t *testing.T) {
	fn := NewResponder()

	t.Run("response body has been set", func(t *testing.T) {
		assert := assert.New(t)
		c := elton.NewContext(nil, nil)
		c.Next = func() error {
			return nil
		}
		c.BodyBuffer = bytes.NewBufferString("abc")
		err := fn(c)
		assert.Nil(err, "response middleware fail")
	})

	t.Run("no http cache", func(t *testing.T) {
		assert := assert.New(t)
		c := elton.NewContext(nil, nil)
		c.Next = func() error {
			return nil
		}
		err := fn(c)
		assert.Nil(err, "response middleware fail")
	})

	t.Run("invalid cache", func(t *testing.T) {
		assert := assert.New(t)
		c := elton.NewContext(nil, nil)
		c.Set(df.Cache, "1")
		c.Next = func() error {
			return nil
		}
		err := fn(c)
		assert.Equal(errCacheInvalid, err, "invalid cache should return error")
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
		assert := assert.New(t)
		req := httptest.NewRequest("GET", "/", nil)
		resp := httptest.NewRecorder()
		req.Header.Set(elton.HeaderAcceptEncoding, "gzip, deflate, br")
		c := elton.NewContext(resp, req)
		c.Set(df.Cache, hc)
		c.Next = func() error {
			return nil
		}
		err := fn(c)
		assert.Nil(err, "brotli cache fail")
		assert.Equal(brBody, c.BodyBuffer.Bytes())
		assert.Equal(200, c.StatusCode)
		assert.NotEqual("", c.GetHeader(df.HeaderAge))
		assert.Equal(responseID, c.GetHeader(responseIDKey))
		assert.Equal("br", c.GetHeader(elton.HeaderContentEncoding))
	})

	t.Run("gzip cache", func(t *testing.T) {
		assert := assert.New(t)
		req := httptest.NewRequest("GET", "/", nil)
		resp := httptest.NewRecorder()
		req.Header.Set(elton.HeaderAcceptEncoding, "gzip, deflate")
		c := elton.NewContext(resp, req)
		c.Set(df.Cache, hc)
		c.Next = func() error {
			return nil
		}
		err := fn(c)
		assert.Nil(err, "gzip cache fail")
		assert.Equal(gzipBody, c.BodyBuffer.Bytes())
		assert.Equal(200, c.StatusCode)
		assert.NotEqual("", c.GetHeader(df.HeaderAge))
		assert.Equal(responseID, c.GetHeader(responseIDKey))
		assert.Equal("gzip", c.GetHeader(elton.HeaderContentEncoding))
	})

	t.Run("gunzip cache", func(t *testing.T) {
		assert := assert.New(t)
		req := httptest.NewRequest("GET", "/", nil)
		resp := httptest.NewRecorder()
		c := elton.NewContext(resp, req)
		c.Set(df.Cache, hc)
		c.Next = func() error {
			return nil
		}
		err := fn(c)
		assert.Nil(err, "gunzip cache fail")
		assert.Equal(buf, c.BodyBuffer.Bytes())
		assert.Equal(200, c.StatusCode)
		assert.NotEqual("", c.GetHeader(df.HeaderAge))
		assert.Equal(responseID, c.GetHeader(responseIDKey))
		assert.Equal("", c.GetHeader(elton.HeaderContentEncoding))
	})

	t.Run("raw body cache", func(t *testing.T) {
		assert := assert.New(t)
		hc.GzipBody = nil
		hc.BrBody = nil
		hc.Body = bytes.NewBuffer(buf)
		req := httptest.NewRequest("GET", "/", nil)
		resp := httptest.NewRecorder()
		c := elton.NewContext(resp, req)
		c.Set(df.Cache, hc)
		c.Next = func() error {
			return nil
		}
		err := fn(c)
		assert.Nil(err, "raw body cache fail")
		assert.Equal(buf, c.BodyBuffer.Bytes())
		assert.Equal(200, c.StatusCode)
		assert.NotEqual("", c.GetHeader(df.HeaderAge))
		assert.Equal(responseID, c.GetHeader(responseIDKey))
		assert.Equal("", c.GetHeader(elton.HeaderContentEncoding))
	})
}
