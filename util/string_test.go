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

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomString(t *testing.T) {
	assert := assert.New(t)
	size := 10
	assert.Equal(size, len(RandomString(size)))
	assert.NotEqual(RandomString(size), RandomString(size))
}

func TestByteSliceToString(t *testing.T) {
	assert := assert.New(t)
	assert.Equal("abcd", ByteSliceToString([]byte("abcd")))
}
