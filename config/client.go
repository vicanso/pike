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

// The client to read data of config,
// such as etcd and file.

package config

type (
	// Client client interface
	Client interface {
		// Get get the data of key
		Get(key string) (data []byte, err error)
		// Set set the data of key
		Set(key string, data []byte) (err error)
		// Delete delete the data of key
		Delete(key string) (err error)
		// List list all sub keys of key
		List(prefix string) (keys []string, err error)
		// Watch watch key change with prefix
		Watch(key string, onChange OnKeyChange)
		// Close
		Close() error
	}
	// OnKeyChange config change's event handler
	OnKeyChange func(string)
)
