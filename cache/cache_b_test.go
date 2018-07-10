package cache

import (
	"net/http"
	"testing"
	"time"
	"unsafe"
)

func BenchmarkSaveResponse(b *testing.B) {
	header := make(http.Header)
	header["X-Token"] = []string{
		"abcd",
	}
	data := make([]byte, 10*1024)
	resp := &Response{
		Header:   header,
		GzipBody: data,
		BrBody:   data,
	}
	key := []byte("ABCD")
	c := Client{
		Path: dbPath,
	}
	err := c.Init()
	if err != nil {
		panic(err)
	}
	defer c.Close()
	for i := 0; i < b.N; i++ {
		err := c.SaveResponse(key, resp)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkGetResponse(b *testing.B) {
	header := make(http.Header)
	header["X-Token"] = []string{
		"abcd",
	}
	data := make([]byte, 10*1024)
	resp := &Response{
		Header:   header,
		GzipBody: data,
		BrBody:   data,
	}
	key := []byte("ABCD")
	c := Client{
		Path: dbPath,
	}
	err := c.Init()
	if err != nil {
		panic(err)
	}
	err = c.SaveResponse(key, resp)
	if err != nil {
		panic(err)
	}
	for i := 0; i < b.N; i++ {
		_, err := c.GetResponse(key)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkBytesToString(b *testing.B) {
	key := []byte("ABCD")
	for i := 0; i < b.N; i++ {
		_ = string(key)
	}
}

func BenchmarkByteSliceToString(b *testing.B) {
	key := []byte("ABCD")
	for i := 0; i < b.N; i++ {
		_ = *(*string)(unsafe.Pointer(&key))
	}
}

func BenchmarkTimeNowNano(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = time.Now().UnixNano()
	}
}

func BenchmarkTimeNowSec(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = time.Now().Unix()
	}
}

func BenchmarkGetRequestStatus(b *testing.B) {
	key := []byte("ABCDEFG")
	c := Client{
		Path: dbPath,
	}
	err := c.Init()
	if err != nil {
		panic(err)
	}
	c.GetRequestStatus(key)
	c.HitForPass(key, 300)
	for j := 0; j < 4; j++ {
		go func() {
			for i := 0; i < b.N; i++ {
				c.GetRequestStatus(key)
			}
		}()
	}
	for i := 0; i < b.N; i++ {
		c.GetRequestStatus(key)
	}
}
