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

func TestCacheConfig(t *testing.T) {
	assert := assert.New(t)
	c := &Cache{
		Name: "tiny",
	}
	defer c.Delete()

	err := c.Fetch()
	assert.Nil(err)
	assert.Empty(c.HitForPass)
	assert.Empty(c.Zone)
	assert.Empty(c.Size)

	hitForPass := 300
	zone := 1
	size := 10
	c.HitForPass = hitForPass
	c.Zone = zone
	c.Size = size
	err = c.Save()
	assert.Nil(err)

	nc := &Cache{
		Name: c.Name,
	}
	err = nc.Fetch()
	assert.Nil(err)
	assert.Equal(hitForPass, nc.HitForPass)
	assert.Equal(zone, nc.Zone)
	assert.Equal(size, nc.Size)

	caches, err := GetCaches()
	assert.Nil(err)
	nc = caches.Get(c.Name)
	assert.Equal(c, nc)
}
