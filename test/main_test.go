package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
)

type (
	// Result response result
	Result struct {
		Count   int    `json:"count"`
		Account string `json:"account"`
	}
)

const (
	maxCount = 10
)

func getURL(url string) string {
	return "http://127.0.0.1:3015/api" + url
}

func FindIntIndex(collection []int, value int) int {
	index := -1
	for i, v := range collection {
		if index != -1 {
			return index
		}
		if v == value {
			index = i
		}
	}
	return index
}

func includes(collection []int, value int) bool {
	index := -1
	for i, v := range collection {
		if index != -1 {
			break
		}
		if v == value {
			index = i
		}
	}
	return index != -1
}

func doRequset(method, url string, body io.Reader) (result *Result, resp *http.Response, err error) {
	c := http.Client{}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return
	}
	resp, err = c.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	result = &Result{}
	err = json.Unmarshal(buf, result)
	if err != nil {
		return
	}
	return
}

func TestPostRequest(t *testing.T) {
	url := getURL("/users")
	done := make(chan int)
	mutex := sync.Mutex{}
	countList := []int{}
	var count int32
	fail := func(msg string, args ...interface{}) {
		done <- 0
		t.Fatalf(msg, args...)
	}

	for index := 0; index < maxCount; index++ {
		go func() {
			result, resp, err := doRequset("POST", url, nil)
			if err != nil {
				fail("do post request fail, %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				fail("response status is not ok")
			}
			if resp.Header.Get("X-Status") != "pass" {
				fail("x-status of response should be pass")
			}

			n := atomic.AddInt32(&count, 1)
			mutex.Lock()
			countList = append(countList, result.Count)
			defer mutex.Unlock()
			if int(n) == maxCount {
				done <- 1
			}
		}()
	}
	if <-done != 1 {
		t.Fatalf("post request check fail")
	}
	for index := 0; index < maxCount; index++ {
		if !includes(countList, index+1) {
			t.Fatalf("the count check fail")
		}
	}
}

func TestPutRequest(t *testing.T) {
	url := getURL("/users/1")
	done := make(chan int)
	mutex := sync.Mutex{}
	countList := []int{}
	var count int32
	fail := func(msg string, args ...interface{}) {
		done <- 0
		t.Fatalf(msg, args...)
	}

	for index := 0; index < maxCount; index++ {
		go func() {
			result, resp, err := doRequset("PUT", url, nil)
			if err != nil {
				fail("do put request fail, %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				fail("response status is not ok")
			}
			if resp.Header.Get("X-Status") != "pass" {
				fail("x-status of response should be pass")
			}
			n := atomic.AddInt32(&count, 1)
			mutex.Lock()
			countList = append(countList, result.Count)
			defer mutex.Unlock()
			if int(n) == maxCount {
				done <- 1
			}
		}()
	}
	if <-done != 1 {
		t.Fatalf("put request check fail")
	}
	for index := 0; index < maxCount; index++ {
		if !includes(countList, index+1) {
			t.Fatalf("the count check fail")
		}
	}
}

func TestPatchRequest(t *testing.T) {
	url := getURL("/users/1")
	done := make(chan int)
	mutex := sync.Mutex{}
	countList := []int{}
	var count int32
	fail := func(msg string, args ...interface{}) {
		done <- 0
		t.Fatalf(msg, args...)
	}

	for index := 0; index < maxCount; index++ {
		go func() {
			result, resp, err := doRequset("PATCH", url, nil)
			if err != nil {
				fail("do patch request fail, %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				fail("response status is not ok")
			}
			if resp.Header.Get("X-Status") != "pass" {
				fail("x-status of response should be pass")
			}
			n := atomic.AddInt32(&count, 1)
			mutex.Lock()
			countList = append(countList, result.Count)
			defer mutex.Unlock()
			if int(n) == maxCount {
				done <- 1
			}
		}()
	}
	if <-done != 1 {
		t.Fatalf("patch request check fail")
	}
	for index := 0; index < maxCount; index++ {
		if !includes(countList, index+1) {
			t.Fatalf("the count check fail")
		}
	}
}

func TestDeleteRequest(t *testing.T) {
	url := getURL("/users/1")
	done := make(chan int)
	mutex := sync.Mutex{}
	countList := []int{}
	var count int32
	fail := func(msg string, args ...interface{}) {
		done <- 0
		t.Fatalf(msg, args...)
	}

	for index := 0; index < maxCount; index++ {
		go func() {
			result, resp, err := doRequset("DELETE", url, nil)
			if err != nil {
				fail("do delete request fail, %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				fail("response status is not ok")
			}
			if resp.Header.Get("X-Status") != "pass" {
				fail("x-status of response should be pass")
			}
			n := atomic.AddInt32(&count, 1)
			mutex.Lock()
			countList = append(countList, result.Count)
			defer mutex.Unlock()
			if int(n) == maxCount {
				done <- 1
			}
		}()
	}
	if <-done != 1 {
		t.Fatalf("delete request check fail")
	}
	for index := 0; index < maxCount; index++ {
		if !includes(countList, index+1) {
			t.Fatalf("the count check fail")
		}
	}
}

func TestGetNoCacheRequest(t *testing.T) {
	url := getURL("/no-cache")
	done := make(chan int)
	mutex := sync.Mutex{}
	countList := []int{}
	statusList := []string{}
	var count int32
	fail := func(msg string, args ...interface{}) {
		done <- 0
		t.Fatalf(msg, args...)
	}

	for index := 0; index < maxCount; index++ {
		go func() {
			result, resp, err := doRequset("GET", url, nil)
			if err != nil {
				fail("do get request fail, %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				fail("response status is not ok")
			}
			n := atomic.AddInt32(&count, 1)
			mutex.Lock()
			statusList = append(statusList, resp.Header.Get("X-Status"))
			countList = append(countList, result.Count)
			defer mutex.Unlock()
			if int(n) == maxCount {
				done <- 1
			}
		}()
	}
	if <-done != 1 {
		t.Fatalf("get no-cache request check fail")
	}
	for index := 0; index < maxCount; index++ {
		if !includes(countList, index+1) {
			t.Fatalf("the count check fail")
		}
	}
	// 首次无缓存时，为fetching，后续为hitForPass
	for index, v := range statusList {
		if index == 0 {
			if v != "fetching" {
				t.Fatalf("the first request should be fetching")
			}
		} else if v != "hitForPass" {
			t.Fatalf("the rest request should be hit for pass")
		}
	}
}

func TestGetNoStoreRequest(t *testing.T) {
	url := getURL("/no-store")
	done := make(chan int)
	mutex := sync.Mutex{}
	countList := []int{}
	statusList := []string{}
	var count int32
	fail := func(msg string, args ...interface{}) {
		done <- 0
		t.Fatalf(msg, args...)
	}

	for index := 0; index < maxCount; index++ {
		go func() {
			result, resp, err := doRequset("GET", url, nil)
			if err != nil {
				fail("do get request fail, %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				fail("response status is not ok")
			}
			n := atomic.AddInt32(&count, 1)
			mutex.Lock()
			statusList = append(statusList, resp.Header.Get("X-Status"))
			countList = append(countList, result.Count)
			defer mutex.Unlock()
			if int(n) == maxCount {
				done <- 1
			}
		}()
	}
	if <-done != 1 {
		t.Fatalf("get no-store request check fail")
	}
	for index := 0; index < maxCount; index++ {
		if !includes(countList, index+1) {
			t.Fatalf("the count check fail")
		}
	}
	// 首次无缓存时，为fetching，后续为hitForPass
	for index, v := range statusList {
		if index == 0 {
			if v != "fetching" {
				t.Fatalf("the first request should be fetching")
			}
		} else if v != "hitForPass" {
			t.Fatalf("the rest request should be hit for pass")
		}
	}
}

func TestGetPrivateCache(t *testing.T) {
	url := getURL("/private-cache")
	done := make(chan int)
	mutex := sync.Mutex{}
	countList := []int{}
	statusList := []string{}
	var count int32
	fail := func(msg string, args ...interface{}) {
		done <- 0
		t.Fatalf(msg, args...)
	}

	for index := 0; index < maxCount; index++ {
		go func() {
			result, resp, err := doRequset("GET", url, nil)
			if err != nil {
				fail("do get request fail, %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				fail("response status is not ok")
			}
			n := atomic.AddInt32(&count, 1)
			mutex.Lock()
			statusList = append(statusList, resp.Header.Get("X-Status"))
			countList = append(countList, result.Count)
			defer mutex.Unlock()
			if int(n) == maxCount {
				done <- 1
			}
		}()
	}
	if <-done != 1 {
		t.Fatalf("get private cache request check fail")
	}
	for index := 0; index < maxCount; index++ {
		if !includes(countList, index+1) {
			t.Fatalf("the count check fail")
		}
	}
	// 首次无缓存时，为fetching，后续为hitForPass
	for index, v := range statusList {
		if index == 0 {
			if v != "fetching" {
				t.Fatalf("the first request should be fetching")
			}
		} else if v != "hitForPass" {
			t.Fatalf("the rest request should be hit for pass")
		}
	}
}

func TestGetMaxAgeZero(t *testing.T) {
	url := getURL("/max-age-zero")
	done := make(chan int)
	mutex := sync.Mutex{}
	countList := []int{}
	statusList := []string{}
	var count int32
	fail := func(msg string, args ...interface{}) {
		done <- 0
		t.Fatalf(msg, args...)
	}

	for index := 0; index < maxCount; index++ {
		go func() {
			result, resp, err := doRequset("GET", url, nil)
			if err != nil {
				fail("do get request fail, %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				fail("response status is not ok")
			}
			n := atomic.AddInt32(&count, 1)
			mutex.Lock()
			statusList = append(statusList, resp.Header.Get("X-Status"))
			countList = append(countList, result.Count)
			defer mutex.Unlock()
			if int(n) == maxCount {
				done <- 1
			}
		}()
	}
	if <-done != 1 {
		t.Fatalf("get max-age=0 request check fail")
	}
	for index := 0; index < maxCount; index++ {
		if !includes(countList, index+1) {
			t.Fatalf("the count check fail")
		}
	}
	// 首次无缓存时，为fetching，后续为hitForPass
	for index, v := range statusList {
		if index == 0 {
			if v != "fetching" {
				t.Fatalf("the first request should be fetching")
			}
		} else if v != "hitForPass" {
			t.Fatalf("the rest request should be hit for pass")
		}
	}
}

func TestGetCacheable(t *testing.T) {
	url := getURL("/cacheable")
	done := make(chan int)
	mutex := sync.Mutex{}
	countList := []int{}
	statusList := []string{}
	var count int32
	fail := func(msg string, args ...interface{}) {
		done <- 0
		t.Fatalf(msg, args...)
	}

	for index := 0; index < maxCount; index++ {
		go func() {
			result, resp, err := doRequset("GET", url, nil)
			if err != nil {
				fail("do get request fail, %v", err)
			}
			if resp.StatusCode != http.StatusOK {
				fail("response status is not ok")
			}
			n := atomic.AddInt32(&count, 1)
			mutex.Lock()
			statusList = append(statusList, resp.Header.Get("X-Status"))
			countList = append(countList, result.Count)
			defer mutex.Unlock()
			if int(n) == maxCount {
				done <- 1
			}
		}()
	}
	if <-done != 1 {
		t.Fatalf("get cacheable request check fail")
	}
	for index := 0; index < maxCount; index++ {
		// 因为是可缓存，所有请求响应都一样
		if countList[index] != 1 {
			t.Fatalf("the count check fail")
		}
	}
	// 首次无缓存时，为fetching，后续为cacheable
	for index, v := range statusList {
		if index == 0 {
			if v != "fetching" {
				t.Fatalf("the first request should be fetching")
			}
		} else if v != "cacheable" {
			t.Fatalf("the rest request should be cacheable")
		}
	}
}
