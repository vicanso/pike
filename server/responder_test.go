// MIT License

// Copyright (c) 2020 Tree Xie

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package server

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vicanso/elton"
	"github.com/vicanso/pike/cache"
)

func TestResponderMiddleware(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		create    func() *elton.Context
		headerAge string
		body      []byte
	}{
		{
			create: func() *elton.Context {
				c := elton.NewContext(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
				return c
			},
			headerAge: "",
		},
		{
			create: func() *elton.Context {
				c := elton.NewContext(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
				setHTTPResp(c, &cache.HTTPResponse{
					RawBody: []byte("abcd"),
				})
				setHTTPRespAge(c, 10)
				return c
			},
			headerAge: "10",
			body:      []byte("abcd"),
		},
	}
	fn := NewResponder()
	for _, tt := range tests {
		c := tt.create()
		c.Next = func() error {
			return nil
		}
		err := fn(c)
		if err != nil {
			assert.Equal(ErrInvalidResponse, err)
		} else {
			assert.Equal(tt.headerAge, c.GetHeader("Age"))
			assert.Equal(tt.body, c.BodyBuffer.Bytes())
		}
	}
}
