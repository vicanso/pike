package cache

import (
	"net/http"
	"testing"
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

func BenchmarkGetRequestStatus(b *testing.B) {
	key := []byte("ABCD")
	c := Client{
		Path: dbPath,
	}
	err := c.Init()
	if err != nil {
		panic(err)
	}
	c.HitForPass(key, 300)
	for i := 0; i < b.N; i++ {
		c.GetRequestStatus(key)
	}
}
