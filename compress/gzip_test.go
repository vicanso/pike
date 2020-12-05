package compress

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoGzip(t *testing.T) {
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
			resultSize: 660,
		},
		{
			data:       compressTestData,
			level:      0,
			resultSize: 660,
		},
		{
			data:       compressTestData,
			level:      0,
			resultSize: 660,
		},
	}
	for _, tt := range tests {
		data, err := doGzip(tt.data, tt.level)
		assert.Nil(err)
		assert.NotEqual(tt.data, data)
		assert.Equal(tt.resultSize, len(data))
	}
}

func TestDoGunzip(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		data []byte
	}{
		{
			data: compressTestData,
		},
	}
	for _, tt := range tests {
		data, err := doGzip(tt.data, 0)
		assert.Nil(err)
		assert.NotEmpty(tt.data, data)
		assert.NotNil(data)
		data, err = doGunzip(data)
		assert.Nil(err)
		assert.Equal(tt.data, data)
	}
}
