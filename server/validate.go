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
	"encoding/json"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/vicanso/hes"
	"github.com/vicanso/pike/config"
)

var (
	customTypeTagMap = govalidator.CustomTypeTagMap
)

func init() {
	govalidator.SetFieldsRequiredByDefault(true)

	add("xName", func(i interface{}, _ interface{}) bool {
		name, ok := i.(string)
		if !ok {
			return false
		}
		return isConfigName(name)
	})
	add("xNames", func(i interface{}, _ interface{}) bool {
		arr, ok := i.([]string)
		if !ok {
			return false
		}
		for _, item := range arr {
			if !isConfigName(item) {
				return false
			}
		}
		return true
	})
	add("xPrefixs", func(i interface{}, _ interface{}) bool {
		arr, ok := i.([]string)
		if !ok {
			return false
		}
		for _, item := range arr {
			if !isURLPath(item) {
				return false
			}
		}
		return true
	})
	add("xRewrites", func(i interface{}, _ interface{}) bool {
		arr, ok := i.([]string)
		if !ok {
			return false
		}
		for _, item := range arr {
			if item == "" || !strings.ContainsRune(item, ':') {
				return false
			}
		}
		return true
	})
	add("xHosts", func(i interface{}, _ interface{}) bool {
		arr, ok := i.([]string)
		if !ok {
			return false
		}
		for _, item := range arr {
			if !govalidator.IsHost(item) {
				return false
			}
		}
		return true
	})
	add("xHeader", func(i interface{}, _ interface{}) bool {
		arr, ok := i.([]string)
		if !ok {
			return false
		}
		for _, item := range arr {
			if item == "" || !strings.ContainsRune(item, ':') {
				return false
			}
		}
		return true
	})
	add("xURLPath", func(i interface{}, _ interface{}) bool {
		value, ok := i.(string)
		if !ok {
			return false
		}
		return isURLPath(value)
	})

	add("xServers", func(i interface{}, _ interface{}) bool {
		_, ok := i.([]config.UpstreamServer)
		return ok
	})
}

func isURLPath(value string) bool {
	if value == "" || value[0] != '/' {
		return false
	}
	return true
}

func isConfigName(value string) bool {
	if value == "" {
		return false
	}
	if !govalidator.IsASCII(value) || !govalidator.RuneLength(value, "1", "20") {
		return false
	}
	return true
}

func doValidate(s interface{}, data interface{}) (err error) {
	// statusCode := http.StatusBadRequest
	if data != nil {
		switch data := data.(type) {
		case []byte:
			err = json.Unmarshal(data, s)
			if err != nil {
				he := hes.Wrap(err)
				err = he
				return
			}
		default:
			buf, err := json.Marshal(data)
			if err != nil {
				return err
			}
			err = json.Unmarshal(buf, s)
			if err != nil {
				return err
			}
		}
	}
	_, err = govalidator.ValidateStruct(s)
	return
}

// add add validate
func add(name string, fn govalidator.CustomTypeValidator) {
	_, exists := customTypeTagMap.Get(name)
	if exists {
		panic(name + " is duplicated")
	}
	customTypeTagMap.Set(name, fn)
}
