package httplog

import (
	"bytes"
	"log"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
)

func TestParse(t *testing.T) {
	tags := Parse([]byte("Pike {host} {method} {path} {proto} {query} {remote} {client-ip} {scheme} {uri} {~jt} {>X-Request-Id} {<X-Response-Id} {when} {when-iso} {when-iso-ms} {when-unix} {status} {size} {referer} {userAgent} {latency} {latency-ms}ms"))
	count := 45 
	if len(tags) != count {
		t.Fatalf("the tags length expect %v but %v", count, len(tags))
	}
	startedAt := time.Now()

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("http://aslant.site:5000/users/me?cache-control=no-cache")
	ctx.Request.Header.SetCookie("jt", "cookieValue")
	ctx.Request.Header.SetCanonical([]byte("X-Request-Id"), []byte("requestId"))
	ctx.Response.Header.SetCanonical([]byte("X-Response-Id"), []byte("responseId"))
	ctx.SetBody([]byte("hello world"))
	buf := Format(ctx, tags, startedAt)
	log.Print(string(buf))
	if bytes.Index(buf, []byte("{")) != -1 {
		t.Fatalf("the log of request fail")
	}
}
