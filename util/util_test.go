package util

import (
	"bytes"
	"testing"

	"github.com/valyala/fasthttp"

	"github.com/vicanso/pike/vars"
)

func testPass(t *testing.T, uri, method string, resultExpected bool) {
	passList := [][]byte{
		[]byte("cache-control=no-cache"),
	}
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI(uri)
	ctx.Request.Header.SetMethod(method)
	result := Pass(ctx, passList)
	if result != resultExpected {
		t.Fatalf("unexpected result in Pass %q %q: %v. Expecting %v", method, uri, result, resultExpected)
	}
}

func TestPass(t *testing.T) {
	testPass(t, "http://127.0.0.1/", "GET", false)
	testPass(t, "http://127.0.0.1/", "HEAD", false)

	testPass(t, "http://127.0.0.1/?cache-control=no-cache", "GET", true)

	testPass(t, "http://127.0.0.1/i18ns", "POST", true)
}

func testGetCacheAge(t *testing.T, cacheControl []byte, resultExpected int) {
	ctx := &fasthttp.RequestCtx{}
	if cacheControl != nil {
		ctx.Response.Header.SetCanonical(vars.CacheControl, cacheControl)
	}
	result := GetCacheAge(ctx)
	if result != resultExpected {
		t.Fatalf("unexpected result in GetCacheAge %q: %v. Expecting %v", cacheControl, result, resultExpected)
	}
}

func TestGetCacheAge(t *testing.T) {
	testGetCacheAge(t, nil, 0)
	testGetCacheAge(t, []byte("max-age=30"), 30)
	testGetCacheAge(t, []byte("private,max-age=30"), 0)
	testGetCacheAge(t, []byte("no-store"), 0)
	testGetCacheAge(t, []byte("no-cache"), 0)
	testGetCacheAge(t, []byte("max-age=0"), 0)
	testGetCacheAge(t, []byte("s-maxage=10, max-age=30"), 10)
}

func testSupportCompress(t *testing.T, compress, acceptEncoding []byte, resultExpected bool) {
	ctx := &fasthttp.RequestCtx{}
	if acceptEncoding != nil {
		ctx.Request.Header.SetCanonical(vars.AcceptEncoding, acceptEncoding)
	}
	var result bool
	if bytes.Compare(compress, []byte("br")) == 0 {
		result = SupportBr(ctx)
	} else {
		result = SupportGzip(ctx)
	}
	if result != resultExpected {
		t.Fatalf("unexpected result in Support %q %q: %v. Expecting %v", compress, acceptEncoding, result, resultExpected)
	}
}

func TestSupportCompress(t *testing.T) {
	gzip := []byte("gzip")
	br := []byte("br")
	testSupportCompress(t, gzip, nil, false)
	testSupportCompress(t, br, nil, false)

	testSupportCompress(t, gzip, []byte("gzip"), true)
	testSupportCompress(t, br, []byte("br"), true)

	testSupportCompress(t, gzip, []byte("gzip, br"), true)
	testSupportCompress(t, br, []byte("gzip, br"), true)
}
