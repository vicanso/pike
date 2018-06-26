package benchmark

import (
	"encoding/binary"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/labstack/echo"
)

type RequestStatus struct {
	createdAt uint32
	ttl       uint16
	// 请求状态 fetching hitForPass 等
	status int
	// 如果此请求为fetching，则此时相同的请求会写入一个chan
	waitingChans []chan int
}

type Context struct {
	echo.Context
}

func BenchmarkConvertNoop(b *testing.B) {

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
	}
}

func BenchmarkConvertUint32(b *testing.B) {
	now := time.Now().Unix()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = uint32(now)
	}
}

func BenchmarkConvertUint16(b *testing.B) {

	b.ResetTimer()
	code := 200
	for i := 0; i < b.N; i++ {
		_ = uint16(code)
	}
}

func BenchmarkUint16ToByte(b *testing.B) {
	b.ResetTimer()
	var v uint16 = 200
	for i := 0; i < b.N; i++ {
		buf := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf, v)
	}
}

func BenchmarkCreateRequestStatus(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &RequestStatus{}
	}
}

func BenchmarkRequestStatusMap(b *testing.B) {
	rsMap := make(map[string]*RequestStatus)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, uint32(i))
		rsMap[string(buf)] = &RequestStatus{}
	}
}

func BenchmarkDeleteRequestStatusMap(b *testing.B) {
	rsMap := make(map[string]*RequestStatus)
	for i := 0; i < b.N; i++ {
		buf := make([]byte, 4)
		binary.LittleEndian.PutUint32(buf, uint32(i))
		rsMap[string(buf)] = &RequestStatus{}
	}
	b.ResetTimer()
	for key := range rsMap {
		delete(rsMap, key)
	}
}

func BenchmarkJoinString(b *testing.B) {
	req := &http.Request{
		Method:     "GET",
		Host:       "127.0.0.1",
		RequestURI: "/uesrs/me",
	}
	b.ResetTimer()
	space := " "
	for i := 0; i < b.N; i++ {
		_ = []byte(req.Method + space + req.Host + space + req.RequestURI)
	}
}

func BenchmarkJoinBytes(b *testing.B) {
	req := &http.Request{
		Method:     "GET",
		Host:       "127.0.0.1",
		RequestURI: "/uesrs/me",
	}
	b.ResetTimer()
	space := byte(' ')
	for i := 0; i < b.N; i++ {
		methodLen := len(req.Method)
		hostLen := len(req.Host)
		uriLen := len(req.RequestURI)
		buffer := make([]byte, methodLen+hostLen+uriLen+2)
		len := 0
		copy(buffer[len:], req.Method)
		len += methodLen
		buffer[len] = space
		len++
		copy(buffer[len:], req.Host)
		len += hostLen
		buffer[len] = space
		len++
		copy(buffer[len:], req.RequestURI)
		_ = buffer
	}
}

func BenchmarkConvertContext(b *testing.B) {
	e := echo.New()
	pc := &Context{}
	pc.Context = e.NewContext(nil, nil)
	var c echo.Context = pc
	for i := 0; i < b.N; i++ {
		_, ok := c.(*Context)
		if !ok {
			fmt.Println("conver fail")
		}
	}
}
