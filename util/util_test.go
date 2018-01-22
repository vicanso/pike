package util

import (
	"testing"

	"github.com/valyala/fasthttp"
)

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

func TestGetClientIP(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	ip := GetClientIP(ctx)
	if ip != "0.0.0.0" {
		t.Fatalf("the client ip excpect 0.0.0.0 but %v", ip)
	}
	ctx.Request.Header.SetCanonical([]byte("X-Forwarded-For"), []byte("4.4.4.4,8.8.8.8"))
	ip = GetClientIP(ctx)
	if ip != "4.4.4.4" {
		t.Fatalf("the client ip excpect 4.4.4.4 but %v", ip)
	}
}

func TestShouldCompress(t *testing.T) {
	if ShouldCompress([]byte("text/css; charset=UTF-8")) != true {
		t.Fatalf("the css should be compress")
	}

	if ShouldCompress([]byte("application/javascript; charset=UTF-8")) != true {
		t.Fatalf("the js should be compress")
	}

	if ShouldCompress([]byte("image/png")) != false {
		t.Fatalf("the image shouldn't be compress")
	}

}

func TestTrimHeader(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	header := &ctx.Request.Header
	header.SetCanonical([]byte("Content-Type"), []byte("application/json; charset=utf-8"))
	header.SetCanonical([]byte("X-Response-Id"), []byte("BJJRAyf4f"))
	header.SetCanonical([]byte("Cache-Control"), []byte("no-cache, max-age=0"))
	header.SetCanonical([]byte("Connection"), []byte("keep-alive"))
	header.SetCanonical([]byte("Date"), []byte("Tue, 09 Jan 2018 12:27:02 GMT"))
	str := "User-Agent: fasthttp\r\nContent-Type: application/json; charset=utf-8\r\nX-Response-Id: BJJRAyf4f\r\nCache-Control: no-cache, max-age=0"
	data := string(TrimHeader(header.Header()))
	if data != str {
		t.Fatalf("trim header fail expect %v but %v", str, data)
	}
}

func TestGetDebugVars(t *testing.T) {
	buf := GetDebugVars()
	if len(buf) < 10 {
		t.Fatalf("get the debug vars fail, %v", string(buf))
	}
}

func TestGetEtag(t *testing.T) {
	eTag := GetETag([]byte(""))
	if eTag != "\"0-2jmj7l5rSw0yVb/vlWAYkK/YBwk\"" {
		t.Fatalf("get empty data etag fail")
	}
	buf := []byte("测试使用的响应数据")
	eTag = GetETag(buf)
	if eTag != "\"1b-gQEzXLxF7NjFZ-x0-GK1Pg8NBZA=\"" {
		t.Fatalf("get etag fail")
	}
}
