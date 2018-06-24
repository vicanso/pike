package performance

import (
	"runtime"
	"sync/atomic"
	"time"

	"github.com/vicanso/pike/cache"

	"github.com/vicanso/pike/vars"
)

const (
	minStatus = 1
	maxStatus = 5
)

var (
	spdyList = []int{30, 300, 1000, 3000}
	// 并发请求数
	concurrency uint32
	// 程序启动时间
	startedAt = time.Now().UTC().Format(time.RFC3339)
	// 记录当前处理的请求数
	requestCount uint64
	// 1xx状态汇总
	status1Count uint64
	// 2xx状态汇总
	status2Count uint64
	// 3xx状态汇总
	status3Count uint64
	// 4xx状态汇总
	status4Count uint64
	// 5xx状态汇总
	status5Count uint64
	// 少于30ms的处理汇总
	spdy0Count uint64
	// 少于300ms的处理汇总
	spdy1Count uint64
	// 少于1000ms的处理汇总
	spdy2Count uint64
	// 少于3000ms的处理汇总
	spdy3Count uint64
	// 大于3000ms的处理汇总
	spdy4Count uint64
)

type (
	// Stats 性能的统计
	Stats struct {
		// 状态码汇总
		Status map[string]uint64 `json:"status"`
		// spdy 汇总
		Spdy map[string]uint64 `json:"spdy"`
		// 当前并发处理请求数
		Concurrency uint32 `json:"concurrency"`
		// GoMaxProcs 当前使用的cpu数
		GoMaxProcs int `json:"goMaxProcs"`
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
)

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

// AddRequestStats 设置性能统计
func AddRequestStats(status, use int) {
	spdy := 0
	for i, v := range spdyList {
		if use > v {
			spdy = i + 1
		}
	}
	switch spdy {
	case 0:
		atomic.AddUint64(&spdy0Count, 1)
	case 1:
		atomic.AddUint64(&spdy1Count, 1)
	case 2:
		atomic.AddUint64(&spdy2Count, 1)
	case 3:
		atomic.AddUint64(&spdy3Count, 1)
	case 4:
		atomic.AddUint64(&spdy4Count, 1)
	}
	switch status / 100 {
	case 1:
		atomic.AddUint64(&status1Count, 1)
	case 2:
		atomic.AddUint64(&status2Count, 1)
	case 3:
		atomic.AddUint64(&status3Count, 1)
	case 4:
		atomic.AddUint64(&status4Count, 1)
	case 5:
		atomic.AddUint64(&status5Count, 1)
	}
}

// GetStats 获取系统的使用
func GetStats(client *cache.Client) *Stats {
	var mb uint64 = 1024 * 1024
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)

	result := client.GetStats()

	stats := &Stats{
		Status: map[string]uint64{
			"1": status1Count,
			"2": status2Count,
			"3": status3Count,
			"4": status4Count,
			"5": status5Count,
		},
		Spdy: map[string]uint64{
			"0": spdy0Count,
			"1": spdy1Count,
			"2": spdy2Count,
			"3": spdy3Count,
			"4": spdy4Count,
		},
		GoMaxProcs:   runtime.GOMAXPROCS(0),
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
