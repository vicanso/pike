// +build brotli

package cache

import (
	"github.com/google/brotli/go/cbrotli"
)

const (
	maxQuality = 11
)

func doBrotli(buf []byte) ([]byte, error) {
	level := compressLevel
	if level == 0 {
		level = 9
	}
	if level > maxQuality {
		level = maxQuality
	}
	return cbrotli.Encode(buf, cbrotli.WriterOptions{
		Quality: level,
		LGWin:   0,
	})
}
