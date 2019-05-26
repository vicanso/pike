package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vicanso/cod"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/stats"
)

func TestNewInitialization(t *testing.T) {
	assert := assert.New(t)
	cfg := config.New()
	cfg.Viper.Set("header", []string{
		"X-Response-ID:456",
	})
	cfg.Viper.Set("request_header", []string{
		"X-Request-ID:123",
	})

	fn := NewInitialization(cfg, stats.New())
	resp := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	c := cod.NewContext(resp, req)
	c.Next = func() error {
		assert.Equal(c.GetHeader("X-Response-ID"), "")
		assert.Equal(c.GetRequestHeader("X-Request-ID"), "123", "X-Request-ID should be set")
		return nil
	}
	err := fn(c)
	assert.Nil(err, "init middleware fail")
	assert.Equal(c.GetHeader("X-Response-ID"), "456", "X-Response-ID should be set")
}

func TestTooManyRequest(t *testing.T) {
	assert := assert.New(t)
	cfg := config.New()
	max := cfg.GetConcurrency()
	cfg.Viper.Set("concurrency", 1)
	defer cfg.Viper.Set("concurrency", max)
	fn := NewInitialization(cfg, stats.New())

	c1 := cod.NewContext(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	c1.Next = func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}
	go func() {
		err := fn(c1)
		assert.Nil(err)
	}()
	c2 := cod.NewContext(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	time.Sleep(time.Millisecond)
	err := fn(c2)
	assert.Equal(err, errTooManyRequest)
}
