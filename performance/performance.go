package performance

import (
	"runtime"
	"sync/atomic"
	"time"
)

var concurrency uint32
var startedAt = time.Now().Format(time.RFC3339)

// Stats 性能的统计
type Stats struct {
	Concurrency uint32 `json:"concurrency"`
	Sys         int    `json:"sys"`
	HeapSys     int    `json:"heapSys"`
	HeapInuse   int    `json:"heapInuse"`
	StartedAt   string `json:"startedAt"`
}

// IncreaseConcurrency concurrency 加一
func IncreaseConcurrency() {
	atomic.AddUint32(&concurrency, 1)
}

// DecreaseConcurrency concurrency 减一
func DecreaseConcurrency() {
	atomic.AddUint32(&concurrency, ^uint32(0))
}

// GetConcurrency 获取 concurrency
func GetConcurrency() uint32 {
	return concurrency
}

// GetStats 获取Memory的使用
func GetStats() *Stats {
	var mb uint64 = 1024 * 1024
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)
	stats := &Stats{
		Concurrency: GetConcurrency(),
		Sys:         int(m.Sys / mb),
		HeapSys:     int(m.HeapSys / mb),
		HeapInuse:   int(m.HeapInuse / mb),
		StartedAt:   startedAt,
	}
	return stats
}
