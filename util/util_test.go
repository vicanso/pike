package util

import (
	"io/ioutil"
	"log"
	"testing"
	"time"

	"github.com/google/brotli/go/cbrotli"
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

func TestBrotli(t *testing.T) {
	data := []byte("hello world")
	buf, err := Brotli(data, 0)
	if err != nil {
		t.Fatalf("do brotli fail, %v", err)
	}
	buf, err = cbrotli.Decode(buf)
	if err != nil {
		t.Fatalf("do brotli decode fail, %v", err)
	}
	if string(buf) != string(data) {
		t.Fatalf("do brotli fail")
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
	log.Printf("original: %d byte, compress: %d byte", len(buf), len(newBuf))
}

func TestCompressPNG(t *testing.T) {
	buf, _ := ioutil.ReadFile("../assets/images/icon.png")
	newBuf, _ := CompressPNG(buf)
	if len(newBuf) >= len(buf) {
		t.Fatalf("compress png fail")
	}
	log.Printf("original: %d byte, compress: %d byte", len(buf), len(newBuf))
}
