package util

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
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

func TestGetDebugVars(t *testing.T) {
	buf := GetDebugVars()
	if len(buf) < 10 {
		t.Fatalf("get the debug vars fail, %v", string(buf))
	}
}

func TestGetTimeConsuming(t *testing.T) {
	startedAt := time.Now()
	time.Sleep(2 * time.Millisecond)
	ms := GetTimeConsuming(startedAt)
	if ms <= 0 {
		t.Fatalf("the time consuming should be gt 0")
	}
}

func TestTimingConsumingHeader(t *testing.T) {
	header := &fasthttp.RequestHeader{}
	startedAt := time.Now()
	time.Sleep(2 * time.Millisecond)
	key := []byte("Consuming")
	SetTimingConsumingHeader(startedAt, header, key)
}

func TestCompressJPEG(t *testing.T) {
	buf, _ := ioutil.ReadFile("../assets/images/mac.jpg")
	newBuf, _ := CompressJPEG(buf, 70)
	if len(newBuf) >= len(buf) {
		t.Fatalf("compress jpeg fail")
	}
}
