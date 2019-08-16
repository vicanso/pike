package server

import (
	"sync"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/vicanso/elton"
)

func BenchmarkLoadCodByPointer(b *testing.B) {
	var currentIns *elton.Elton
	d := elton.New()
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&currentIns)), unsafe.Pointer(d))
	for i := 0; i < b.N; i++ {
		_ = (*elton.Elton)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&currentIns))))
	}
}

func BenchmarkLoadCodByMap(b *testing.B) {
	d := elton.New()
	m := sync.Map{}
	key := "instance"
	m.Store(key, d)
	for i := 0; i < b.N; i++ {
		v, _ := m.Load(key)
		_ = v.(*elton.Elton)
	}
}
