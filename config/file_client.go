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

// file client for config

package config

import (
	"io/ioutil"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/vicanso/pike/log"
	"go.uber.org/zap"
)

// fileClient file client
type fileClient struct {
	file    string
	watcher *fsnotify.Watcher
}

const defaultPerm os.FileMode = 0600

// NewFileClient create a new file client
func NewFileClient(file string) (client *fileClient, err error) {
	f, err := os.OpenFile(file, os.O_RDONLY|os.O_CREATE, defaultPerm)
	if err != nil {
		return
	}
	defer f.Close()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}

	client = &fileClient{
		file:    file,
		watcher: watcher,
	}
	return
}

// Get get data from file
func (fc *fileClient) Get() (data []byte, err error) {
	return ioutil.ReadFile(fc.file)
}

// Set set data to file
func (fc *fileClient) Set(data []byte) (err error) {
	return ioutil.WriteFile(fc.file, data, defaultPerm)
}

// Watch watch config change
func (fc *fileClient) Watch(onChange OnChange) {

	err := fc.watcher.Add(fc.file)
	if err != nil {
		log.Default().Error("add watch fail",
			zap.String("file", fc.file),
			zap.Error(err),
		)
		return
	}
	for {
		select {
		case event, ok := <-fc.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				onChange()
			}
		case err, ok := <-fc.watcher.Errors:
			if !ok {
				return
			}
			if err != nil {
				log.Default().Error("watch error",
					zap.String("file", fc.file),
					zap.Error(err),
				)
			}
		}
	}
}

// Close close file client
func (fc *fileClient) Close() error {
	return fc.watcher.Close()
}
