package dispatch

import (
	"testing"
	"time"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/vars"

	"github.com/valyala/fasthttp"
)

func TestErrorHandler(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	ErrorHandler(ctx, vars.ErrServiceUnavailable)
	buf := ctx.Response.Body()

	if string(buf) != "service unavailable" || ctx.Response.StatusCode() != 503 {
		t.Fatalf("error handler fail")
	}
}

func TestResponse(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	helloWorld := "hello world"
	data := []byte(helloWorld)
	header := &fasthttp.ResponseHeader{}
	responseID := []byte("X8183211")
	header.SetCanonical([]byte("X-Response-Id"), responseID)

	createdAt := uint32(time.Now().Unix())
	respData := &cache.ResponseData{
		CreatedAt:  createdAt - 20,
		StatusCode: 200,
		TTL:        300,
		Header:     header.Header(),
		Body:       data,
	}
	Response(ctx, respData, 1024)
	body := string(ctx.Response.Body())
	if body != helloWorld {
		t.Fatalf("the response data expect %q but %q", helloWorld, body)
	}
	age := string(ctx.Response.Header.PeekBytes(vars.Age))
	if age != "20" {
		t.Fatalf("the response age expece 20 but %q", age)
	}
	respID := string(ctx.Response.Header.Peek("X-Response-Id"))
	if respID != string(responseID) {
		t.Fatalf("the response id expect %q but %q", respID, string(responseID))
	}

	// 304
	header = &fasthttp.ResponseHeader{}
	hashValue := []byte("ABCD")
	header.SetBytesKV(vars.ETag, hashValue)
	ctx = &fasthttp.RequestCtx{}
	ctx.Request.Header.SetBytesKV(vars.IfNoneMatch, hashValue)
	respData = &cache.ResponseData{
		CreatedAt:  createdAt - 20,
		StatusCode: 200,
		TTL:        300,
		Header:     header.Header(),
		Body:       data,
	}
	Response(ctx, respData, 1024)
	if ctx.Response.StatusCode() != 304 {
		t.Fatalf("the not modified response fail")
	}

	// gzip
	header = &fasthttp.ResponseHeader{}
	ctx = &fasthttp.RequestCtx{}
	ctx.Request.Header.SetCanonical(vars.AcceptEncoding, []byte("gzip, deflate"))
	buf := make([]byte, 4096)
	for index := 0; index < 4096; index++ {
		buf[index] = byte('A')
	}
	respData = &cache.ResponseData{
		ShouldCompress: true,
		CreatedAt:      createdAt - 20,
		StatusCode:     200,
		TTL:            300,
		Header:         header.Header(),
		Body:           buf,
	}
	Response(ctx, respData, 1024)
	if len(ctx.Response.Body()) != 43 {
		t.Fatalf("the gzip response fail")
	}
}
