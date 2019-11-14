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
	"net/http/httptest"
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

func TestGetIdentity(t *testing.T) {
	assert := assert.New(t)
	req := httptest.NewRequest("GET", "/users/v1/me?type=vip", nil)
	req.Host = "aslant.site"
	assert.Equal("GET aslant.site /users/v1/me?type=vip", string(GetIdentity(req)))
}

func TestGenerateETag(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(`"0-2jmj7l5rSw0yVb_vlWAYkK_YBwk="`, GenerateETag(nil))
	assert.Equal(`"4-gf6L_odXbD7LIkJvjleEc4KRes8="`, GenerateETag([]byte("abcd")))
}
