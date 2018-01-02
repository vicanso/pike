package dispatch

import (
	"bytes"
	"testing"

	"../util"
	"../vars"

	"github.com/valyala/fasthttp"
)

func TestGetResponseHeader(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	data := []byte("hello world")
	ctx.Response.SetBody(data)
	ctx.Response.Header.SetContentLength(len(data))
	ctx.Response.Header.SetCanonical(vars.CacheControl, []byte("public, max-age=30"))

	header := GetResponseHeader(&ctx.Response)
	if len(header) != 109 {
		t.Fatalf("get the header from response fail")
	}
}

func TestGetResponseBody(t *testing.T) {
	helloWorld := "hello world"
	data, _ := util.Gzip([]byte(helloWorld))
	ctx := &fasthttp.RequestCtx{}
	ctx.Response.Header.SetCanonical(vars.ContentEncoding, vars.Gzip)
	ctx.SetBody(data)

	body, err := GetResponseBody(&ctx.Response)
	if err != nil {
		t.Fatalf("get the response body fail, %v", err)
	}
	if string(body) != helloWorld {
		t.Fatalf("get the response body fail")
	}
}

func TestErrorHandler(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	ErrorHandler(ctx, vars.ErrServiceUnavailable)
	buf := ctx.Response.Body()

	if string(buf) != "service unavailable" || ctx.Response.StatusCode() != 503 {
		t.Fatalf("error handler fail")
	}
}

func TestResponseGzipBytes(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	helloWorld := "hello world"
	data, _ := util.Gzip([]byte(helloWorld))
	header := &fasthttp.ResponseHeader{}
	responseID := []byte("X8183211")
	header.SetCanonical([]byte("X-Response-Id"), responseID)
	ResponseGzipBytes(ctx, header.Header(), data)
	buf, _ := ctx.Response.BodyGunzip()
	if string(buf) != helloWorld {
		t.Fatalf("response gzip bytes fail")
	}
	if bytes.Compare(ctx.Response.Header.Peek("X-Response-Id"), responseID) != 0 {
		t.Fatalf("set response header fail")
	}
}
