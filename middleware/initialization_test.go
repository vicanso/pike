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
	bc := config.BasicConfig{
		Concurrency: 100,
		ResponseHeader: []string{
			"X-Response-ID:456",
		},
		RequestHeader: []string{
			"X-Request-ID:123",
		},
	}

	fn := NewInitialization(bc, stats.New())
	resp := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	c := cod.NewContext(resp, req)
	c.Next = func() error {
		assert.Equal("", c.GetHeader("X-Response-ID"))
		assert.Equal("123", c.GetRequestHeader("X-Request-ID"), "X-Request-ID should be set")
		return nil
	}
	err := fn(c)
	assert.Nil(err, "init middleware fail")
	assert.Equal("456", c.GetHeader("X-Response-ID"), "X-Response-ID should be set")
}

func TestTooManyRequest(t *testing.T) {
	assert := assert.New(t)
	bc := config.BasicConfig{
		Concurrency: 1,
	}
	fn := NewInitialization(bc, stats.New())

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
	assert.Equal(errTooManyRequest, err)
}
