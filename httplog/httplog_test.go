package httplog

import (
	"bytes"
	"log"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
)

func TestParse(t *testing.T) {
	tags := Parse([]byte("Pike {host} {method} {path} {proto} {query} {remote} {scheme} {uri} {when} {when_iso} {when_unix} {status} {size} {referer} {user-agent} {latency} {latency_ms}ms"))
	count := 35
	if len(tags) != count {
		t.Fatalf("the tags length expect %v but %v", count, len(tags))
	}
	startedAt := time.Now()

	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("http://aslant.site:5000/users/me?cache-control=no-cache")
	ctx.SetBody([]byte("hello world"))
	buf := Format(ctx, tags, startedAt)
	log.Print(string(buf))
	if bytes.Index(buf, []byte("{")) != -1 {
		t.Fatalf("the log of request fail")
	}
}
