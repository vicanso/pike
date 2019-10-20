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
	"encoding/hex"
	"testing"

	"github.com/minio/highwayhash"
)

func BenchmarkSha256(b *testing.B) {
	data := []byte("GET tiny.aslant.site /users/v1/login-token?type=vip")
	for i := 0; i < b.N; i++ {
		h := sha256.New()
		h.Write(data)
		h.Sum(nil)
	}
}

func BenchmarkHighwayHash(b *testing.B) {
	data := []byte("GET tiny.aslant.site /users/v1/login-token?type=vip")

	key, _ := hex.DecodeString("000102030405060708090A0B0C0D0E0FF0E0D0C0B0A090807060504030201000") // use your own key here

	for i := 0; i < b.N; i++ {
		highwayhash.Sum64(data, key)
	}
}
