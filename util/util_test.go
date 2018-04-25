package util

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/google/brotli/go/cbrotli"
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

func TestCompressJPEG(t *testing.T) {
	buf, _ := ioutil.ReadFile("../assets/images/mac.jpg")
	newBuf, _ := CompressJPEG(buf, 70)
	if len(newBuf) >= len(buf) {
		t.Fatalf("compress jpeg fail")
	}
	log.Printf("original: %d byte, compress: %d byte", len(buf), len(newBuf))
}
