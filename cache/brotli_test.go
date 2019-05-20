// +build brotli

package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBrotliCompress(t *testing.T) {
	assert := assert.New(t)
	buf, err := doBrotli([]byte("abcd"))
	assert.Nil(err, "brotli compress fail")
	assert.NotEqual(len(buf), 0, "brotli compress fail")
}
