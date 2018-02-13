package httplog

import (
	"bytes"
	"log"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
)

func TestParse(t *testing.T) {
	tags := Parse([]byte("Pike {host}{method} {path} {proto} {query} {remote} {client-ip} {scheme} {uri} {~jt} {>X-Request-Id} {<X-Response-Id} {when} {when-iso} {when-iso-ms} {when-unix} {status} {payload-size} {size} {referer} {userAgent} {latency} {latency-ms}ms"))
	count := 46
	if len(tags) != count {
		t.Fatalf("the tags length expect %v but %v", count, len(tags))
	}
	startedAt := time.Now()

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("http://aslant.site:5000/users/login?cache-control=no-cache")
	ctx.Request.Header.Set("Referer", "http://pike.aslant.site/")
	ctx.Request.Header.Set("User-Agent", "fasthttp/client")
	ctx.Request.Header.SetMethod("POST")
	ctx.Request.SetBody([]byte("{\"name\": \"vicanso\"}"))
	ctx.Request.Header.SetCookie("jt", "cookieValue")
	ctx.Request.Header.SetCanonical([]byte("X-Request-Id"), []byte("requestId"))
	ctx.Response.Header.SetCanonical([]byte("X-Response-Id"), []byte("responseId"))
	ctx.SetBody([]byte("hello world"))
	buf := Format(ctx, tags, startedAt)
	log.Print(string(buf))
	if bytes.Index(buf, []byte("{")) != -1 {
		t.Fatalf("the log of request fail")
	}

	tags = Parse([]byte(""))
	if len(tags) != 0 {
		t.Fatalf("the empty log format should be null")
	}
}
