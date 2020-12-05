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
	"strconv"

	"github.com/vicanso/elton"
)

// NewResponder create a responder middleware
func NewResponder() elton.Handler {
	return func(c *elton.Context) (err error) {
		err = c.Next()
		if err != nil {
			return
		}
		// 从context中读取http response，该数据由cache中间件设置或proxy中间件设置
		httpResp := getHTTPResp(c)
		if httpResp == nil {
			err = ErrInvalidResponse
			return
		}
		err = httpResp.Fill(c)
		if err != nil {
			return
		}

		// http 响应头放在最后可以覆盖proxy的设置的相同响应头
		// 获取该响应的age，只有从缓存中读取的数据才有age，由cache中间件设置
		age := getHTTPRespAge(c)
		if age > 0 {
			c.SetHeader(headerAge, strconv.Itoa(age))
		}

		c.SetHeader(headerCacheStatus, getCacheStatus(c).String())
		return
	}
}
