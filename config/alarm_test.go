// Copyright 2020 tree xie
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

func TestAlarmConfig(t *testing.T) {
	assert := assert.New(t)
	cfg := NewTestConfig()
	a := &Alarm{
		Name: "upstream",
		cfg:  cfg,
	}
	defer func() {
		_ = a.Delete()
	}()

	err := a.Fetch()
	assert.Nil(err)
	assert.Empty(a.Template)

	a.Template = `{
		"data": ""
	}`
	a.URI = "http://127.0.0.1"
	err = a.Save()
	assert.Nil(err)

	na := &Alarm{
		Name: a.Name,
		cfg:  cfg,
	}
	err = na.Fetch()
	assert.Nil(err)
	assert.Equal(a.Template, na.Template)
	assert.Equal(a.URI, na.URI)

	alarms, err := cfg.GetAlarms()
	assert.Nil(err)
	na = alarms.Get(a.Name)
	assert.Equal(a, na)

}
