package cache

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	"github.com/vicanso/cod"
	"github.com/vicanso/pike/df"
)

func TestGetHTTPCache(t *testing.T) {
	key := []byte("abc")
	hc := GetHTTPCache(key)
	if hc == nil {
		t.Fatalf("get http cache fail")
	}
}

func TestDispatcher(t *testing.T) {
	dsp := NewDispatcher(10)
	t.Run("get status(hit for pass)", func(t *testing.T) {
		k1 := []byte("abc")
		k2 := []byte("abc")
		k3 := []byte("abc")
		hc1 := dsp.GetHTTPCache(k1)
		go func() {
			time.Sleep(time.Millisecond)
			hc1.HitForPass()
			hc1.Unlock()
		}()
		hc2 := dsp.GetHTTPCache(k2)
		defer hc2.RUnlock()
		hc3 := dsp.GetHTTPCache(k3)
		defer hc3.RUnlock()
		if hc2.Status != HitForPass ||
			hc3.Status != HitForPass ||
			hc2 != hc3 {
			t.Fatalf("get http cache fail")
		}
	})

	t.Run("get status(cacheable)", func(t *testing.T) {
		k1 := []byte("def")
		k2 := []byte("def")
		k3 := []byte("def")
		hc1 := dsp.GetHTTPCache(k1)
		go func() {
			time.Sleep(time.Millisecond)
			hc1.Cacheable(10, &cod.Context{
				Headers:    make(http.Header),
				BodyBuffer: bytes.NewBufferString("abcd"),
			})
			hc1.Unlock()
		}()
		hc2 := dsp.GetHTTPCache(k2)
		defer hc2.RUnlock()
		hc3 := dsp.GetHTTPCache(k3)
		defer hc3.RUnlock()
		if hc2.Status != Cacheable ||
			hc3.Status != Cacheable ||
			hc2 != hc3 {
			t.Fatalf("get http cache fail")
		}
	})
}

func TestHitForPass(t *testing.T) {
	hc := HTTPCache{}
	hc.HitForPass()
	if hc.CreatedAt == 0 ||
		hc.ExpiredAt == 0 ||
		hc.Status != HitForPass {
		t.Fatalf("set hit for pass fail")
	}
}

func TestCacheable(t *testing.T) {
	t.Run("cacheable", func(t *testing.T) {
		header := make(http.Header)
		header.Set(cod.HeaderContentType, "text/html")
		header.Set(df.HeaderContentLength, "10")
		buf := make([]byte, 4096)
		maxAge := 10
		c := &cod.Context{
			Headers:    header,
			BodyBuffer: bytes.NewBuffer(buf),
			StatusCode: 200,
		}
		hc := &HTTPCache{}
		hc.Cacheable(maxAge, c)
		if hc.Status != Cacheable ||
			hc.Headers == nil ||
			hc.Body != nil ||
			hc.GzipBody == nil ||
			hc.Headers.Get(df.HeaderContentLength) != "" {
			t.Fatalf("set cacheable fail")
		}
	})

	t.Run("cacheable(gunzip success)", func(t *testing.T) {
		header := make(http.Header)
		header.Set(cod.HeaderContentType, "text/html")
		header.Set(cod.HeaderContentEncoding, "gzip")
		data := []byte("abcd")
		buf, _ := doGzip(data)
		maxAge := 10
		c := &cod.Context{
			Headers:    header,
			BodyBuffer: bytes.NewBuffer(buf),
			StatusCode: 200,
		}
		hc := &HTTPCache{}
		hc.Cacheable(maxAge, c)
		if hc.Status != Cacheable ||
			!bytes.Equal(data, hc.Body.Bytes()) ||
			hc.GzipBody != nil {
			t.Fatalf("gunzip success should cacheable")
		}
	})

	t.Run("gunzip fail should hit for pass", func(t *testing.T) {
		header := make(http.Header)
		header.Set(cod.HeaderContentType, "text/html")
		header.Set(cod.HeaderContentEncoding, "gzip")
		buf := make([]byte, 4096)
		maxAge := 10
		c := &cod.Context{
			Headers:    header,
			BodyBuffer: bytes.NewBuffer(buf),
			StatusCode: 200,
		}
		hc := &HTTPCache{}
		hc.Cacheable(maxAge, c)
		if hc.Status != HitForPass {
			t.Fatalf("gunzip fail should hit for pass")
		}
	})

	t.Run("not support encoding should hit for pass", func(t *testing.T) {
		header := make(http.Header)
		header.Set(cod.HeaderContentType, "text/html")
		header.Set(cod.HeaderContentEncoding, "xx")
		buf := make([]byte, 4096)
		maxAge := 10
		c := &cod.Context{
			Headers:    header,
			BodyBuffer: bytes.NewBuffer(buf),
			StatusCode: 200,
		}
		hc := &HTTPCache{}
		hc.Cacheable(maxAge, c)
		if hc.Status != HitForPass {
			t.Fatalf("not support encoding should hit for pass")
		}
	})
}
