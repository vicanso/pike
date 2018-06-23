package util

import (
	"net/http"
	"testing"
	"time"
)

func TestGzip(t *testing.T) {
	data := []byte("hello world")
	buf, err := Gzip(data, 0)
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

func TestBrotli(t *testing.T) {
	data := []byte("hello world")
	buf, err := BrotliEncode(data, 0)
	if err != nil {
		t.Fatalf("do brotli fail, %v", err)
	}
	originalBuf, err := BrotliDecode(buf)
	if err != nil {
		t.Fatalf("do brotli decode fail, %v", err)
	}
	if string(originalBuf) != string(data) {
		t.Fatalf("do brotli decode fail")
	}
}

func TestGetHeaderValue(t *testing.T) {
	header := http.Header{
		"eTag": []string{
			"ABCD",
		},
		"X-Forward-For": []string{
			"127.0.0.1",
		},
	}
	value := GetHeaderValue(header, "ETag")
	if len(value) != 1 || value[0] != "ABCD" {
		t.Fatalf("get header value fail")
	}

	value = GetHeaderValue(header, "Token")
	if len(value) != 0 {
		t.Fatalf("get the not exists header should return empty string")
	}
}

func TestGetTimeConsuming(t *testing.T) {
	start := time.Now()
	time.Sleep(10 * time.Millisecond)
	ms := GetTimeConsuming(start)
	if ms == 0 {
		t.Fatalf("get time consuming fail")
	}
}

func TestGetHumanReadableSize(t *testing.T) {
	if GetHumanReadableSize(1024*1024) != "1MB" {
		t.Fatalf("1024 * 1024 should be 1MB")
	}
	if GetHumanReadableSize(1024*1024+500*1024) != "1.49MB" {
		t.Fatalf("1024*1024+500*1024 should be 1.49MB")
	}

	if GetHumanReadableSize(1024) != "1KB" {
		t.Fatalf("1024 should be 1KB")
	}
	if GetHumanReadableSize(1024+500) != "1.49KB" {
		t.Fatalf("1024+500 should be 1.49KB")
	}
	if GetHumanReadableSize(500) != "500B" {
		t.Fatalf("500 should be 500B")
	}
}

func TestGetRewriteRegexp(t *testing.T) {
	result := GetRewriteRegexp([]string{
		"/api",
		"/api/*:/$1",
	})
	if len(result) != 1 {
		t.Fatalf("rewrite exgexp fail")
	}
	for reg := range result {
		groups := reg.FindAllStringSubmatch("/api/users/me", -1)
		if groups[0][1] != "users/me" {
		}
	}
}

func TestGetIdentity(t *testing.T) {
	req := &http.Request{
		Method:     "GET",
		Host:       "aslant.site",
		RequestURI: "/users/me",
	}
	id := GetIdentity(req)
	if string(id) != "GET aslant.site /users/me" {
		t.Fatalf("get identity fail")
	}
}
