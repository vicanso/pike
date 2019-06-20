// +build brotli

package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBrotliCompress(t *testing.T) {
	assert := assert.New(t)
	buf, err := doBrotli([]byte("abcd"), 0)
	assert.Nil(err, "brotli compress fail")
	assert.NotEqual(0, len(buf), "brotli compress fail")
}
