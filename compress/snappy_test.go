package compress

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoSnappyDecode(t *testing.T) {
	assert := assert.New(t)
	data := compressTestData
	dst := doSnappyEncode(data)
	assert.NotNil(dst)
	assert.NotEqual(data, dst)

	buf, err := doSnappyDecode(dst)
	assert.Nil(err)
	assert.Equal(data, buf)
}
