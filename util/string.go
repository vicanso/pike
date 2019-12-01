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

package util

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"
	"unsafe"
)

const (
	// spaceByte 空格
	spaceByte = byte(' ')
)

// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// randomString create a random string
func randomString(baseLetters string, n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(baseLetters) {
			b[i] = baseLetters[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

// RandomString create a random string
func RandomString(n int) string {
	return randomString(letterBytes, n)
}

// ByteSliceToString converts a []byte to string without a heap allocation.
func ByteSliceToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// GetIdentity get identity of request
func GetIdentity(req *http.Request) []byte {
	methodLen := len(req.Method)
	hostLen := len(req.Host)
	uriLen := len(req.RequestURI)
	buffer := make([]byte, methodLen+hostLen+uriLen+2)
	len := 0

	copy(buffer[len:], req.Method)
	len += methodLen

	buffer[len] = spaceByte
	len++

	copy(buffer[len:], req.Host)
	len += hostLen

	buffer[len] = spaceByte
	len++

	copy(buffer[len:], req.RequestURI)
	return buffer
}

// GenerateETag generate eTag
func GenerateETag(buf []byte) string {
	size := len(buf)
	if size == 0 {
		return `"0-2jmj7l5rSw0yVb_vlWAYkK_YBwk="`
	}
	h := sha1.New()
	h.Write(buf)
	hash := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return fmt.Sprintf(`"%x-%s"`, size, hash)
}

// ContainesString contain string
func ContainesString(arr []string, str string) bool {
	for _, item := range arr {
		if item == str {
			return true
		}
	}
	return false
}

// ConvertToHTTPHeader convert to http header
func ConvertToHTTPHeader(values []string) http.Header {
	if len(values) == 0 {
		return nil
	}
	h := make(http.Header)
	for _, item := range values {
		arr := strings.Split(item, ":")
		if len(arr) == 2 {
			h.Add(arr[0], arr[1])
		}
	}
	return h
}

// MergeHeader merge header
func MergeHeader(h1, h2 http.Header) {
	for key, values := range h2 {
		for _, value := range values {
			h1.Add(key, value)
		}
	}
}
