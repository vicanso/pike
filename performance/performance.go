package performance

import (
	"runtime"
	"sync/atomic"
	"time"

	"github.com/vicanso/pike/cache"

	"github.com/vicanso/pike/vars"
)

var concurrency uint32
var startedAt = time.Now().UTC().Format(time.RFC3339)

// 记录当前处理的请求数
var requestCount uint64

// Stats 性能的统计
type Stats struct {
	// 当前并发处理请求数
	Concurrency uint32 `json:"concurrency"`
	// 使用内存
	Sys int `json:"sys"`
	// heap sys内存
	HeapSys int `json:"heapSys"`
	// heap使用内存
	HeapInuse int `json:"heapInuse"`
	// 程序启动时间
	StartedAt string `json:"startedAt"`
	// routine数量
	RoutineCount int `json:"routine"`
	// 缓存数量（包括hit for pass 与 cacheable）
	CacheCount int `json:"cacheCount"`
	// 正在请求的数量（请求backend）
	Fetching int `json:"fetching"`
	// 等待中的请求数量（由于有相同的请求为fetching）
	Waiting int `json:"waiting"`
	// 可缓存的请求数量
	Cacheable int `json:"cacheable"`
	// hit for pass的缓存数量
	HitForPass int `json:"hitForPass"`
	// 总的处理请求量
	RequestCount uint64 `json:"requestCount"`
	// version版本号
	Version string `json:"version"`
	// FileSize db数据文件的大小
	FileSize int `json:"fileSize"`
}

// IncreaseConcurrency concurrency 加一
func IncreaseConcurrency() uint32 {
	return atomic.AddUint32(&concurrency, 1)
}

// DecreaseConcurrency concurrency 减一
func DecreaseConcurrency() uint32 {
	return atomic.AddUint32(&concurrency, ^uint32(0))
}

// IncreaseRequestCount 处理请求数加一
func IncreaseRequestCount() uint64 {
	return atomic.AddUint64(&requestCount, 1)
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
func GetStats(client *cache.Client) *Stats {
	var mb uint64 = 1024 * 1024
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)

	result := client.GetStats()

	stats := &Stats{
		Concurrency:  GetConcurrency(),
		Sys:          int(m.Sys / mb),
		HeapSys:      int(m.HeapSys / mb),
		HeapInuse:    int(m.HeapInuse / mb),
		StartedAt:    startedAt,
		RoutineCount: runtime.NumGoroutine(),
		CacheCount:   client.Size(),
		Fetching:     result.Fetching,
		Waiting:      result.Waiting,
		Cacheable:    result.Cacheable,
		HitForPass:   result.HitForPass,
		RequestCount: requestCount,
		Version:      vars.Version,
		FileSize:     result.FileSize,
	}
	return stats
}