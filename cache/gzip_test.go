package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGzip(t *testing.T) {
	assert := assert.New(t)
	buf := []byte("abcd")
	gzipBuf, err := Gzip(buf)
	assert.Nil(err)
	assert.NotEqual(len(gzipBuf), 0)

	gunzipBuf, err := Gunzip(gzipBuf)
	assert.Nil(err)
	assert.Equal(gunzipBuf, buf)
}
