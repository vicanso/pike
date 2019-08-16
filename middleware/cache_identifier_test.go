package middleware

import (
	"bytes"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vicanso/elton"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
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
		h.Add(elton.HeaderSetCookie, "abc")
		assert.Equal(t, 0, getCacheAge(h), "set cookie response's max age should be 0")
	})

	t.Run("not set cache control", func(t *testing.T) {
		h := make(http.Header)
		assert.Equal(t, 0, getCacheAge(h), "not set cache control header's age should be 0")
	})

	t.Run("no-cache", func(t *testing.T) {
		h := make(http.Header)
		h.Add(elton.HeaderCacheControl, "no-cache")
		assert.Equal(t, 0, getCacheAge(h), "no cache header's age should be 0")
	})

	t.Run("no-store", func(t *testing.T) {
		h := make(http.Header)
		h.Add(elton.HeaderCacheControl, "no-store")
		assert.Equal(t, 0, getCacheAge(h), "no store header's age should be 0")
	})

	t.Run("private", func(t *testing.T) {
		h := make(http.Header)
		h.Add(elton.HeaderCacheControl, "private, max-age=10")
		assert.Equal(t, 0, getCacheAge(h), "private max age header's age should be 0")
	})

	t.Run("s-maxage", func(t *testing.T) {
		h := make(http.Header)
		h.Add(elton.HeaderCacheControl, "s-maxage=10, max-age=300")
		assert.Equal(t, 10, getCacheAge(h), "max age should get s-maxage first")
	})

	t.Run("max-age", func(t *testing.T) {
		h := make(http.Header)
		h.Add(elton.HeaderCacheControl, "max-age=300")
		assert.Equal(t, 300, getCacheAge(h), "max age should get max-age field")
	})

	t.Run("max-age with age", func(t *testing.T) {
		h := make(http.Header)
		h.Add(elton.HeaderCacheControl, "max-age=300")
		h.Add(df.HeaderAge, "10")
		assert.Equal(t, 290, getCacheAge(h), "max age should minus age field")
	})
}

func TestNewCacheIdentifier(t *testing.T) {
	dsp := cache.NewDispatcher(cache.Options{})
	bc := config.BasicConfig{}
	fn := NewCacheIdentifier(bc, dsp)
	t.Run("pass(post)", func(t *testing.T) {
		assert := assert.New(t)
		req := httptest.NewRequest("POST", "/", nil)
		c := &elton.Context{
			Request: req,
		}
		c.Next = func() error {
			return nil
		}
		err := fn(c)
		assert.Nil(err)
		assert.Equal(cache.Pass, c.Get(df.Status).(int), "post request should pass")
	})

	t.Run("pass(no cache)", func(t *testing.T) {
		assert := assert.New(t)
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(elton.HeaderCacheControl, "no-cache")
		c := &elton.Context{
			Request: req,
		}
		c.Next = func() error {
			return nil
		}
		err := fn(c)
		assert.Nil(err)
		assert.Equal(cache.Pass, c.Get(df.Status).(int), "get request(no cache) should pass")
	})

	t.Run("hit for pass", func(t *testing.T) {
		assert := assert.New(t)
		url := "/" + randomString(20)
		c1 := elton.NewContext(httptest.NewRecorder(), httptest.NewRequest("GET", url, nil))
		done := make(chan bool)
		go func() {
			c2 := elton.NewContext(httptest.NewRecorder(), httptest.NewRequest("GET", url, nil))
			c2.Next = func() error {
				return nil
			}
			err := fn(c2)
			assert.Nil(err)
			assert.Equal(cache.HitForPass, c2.Get(df.Status).(int), "the second request should hit for pass")
			done <- true
		}()
		c1.Next = func() error {
			time.Sleep(time.Second)
			c1.BodyBuffer = bytes.NewBufferString("abc")
			return nil
		}
		err := fn(c1)
		assert.Nil(err)
		assert.Equal(cache.Fetch, c1.Get(df.Status).(int), "the first request should fetch")
		<-done
	})

	t.Run("cacheable", func(t *testing.T) {
		assert := assert.New(t)
		url := "/" + randomString(20)
		c1 := elton.NewContext(httptest.NewRecorder(), httptest.NewRequest("GET", url, nil))
		done := make(chan bool)
		go func() {
			c2 := elton.NewContext(httptest.NewRecorder(), httptest.NewRequest("GET", url, nil))
			c2.Next = func() error {
				return nil
			}
			err := fn(c2)
			assert.Nil(err)
			assert.Equal(cache.Cacheable, c2.Get(df.Status).(int), "the second request should cacheable")
			done <- true
		}()
		c1.Next = func() error {
			time.Sleep(time.Second)
			c1.SetHeader(elton.HeaderCacheControl, "max-age=10")
			c1.BodyBuffer = bytes.NewBufferString("abc")
			return nil
		}
		err := fn(c1)
		assert.Nil(err)
		assert.Equal(cache.Fetch, c1.Get(df.Status).(int), "the first request should fetch")
		<-done
	})

	t.Run("set max-age but cache fail", func(t *testing.T) {
		assert := assert.New(t)
		url := "/" + randomString(20)
		c1 := elton.NewContext(httptest.NewRecorder(), httptest.NewRequest("GET", url, nil))
		done := make(chan bool)
		go func() {
			c2 := elton.NewContext(httptest.NewRecorder(), httptest.NewRequest("GET", url, nil))
			c2.Next = func() error {
				return nil
			}
			err := fn(c2)
			assert.Nil(err)
			assert.Equal(cache.HitForPass, c2.Get(df.Status).(int), "the second request should hit for pass")
			done <- true
		}()
		c1.Next = func() error {
			time.Sleep(time.Second)
			c1.SetHeader(elton.HeaderCacheControl, "max-age=10")
			c1.SetHeader(elton.HeaderContentEncoding, "gzip")
			c1.BodyBuffer = bytes.NewBufferString("abc")
			return nil
		}
		err := fn(c1)
		assert.Nil(err)
		assert.Equal(cache.Fetch, c1.Get(df.Status).(int), "the first request should fetch")
		<-done
	})
}
