// +build !brotli

package cache

import (
	"errors"
)

// doBrotli brotli compress
func doBrotli(buf []byte, level int) ([]byte, error) {
	return nil, errors.New("not support brotli")
}
