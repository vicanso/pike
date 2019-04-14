package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vicanso/cod"
	"github.com/vicanso/pike/df"
	"github.com/vicanso/pike/upstream"

	gock "gopkg.in/h2non/gock.v1"
)

func TestNewProxy(t *testing.T) {
	us := upstream.New(upstream.Backend{
		Backends: []string{
			"http://127.0.0.1:7001",
		},
		Hosts: []string{
			"aslant.site",
		},
		Header: []string{
			"X-Server:test",
		},
		RequestHeader: []string{
			"X-Request-ID:123",
		},
	})
	for _, item := range us.Server.GetUpstreamList() {
		item.Healthy()
	}
	upstreams := make(upstream.Upstreams, 0)
	upstreams = append(upstreams, us)
	// upstream.Add(us)

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
	done := false
	c.Set(df.ProxyDoneCallback, func() {
		done = true
	})
	fn := NewProxy(upstreams)
	err := fn(c)
	if err != nil ||
		c.StatusCode != http.StatusOK ||
		strings.TrimSpace(c.BodyBuffer.String()) != `{"foo":"bar"}` ||
		!done {
		t.Fatalf("proxy handler fail, %v", err)
	}

}
