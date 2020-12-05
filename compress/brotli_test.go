package compress

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoBrotli(t *testing.T) {
	assert := assert.New(t)
	// 压缩级别超出则与默认压缩级别一样
	tests := []struct {
		data       []byte
		level      int
		resultSize int
	}{
		{
			data:       compressTestData,
			level:      0,
			resultSize: 589,
		},
		{
			data:       compressTestData,
			level:      12,
			resultSize: 589,
		},
		{
			data:       compressTestData,
			level:      8,
			resultSize: 592,
		},
	}
	for _, tt := range tests {
		data, err := doBrotli(tt.data, tt.level)
		assert.Nil(err)
		assert.Equal(tt.resultSize, len(data))
		assert.NotEqual(tt.data, data)
	}
}

func TestDoBrotliDecode(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		data []byte
	}{
		{
			data: compressTestData,
		},
	}
	for _, tt := range tests {
		data, err := doBrotli(tt.data, 0)
		assert.Nil(err)
		assert.NotNil(data)
		assert.NotEqual(tt.data, data)
		data, err = doBrotliDecode(data)
		assert.Nil(err)
		assert.Equal(tt.data, data)
	}
}
