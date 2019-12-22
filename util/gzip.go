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

// gzip compress

package util

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
)

// Gunzip gunzip
func Gunzip(buf []byte) ([]byte, error) {
	if len(buf) == 0 {
		return nil, nil
	}
	r, err := gzip.NewReader(bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return ioutil.ReadAll(r)
}

func doGzip(buf []byte, level int) (*bytes.Buffer, error) {
	buffer := new(bytes.Buffer)
	if level <= 0 || level > gzip.BestCompression {
		level = gzip.DefaultCompression
	}
	// 已处理了压缩级别范围，因此不会出错1
	w, _ := gzip.NewWriterLevel(buffer, level)
	defer w.Close()
	_, err := w.Write(buf)
	if err != nil {
		return nil, err
	}
	return buffer, nil
}

// Gzip gzip function
func Gzip(buf []byte, level int) ([]byte, error) {
	buffer, err := doGzip(buf, level)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
