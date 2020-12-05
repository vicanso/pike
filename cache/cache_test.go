// MIT License

// Copyright (c) 2020 Tree Xie

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vicanso/pike/config"
)

func TestConvertConfig(t *testing.T) {
	assert := assert.New(t)

	name := "cache-test"
	size := 100
	hitForPass := 60
	configs := []config.CacheConfig{
		{
			Name:       name,
			Size:       size,
			HitForPass: "1m",
		},
	}
	opts := convertConfigs(configs)
	assert.Equal(1, len(opts))
	assert.Equal(name, opts[0].Name)
	assert.Equal(size, opts[0].Size)
	assert.Equal(hitForPass, opts[0].HitForPass)
}

func TestDefaultDispatcher(t *testing.T) {
	assert := assert.New(t)
	name := "test"
	assert.Nil(GetDispatcher(name))
	ResetDispatchers([]config.CacheConfig{
		{
			Name:       name,
			Size:       100,
			HitForPass: "1m",
		},
	})
	assert.NotNil(GetDispatcher(name))
}
