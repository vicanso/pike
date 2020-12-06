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

package util

import (
	"sync"

	"github.com/vicanso/hes"
)

const errCategory = "pike"

type DeleteMatch func(string) bool

// MapDelete delete item form sync map
func MapDelete(m *sync.Map, match DeleteMatch) []interface{} {
	result := make([]interface{}, 0)
	// m.Range中是先复制了仅读，因此可以直接在此处删除
	m.Range(func(k, _ interface{}) bool {
		key, ok := k.(string)
		if !ok {
			return true
		}
		// 如果不匹配，则不需要删除
		if !match(key) {
			return true
		}

		value, loaded := m.LoadAndDelete(key)
		if loaded {
			result = append(result, value)
		}
		return true
	})

	return result
}

// NewError create a new http error
func NewError(message string, statusCode int) error {
	return &hes.Error{
		Message:    message,
		StatusCode: statusCode,
		Category:   errCategory,
	}
}
