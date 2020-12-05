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
	"compress/gzip"
	"io/ioutil"
)

// doGunzip gunzip
func doGunzip(buf []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return ioutil.ReadAll(r)
}

func gzipFn(buf []byte, level int) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)
	if level <= 0 || level > gzip.BestCompression {
		level = gzip.DefaultCompression
	}
	w, err := gzip.NewWriterLevel(buffer, level)
	if err != nil {
		return nil, err
	}
	defer w.Close()
	_, err = w.Write(buf)
	if err != nil {
		return nil, err
	}
	return buffer, nil
}

// doGzip gzip function
func doGzip(buf []byte, level int) ([]byte, error) {
	buffer, err := gzipFn(buf, level)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
