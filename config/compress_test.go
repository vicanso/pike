// Copyright 2019 tree xie
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompressConfig(t *testing.T) {
	assert := assert.New(t)
	c := &Compress{
		Name: "tiny",
	}
	defer func() {
		_ = c.Delete()
	}()

	err := c.Fetch()
	assert.Nil(err)
	assert.Empty(c.Level)
	assert.Empty(c.MinLength)
	assert.Empty(c.Filter)

	level := 9
	minLength := 1024
	filter := "abcd"
	compressDescription := "compress description"
	c.Level = level
	c.MinLength = minLength
	c.Filter = filter
	c.Description = compressDescription
	err = c.Save()
	assert.Nil(err)

	nc := &Compress{
		Name: c.Name,
	}
	err = nc.Fetch()
	assert.Nil(err)
	assert.Equal(level, nc.Level)
	assert.Equal(minLength, nc.MinLength)
	assert.Equal(filter, c.Filter)
	assert.Equal(compressDescription, c.Description)

	compresses, err := GetCompresses()
	assert.Nil(err)
	assert.Equal(1, len(compresses))

	nc = compresses.Get(c.Name)
	assert.Equal(c, nc)
}
