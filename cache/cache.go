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

package cache

import (
	"time"
	"unsafe"

	"github.com/vicanso/pike/config"
)

// byteSliceToString converts a []byte to string without a heap allocation.
func byteSliceToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

var defaultDispatchers = NewDispatchers(nil)

// GetDispatcher get dispatcher form default dispatchers
func GetDispatcher(name string) *dispatcher {
	return defaultDispatchers.Get(name)
}

// RemoveHTTPCache remove http cache form default dispatchers
func RemoveHTTPCache(name string, key []byte) {
	defaultDispatchers.RemoveHTTPCache(name, key)
}

func convertConfigs(configs []config.CacheConfig) []DispatcherOption {
	opts := make([]DispatcherOption, 0)
	for _, item := range configs {
		d, _ := time.ParseDuration(item.HitForPass)
		opts = append(opts, DispatcherOption{
			Name:       item.Name,
			Size:       item.Size,
			HitForPass: int(d.Seconds()),
		})
	}
	return opts
}

// ResetDispatchers reset default dispatchers
func ResetDispatchers(configs []config.CacheConfig) {
	defaultDispatchers.Reset(convertConfigs(configs))
}
