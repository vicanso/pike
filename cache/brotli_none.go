// +build !brotli

package cache

import "errors"

// doBrotli brotli compress
func doBrotli(buf []byte) ([]byte, error) {
	return nil, errors.New("not support brotli")
}
