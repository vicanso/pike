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

package compress

import (
	"github.com/pierrec/lz4"
)

func doLZ4Encode(data []byte, level int) ([]byte, error) {
	buf := make([]byte, len(data))
	n, err := lz4.CompressBlock(data, buf, nil)
	if err != nil {
		return nil, err
	}
	buf = buf[:n]
	return buf, nil
}

func doLZ4Decode(buf []byte) ([]byte, error) {
	dst := make([]byte, 10*len(buf))
	n, err := lz4.UncompressBlock(buf, dst)
	if err != nil {
		return nil, err
	}
	dst = dst[:n]
	return dst, nil
}
