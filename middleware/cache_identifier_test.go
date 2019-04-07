package middleware

import (
	"bytes"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vicanso/cod"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/df"
)

func randomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(65 + rand.Intn(25)) //A=65 and Z = 65+25
	}
	return string(bytes)
}

func TestGetCacheAge(t *testing.T) {
	t.Run("set cookie", func(t *testing.T) {
		h := make(http.Header)
		h.Add(cod.HeaderSetCookie, "abc")
		if getCacheAge(h) != 0 {
			t.Fatalf("set cookie header's age should be 0")
		}
	})

	t.Run("not set cache control", func(t *testing.T) {
		h := make(http.Header)
		if getCacheAge(h) != 0 {
			t.Fatalf("not set cache control header's age should be 0")
		}
	})

	t.Run("no-cache", func(t *testing.T) {
		h := make(http.Header)
		h.Add(cod.HeaderCacheControl, "no-cache")
		if getCacheAge(h) != 0 {
			t.Fatalf("no cache header's age should be 0")
		}
	})

	t.Run("no-store", func(t *testing.T) {
		h := make(http.Header)
		h.Add(cod.HeaderCacheControl, "no-store")
		if getCacheAge(h) != 0 {
			t.Fatalf("no store header's age should be 0")
		}
	})

	t.Run("private", func(t *testing.T) {
		h := make(http.Header)
		h.Add(cod.HeaderCacheControl, "private, max-age=10")
		if getCacheAge(h) != 0 {
			t.Fatalf("private max age header's age should be 0")
		}
	})

	t.Run("s-maxage", func(t *testing.T) {
		h := make(http.Header)
		h.Add(cod.HeaderCacheControl, "s-maxage=10, max-age=300")
		if getCacheAge(h) != 10 {
			t.Fatalf("get s-maxage fail")
		}
	})

	t.Run("max-age", func(t *testing.T) {
		h := make(http.Header)
		h.Add(cod.HeaderCacheControl, "max-age=300")
		if getCacheAge(h) != 300 {
			t.Fatalf("get max-age fail")
		}
	})

	t.Run("max-age with age", func(t *testing.T) {
		h := make(http.Header)
		h.Add(cod.HeaderCacheControl, "max-age=300")
		h.Add(df.HeaderAge, "10")
		if getCacheAge(h) != 290 {
			t.Fatalf("get max-age fail")
		}
	})
}

func TestNewCacheIdentifier(t *testing.T) {
	fn := NewCacheIdentifier()
	t.Run("pass", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", nil)
		c := &cod.Context{
			Request: req,
		}
		c.Next = func() error {
			return nil
		}
		err := fn(c)
		if err != nil {
			t.Fatalf("post request pass fail, %v", err)
		}
		if c.Get(df.Status).(int) != cache.Pass {
			t.Fatalf("post request should pass")
		}
	})

	t.Run("hit for pass", func(t *testing.T) {
		url := "/" + randomString(20)
		c1 := cod.NewContext(httptest.NewRecorder(), httptest.NewRequest("GET", url, nil))
		done := make(chan bool)
		go func() {
			c2 := cod.NewContext(httptest.NewRecorder(), httptest.NewRequest("GET", url, nil))
			c2.Next = func() error {
				return nil
			}
			err := fn(c2)
			if err != nil {
				t.Fatalf("the second hit for pass request fail, %v", err)
			}
			if c2.Get(df.Status).(int) != cache.HitForPass {
				t.Fatalf("the second request should hit for pass")
			}
			done <- true
		}()
		c1.Next = func() error {
			time.Sleep(time.Second)
			c1.BodyBuffer = bytes.NewBufferString("abc")
			return nil
		}
		err := fn(c1)
		if err != nil {
			t.Fatalf("handle hit for pass request fail, %v", err)
		}
		if c1.Get(df.Status).(int) != cache.Fetch {
			t.Fatalf("the first request should be fetch")
		}
		<-done
	})

	t.Run("cacheable", func(t *testing.T) {
		url := "/" + randomString(20)
		c1 := cod.NewContext(httptest.NewRecorder(), httptest.NewRequest("GET", url, nil))
		done := make(chan bool)
		go func() {
			c2 := cod.NewContext(httptest.NewRecorder(), httptest.NewRequest("GET", url, nil))
			c2.Next = func() error {
				return nil
			}
			err := fn(c2)
			if err != nil {
				t.Fatalf("the second request fail, %v", err)
			}
			if c2.Get(df.Status).(int) != cache.Cacheable {
				t.Fatalf("the second request should cacheable")
			}
			done <- true
		}()
		c1.Next = func() error {
			time.Sleep(time.Second)
			c1.SetHeader(cod.HeaderCacheControl, "max-age=10")
			c1.BodyBuffer = bytes.NewBufferString("abc")
			return nil
		}
		err := fn(c1)
		if err != nil {
			t.Fatalf("handle cacheable request fail, %v", err)
		}
		if c1.Get(df.Status).(int) != cache.Fetch {
			t.Fatalf("the first request should be fetch")
		}
		<-done
	})

	t.Run("set max-age but cache fail", func(t *testing.T) {
		url := "/" + randomString(20)
		c1 := cod.NewContext(httptest.NewRecorder(), httptest.NewRequest("GET", url, nil))
		done := make(chan bool)
		go func() {
			c2 := cod.NewContext(httptest.NewRecorder(), httptest.NewRequest("GET", url, nil))
			c2.Next = func() error {
				return nil
			}
			err := fn(c2)
			if err != nil {
				panic("the second request fail, " + err.Error())
			}
			if c2.Get(df.Status).(int) != cache.HitForPass {
				panic("the second request should be hit for pass")
			}
			done <- true
		}()
		c1.Next = func() error {
			time.Sleep(time.Second)
			c1.SetHeader(cod.HeaderCacheControl, "max-age=10")
			c1.SetHeader(cod.HeaderContentEncoding, "gzip")
			c1.BodyBuffer = bytes.NewBufferString("abc")
			return nil
		}
		err := fn(c1)
		if err != nil {
			t.Fatalf("handle cacheable request fail, %v", err)
		}
		if c1.Get(df.Status).(int) != cache.Fetch {
			t.Fatalf("the first request should be fetch")
		}
		<-done
	})
}
