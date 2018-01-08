package performance

import (
	"runtime"
	"sync/atomic"
	"time"

	"../cache"
)

var concurrency uint32
var startedAt = time.Now().Format(time.RFC3339)

// 记录每分钟的处理请求数
var requestCountList []uint32

// 记录当前日期，日期变化时，清除每分钟的请求数统计
var currentDay = -1

// Stats 性能的统计
type Stats struct {
	Concurrency      uint32   `json:"concurrency"`
	Sys              int      `json:"sys"`
	HeapSys          int      `json:"heapSys"`
	HeapInuse        int      `json:"heapInuse"`
	StartedAt        string   `json:"startedAt"`
	RequestCountList []uint32 `json:"requestCountList"`
	CacheCount       int      `json:"cacheCount"`
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
	now := time.Now()
	day := now.Day()
	if currentDay != day {
		// 为了更好的性能，此处并没有对并发做锁处理
		// 因为只是统计数据，如果真出错，数据就只是重置了多次
		// 而且只是在日期变化时才有变化
		requestCountList = make([]uint32, 24*60)
		currentDay = day
	}
	hour := now.Hour()
	minute := now.Minute()
	index := hour*60 + minute
	atomic.AddUint32(&requestCountList[index], 1)
}

// GetRequstCountList 获取处理请求数列表
func GetRequstCountList() []uint32 {
	return requestCountList
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
	stats := &Stats{
		Concurrency:      GetConcurrency(),
		Sys:              int(m.Sys / mb),
		HeapSys:          int(m.HeapSys / mb),
		HeapInuse:        int(m.HeapInuse / mb),
		StartedAt:        startedAt,
		RequestCountList: requestCountList,
		CacheCount:       cache.Size(),
	}
	return stats
}
