// +build brotli

package cache

import (
	"testing"
)

func TestBrotliCompress(t *testing.T) {
	buf, err := doBrotli([]byte("abcd"))
	if err != nil || len(buf) == 0 {
		t.Fatalf("br fail, %v", err)
	}
}
