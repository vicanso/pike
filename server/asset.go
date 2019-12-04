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
	"bytes"
	"io"
	"os"

	"github.com/gobuffalo/packr/v2"
)

var (
	box = packr.New("asset", "../web/build")
)

type (
	assetFiles struct{}
)

func (*assetFiles) Exists(file string) bool {
	return box.Has(file)
}
func (*assetFiles) Get(file string) ([]byte, error) {
	return box.Find(file)
}
func (*assetFiles) Stat(file string) os.FileInfo {
	return nil
}
func (af *assetFiles) NewReader(file string) (io.Reader, error) {
	buf, err := af.Get(file)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buf), nil
}
