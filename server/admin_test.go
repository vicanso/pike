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
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vicanso/elton"
	"github.com/vicanso/hes"
	"github.com/vicanso/pike/config"
)

func TestNewAdminValidateMiddlewares(t *testing.T) {
	assert := assert.New(t)
	adminConfig := &config.Admin{
		EnabledInternetAccess: false,
		User:                  "tree.xie",
		Password:              "password",
	}
	handlers := newAdminValidateMiddlewares(adminConfig)
	assert.Equal(2, len(handlers))

	t.Run("intranet access", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(elton.HeaderXRealIP, "1.1.1.1")
		c := elton.NewContext(nil, req)
		c.Next = func() error {
			return nil
		}
		err := handlers[0](c)
		assert.NotNil(err)
		he, ok := err.(*hes.Error)
		assert.True(ok)
		assert.Equal(403, he.StatusCode)

		req.Header.Del(elton.HeaderXRealIP)
		req.RemoteAddr = "192.168.1.1:3000"
		c = elton.NewContext(nil, req)
		done := false
		c.Next = func() error {
			done = true
			return nil
		}
		err = handlers[0](c)
		assert.Nil(err)
		assert.True(done)
	})

	t.Run("basic auth", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		resp := httptest.NewRecorder()
		c := elton.NewContext(resp, req)
		c.Next = func() error {
			return nil
		}
		err := handlers[1](c)
		assert.NotNil(err)
		he, ok := err.(*hes.Error)
		assert.True(ok)
		assert.Equal(http.StatusUnauthorized, he.StatusCode)

		// 错误的账号或密码
		req.Header.Set("Authorization", "basic dHJlZS54aWU6cGFzcw==")
		err = handlers[1](c)
		assert.NotNil(err)
		he, ok = err.(*hes.Error)
		assert.True(ok)
		assert.Equal(http.StatusUnauthorized, he.StatusCode)

		// 正确的账号密码
		req.Header.Set("Authorization", "basic dHJlZS54aWU6cGFzc3dvcmQ=")
		done := false
		c.Next = func() error {
			done = true
			return nil
		}
		err = handlers[1](c)
		assert.Nil(err)
		assert.True(done)
	})
}
