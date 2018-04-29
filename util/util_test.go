package util

import (
	"net/http"
	"testing"
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
