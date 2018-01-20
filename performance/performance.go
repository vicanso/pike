package performance

import (
	"runtime"
	"sync/atomic"
	"time"

	"github.com/vicanso/pike/cache"
)

var concurrency uint32
var startedAt = time.Now().UTC().Format(time.RFC3339)

// 记录当前处理的请求数
var requestCount uint64

// Stats 性能的统计
type Stats struct {
	Concurrency  uint32 `json:"concurrency"`
	Sys          int    `json:"sys"`
	HeapSys      int    `json:"heapSys"`
	HeapInuse    int    `json:"heapInuse"`
	StartedAt    string `json:"startedAt"`
	RoutineCount int    `json:"routine"`
	CacheCount   int    `json:"cacheCount"`
	Fetching     int    `json:"fetching"`
	Waiting      int    `json:"waiting"`
	Cacheable    int    `json:"cacheable"`
	HitForPass   int    `json:"hitForPass"`
	RequestCount uint64 `json:"requestCount"`
	LSM          int    `json:"lsm"`
	VLog         int    `json:"vLog"`
}

// IncreaseConcurrency concurrency 加一
func IncreaseConcurrency() {
	atomic.AddUint32(&concurrency, 1)
}

// DecreaseConcurrency concurrency 减一
func DecreaseConcurrency() {
	atomic.AddUint32(&concurrency, ^uint32(0))
}

// IncreaseRequestCount 处理请求数加一
func IncreaseRequestCount() {
	atomic.AddUint64(&requestCount, 1)
}

// GetRequstCount 获取处理请求数
func GetRequstCount() uint64 {
	return requestCount
}

// GetConcurrency 获取 concurrency
func GetConcurrency() uint32 {
	return concurrency
}

// GetStats 获取系统的使用
func GetStats() *Stats {
	var mb uint64 = 1024 * 1024
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)
	fetching, waiting, cacheable, hitForPass := cache.Stats()
	lsm, vlog := cache.DataSize()
	stats := &Stats{
		Concurrency:  GetConcurrency(),
		Sys:          int(m.Sys / mb),
		HeapSys:      int(m.HeapSys / mb),
		HeapInuse:    int(m.HeapInuse / mb),
		StartedAt:    startedAt,
		RoutineCount: runtime.NumGoroutine(),
		CacheCount:   cache.Size(),
		Fetching:     fetching,
		Waiting:      waiting,
		Cacheable:    cacheable,
		HitForPass:   hitForPass,
		RequestCount: requestCount,
		LSM:          lsm,
		VLog:         vlog,
	}
	return stats
}
