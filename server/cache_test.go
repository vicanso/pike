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

package server

import (
	"net/http"
	"testing"

	"github.com/vicanso/elton"

	"github.com/stretchr/testify/assert"
)

func TestRequestIsPass(t *testing.T) {
	assert := assert.New(t)
	req := &http.Request{
		Method: "GET",
	}
	assert.False(requestIsPass(req))

	req.Method = "HEAD"
	assert.False(requestIsPass(req))

	req.Method = "POST"
	assert.True(requestIsPass(req))
}

func TestGetCacheAge(t *testing.T) {
	assert := assert.New(t)
	h := make(http.Header)

	h.Set(elton.HeaderSetCookie, "abc")
	assert.Equal(0, getCacheAge(h))
	h.Del(elton.HeaderSetCookie)

	assert.Equal(0, getCacheAge(h))

	h.Set(elton.HeaderCacheControl, "no-cache")
	assert.Equal(0, getCacheAge(h))

	h.Set(elton.HeaderCacheControl, "public, max-age=10, s-maxage=2")
	assert.Equal(2, getCacheAge(h))

	h.Set(elton.HeaderCacheControl, "public, max-age=10")
	assert.Equal(10, getCacheAge(h))

	h.Set(headerAge, "2")
	assert.Equal(8, getCacheAge(h))
}
