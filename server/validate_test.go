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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vicanso/pike/config"
)

func TestDoValidate(t *testing.T) {
	assert := assert.New(t)
	err := doValidate(new(config.Server), []byte(`{
		"name": "test",
		"cache": "commonCache",
		"compress": "commonCompress",
		"locations": ["l1", "l2"],
		"addr": ":3000"
	}`))
	assert.Nil(err)

	err = doValidate(new(config.Location), []byte(`{
		"name": "l1",
		"upstream": "u1",
		"prefixs": ["/api"],
		"rewrites": ["/api/*:/$1"],
		"hosts": ["aslant.site"],
		"responseHeader": ["X-Response-ID:123"]
	}`))
	assert.Nil(err)

	err = doValidate(new(config.Upstream), []byte(`{
		"name": "u1",
		"healthCheck": "/ping",
		"servers": [
			{
				"addr": "127.0.0.1:3000"
			}
		]
	}`))
	assert.Nil(err)

	err = doValidate(new(config.Admin), map[string]string{
		"prefix": "/pike",
	})
	assert.Nil(err)
}
