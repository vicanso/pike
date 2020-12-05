package compress

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoLZ4Decode(t *testing.T) {
	assert := assert.New(t)
	data := compressTestData
	result, err := doLZ4Encode(data, 0)
	assert.Nil(err)
	assert.NotNil(result)
	assert.NotEqual(data, result)

	result, err = doLZ4Decode(result)
	assert.Nil(err)
	assert.Equal(data, result)
}
