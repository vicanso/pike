package util

import (
	"bytes"
	"testing"
	"time"

	"../vars"

	"github.com/valyala/fasthttp"
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

func testGetCacheAge(t *testing.T, cacheControl []byte, resultExpected uint32) {
	ctx := &fasthttp.RequestCtx{}
	if cacheControl != nil {
		ctx.Response.Header.SetCanonical(vars.CacheControl, cacheControl)
	}
	result := GetCacheAge(&ctx.Response.Header)
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

func TestConvert(t *testing.T) {
	ttl := uint16(1000)
	buf := ConvertUint16ToBytes(ttl)
	if ConvertBytesToUint16(buf) != ttl {
		t.Fatalf("the convert uint16 fail")
	}
	now := uint32(time.Now().Unix())
	buf = ConvertUint32ToBytes(now)
	if ConvertBytesToUint32(buf) != now {
		t.Fatalf("the convert uint32 fail")
	}
}

func TestSeconds(t *testing.T) {
	now := uint32(time.Now().Unix())
	buf := GetNowSecondsBytes()
	seconds := ConvertToSeconds(buf)
	if now != seconds {
		t.Fatalf("the seconds function fail")
	}
}

func TestGenRequestKey(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("http://127.0.0.1:5018/users/me")
	key := string(GenRequestKey(ctx))
	if key != "GEThttp://127.0.0.1:5018/users/me" {
		t.Fatalf("gen request key fail, %q", key)
	}
}

func TestGzip(t *testing.T) {
	data := []byte("hello world")
	buf, err := Gzip(data)
	if err != nil {
		t.Fatalf("do gzip fail, %v", err)
	}
	buf, err = Gunzip(buf)
	if err != nil {
		t.Fatalf("do gunzip fail, %v", err)
	}
	if string(buf) != string(data) {
		t.Fatalf("do gunzip fail")
	}
}
