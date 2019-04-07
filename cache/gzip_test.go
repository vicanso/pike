package cache

import (
	"bytes"
	"testing"
)

func TestGzip(t *testing.T) {
	buf := []byte("abcd")
	gzipBuf, err := Gzip(buf)
	if err != nil || len(gzipBuf) == 0 {
		t.Fatalf("gzip data fail, %v", err)
	}

	gunzipBuf, err := Gunzip(gzipBuf)
	if err != nil || !bytes.Equal(gunzipBuf, buf) {
		t.Fatalf("gunzip data fail, %v", err)
	}
}
