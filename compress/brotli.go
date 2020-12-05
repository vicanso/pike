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
	"bytes"
	"io/ioutil"

	"github.com/andybalholm/brotli"
)

const (
	defaultBrQuality = 6
)

func brotliEncode(buf []byte, level int) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)
	if level <= 0 || level > 11 {
		level = defaultBrQuality
	}
	w := brotli.NewWriterLevel(buffer, level)
	defer w.Close()
	_, err := w.Write(buf)
	if err != nil {
		return nil, err
	}
	return buffer, nil
}

// doBrotli brotli compress
func doBrotli(buf []byte, level int) ([]byte, error) {
	buffer, err := brotliEncode(buf, level)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// doBrotliDecode brotli decode
func doBrotliDecode(buf []byte) ([]byte, error) {
	if len(buf) == 0 {
		return nil, nil
	}
	r := brotli.NewReader(bytes.NewBuffer(buf))
	return ioutil.ReadAll(r)
}
