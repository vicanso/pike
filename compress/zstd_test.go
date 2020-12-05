package compress

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoZSTDDecode(t *testing.T) {
	assert := assert.New(t)
	data := compressTestData
	dst, err := doZSTDEncode(data, 1)
	assert.Nil(err)
	assert.NotNil(dst)
	assert.NotEqual(data, dst)

	buf, err := doZSTDDecode(dst)
	assert.Nil(err)
	assert.Equal(data, buf)
}
