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

package config

import (
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/go-playground/validator/v10"
	us "github.com/vicanso/upstream"
)

var defaultValidator = validator.New()

func init() {

	addAlias("xName", "max=20")
	addValidate("xDuration", func(fl validator.FieldLevel) bool {
		value, ok := toString(fl)
		if !ok {
			return false
		}
		_, err := time.ParseDuration(value)
		return err == nil
	})
	addValidate("xAddr", func(fl validator.FieldLevel) bool {
		value, ok := toString(fl)
		if !ok {
			return false
		}
		urlInfo, err := url.Parse(value)
		if err != nil {
			return false
		}
		return contains([]string{"http", "https"}, urlInfo.Scheme)
	})
	addValidate("xURLPath", func(fl validator.FieldLevel) bool {
		value, ok := toString(fl)
		if !ok {
			return false
		}
		return value != "" && value[0] == '/'
	})
	addValidate("xDivide", func(fl validator.FieldLevel) bool {
		value, ok := toString(fl)
		if !ok {
			return false
		}
		arr := strings.Split(value, ":")
		return len(arr) == 2
	})
	addValidate("xSize", func(fl validator.FieldLevel) bool {
		value, ok := toString(fl)
		if !ok {
			return false
		}
		_, err := humanize.ParseBytes(value)
		return err == nil
	})
	addValidate("xFilter", func(fl validator.FieldLevel) bool {
		value, ok := toString(fl)
		if !ok {
			return false
		}
		_, err := regexp.Compile(value)
		return err == nil
	})
	addValidate("xPolicy", func(fl validator.FieldLevel) bool {
		value, ok := toString(fl)
		if !ok {
			return false
		}
		return contains([]string{
			us.PolicyFirst,
			us.PolicyRandom,
			us.PolicyRoundRobin,
			us.PolicyLeastconn,
		}, value)
	})
}

// toString 转换为string
func toString(fl validator.FieldLevel) (string, bool) {
	value := fl.Field()
	if value.Kind() != reflect.String {
		return "", false
	}
	return value.String(), true
}

func contains(arr []string, str string) bool {
	found := false
	for _, item := range arr {
		if item == str {
			found = true
			break
		}
	}
	return found
}

func addValidate(tag string, fn validator.Func, args ...bool) {
	err := defaultValidator.RegisterValidation(tag, fn, args...)
	if err != nil {
		panic(err)
	}
}

func addAlias(alias, tags string) {
	defaultValidator.RegisterAlias(alias, tags)
}
