package server

import (
	"sync"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/vicanso/cod"
)

func BenchmarkLoadCodByPointer(b *testing.B) {
	var currentIns *cod.Cod
	d := cod.New()
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&currentIns)), unsafe.Pointer(d))
	for i := 0; i < b.N; i++ {
		_ = (*cod.Cod)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&currentIns))))
	}
}

func BenchmarkLoadCodByMap(b *testing.B) {
	d := cod.New()
	m := sync.Map{}
	key := "instance"
	m.Store(key, d)
	for i := 0; i < b.N; i++ {
		v, _ := m.Load(key)
		_ = v.(*cod.Cod)
	}
}
