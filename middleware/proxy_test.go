package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vicanso/cod"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/df"
	"github.com/vicanso/pike/upstream"

	gock "gopkg.in/h2non/gock.v1"
)

func TestNewProxy(t *testing.T) {
	assert := assert.New(t)
	us := upstream.New(config.BackendConfig{
		Backends: []string{
			"http://127.0.0.1:7001",
		},
		Hosts: []string{
			"aslant.site",
		},
		ResponseHeader: []string{
			"X-Server:test",
		},
		RequestHeader: []string{
			"X-Request-ID:123",
		},
	}, nil)
	for _, item := range us.Server.GetUpstreamList() {
		item.Healthy()
	}
	upstreams := make(upstream.Upstreams, 0)
	upstreams = append(upstreams, us)
	director := &upstream.Director{
		Upstreams: upstreams,
	}

	defer gock.Off()
	gock.New("http://127.0.0.1:7001").
		Get("/").
		Reply(200).
		JSON(map[string]string{
			"foo": "bar",
		})

	req := httptest.NewRequest("GET", "http://aslant.site/", nil)
	req.Header.Set(cod.HeaderAcceptEncoding, df.GZIP)
	req.Header.Set(cod.HeaderIfModifiedSince, "modified date")
	req.Header.Set(cod.HeaderIfNoneMatch, `"ETag"`)
	resp := httptest.NewRecorder()
	c := cod.NewContext(resp, req)
	c.Next = func() error {
		return nil
	}
	fn := NewProxy(director)
	err := fn(c)
	assert.Nil(err, "proxy middleware fail")
	assert.Equal(http.StatusOK, c.StatusCode, "proxy should be 200")
	assert.Equal(`{"foo":"bar"}`, strings.TrimSpace(c.BodyBuffer.String()))
}
