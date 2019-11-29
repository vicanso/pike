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

package cache

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vicanso/pike/config"
)

func BenchmarkSha256(b *testing.B) {
	data := []byte("GET tiny.aslant.site /users/v1/login-token?type=vip")
	for i := 0; i < b.N; i++ {
		h := sha256.New()
		h.Write(data)
		h.Sum(nil)
	}
}

func BenchmarkMemhash(b *testing.B) {
	data := []byte("GET tiny.aslant.site /users/v1/login-token?type=vip")
	for i := 0; i < b.N; i++ {
		MemHash(data)
	}
}

func TestDispatcher(t *testing.T) {
	assert := assert.New(t)
	name := "test"
	cachesConfig := config.Caches{
		&config.Cache{
			Name: name,
		},
	}
	dispatchers := NewDispatchers(cachesConfig)
	disp := dispatchers.Get(name)
	assert.NotNil(disp)

	key := []byte("abcd")
	c1 := disp.GetHTTPCache(key)
	c2 := disp.GetHTTPCache(key)
	assert.Equal(c1, c2)
}
