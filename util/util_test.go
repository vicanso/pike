package util

import (
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
	buf, err := Brotli(data, 0)
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
