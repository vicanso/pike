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

// brotli compress

package util

import (
	"bytes"
	"io/ioutil"

	"github.com/andybalholm/brotli"
)

const (
	defaultBrQuality = 8
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

// Brotli brotli compress
func Brotli(buf []byte, level int) ([]byte, error) {
	buffer, err := brotliEncode(buf, level)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// BrotliDecode brotli decode
func BrotliDecode(buf []byte) ([]byte, error) {
	if len(buf) == 0 {
		return nil, nil
	}
	r := brotli.NewReader(bytes.NewBuffer(buf))
	return ioutil.ReadAll(r)
}
