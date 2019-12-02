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

func TestAdminConfig(t *testing.T) {
	assert := assert.New(t)
	defer func() {
		new(Admin).Delete()
	}()
	admin, err := GetAdmin()
	assert.Nil(err)
	assert.Equal(defaultAdminPrefix, admin.Prefix)
	assert.Empty(admin.User)
	assert.Empty(admin.Password)

	user := "foo"
	password := "bar"
	prefix := "/user-admin"
	description := "admin description"
	admin.User = user
	admin.Prefix = prefix
	admin.Password = password
	admin.Description = description
	err = admin.Save()
	assert.Nil(err)

	admin = new(Admin)
	err = admin.Fetch()
	assert.Nil(err)
	assert.Equal(user, admin.User)
	assert.Equal(password, admin.Password)
	assert.Equal(prefix, admin.Prefix)
	assert.Equal(description, admin.Description)
}
