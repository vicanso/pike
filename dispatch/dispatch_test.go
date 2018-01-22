package dispatch

import (
	"testing"
	"time"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/util"
	"github.com/vicanso/pike/vars"

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
	Response(ctx, respData)
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
}
