package benchmark

import (
	"encoding/binary"
	"testing"
	"time"
)

type RequestStatus struct {
	createdAt uint32
	ttl       uint16
	// 请求状态 fetching hitForPass 等
	status int
	// 如果此请求为fetching，则此时相同的请求会写入一个chan
	waitingChans []chan int
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
