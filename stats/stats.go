package stats

import (
	"runtime"
	"sync/atomic"

	"github.com/vicanso/pike/df"
)

const (
	minStatus = 1
	maxStatus = 5
)

var (
	spdyList = []int{30, 300, 1000, 3000}
)

type (
	// Info 性能的统计
	Info struct {
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
		// recover的数量
		RecoverCount uint64 `json:"recoverCount"`
		// 总的处理请求量
		RequestCount uint64 `json:"requestCount"`
		// version版本号
		Version string `json:"version"`
		// 构建时间
		BuildedAt string `json:"buildedAt"`
		// 当前版本的git commit id
		CommitID string `json:"commitId"`
		// 编译的go版本
		GoVersion string `json:"goVersion"`
	}
	// Stats application stats
	Stats struct {
		// 并发请求数
		concurrency uint32
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
		// 出现recover的次数
		recoverCount uint64
	}
)

// New create a new stats instance
func New() *Stats {
	return &Stats{}
}

// IncreaseConcurrency concurrency 加一
func (s *Stats) IncreaseConcurrency() uint32 {
	return atomic.AddUint32(&s.concurrency, 1)
}

// DecreaseConcurrency concurrency 减一
func (s *Stats) DecreaseConcurrency() uint32 {
	return atomic.AddUint32(&s.concurrency, ^uint32(0))
}

// IncreaseRequestCount 处理请求数加一
func (s *Stats) IncreaseRequestCount() uint64 {
	return atomic.AddUint64(&s.requestCount, 1)
}

// IncreaseRecoverCount recover的次数加一
func (s *Stats) IncreaseRecoverCount() uint64 {
	return atomic.AddUint64(&s.recoverCount, 1)
}

// GetRequestCount 获取处理请求数
func (s *Stats) GetRequestCount() uint64 {
	return atomic.LoadUint64(&s.requestCount)
}

// GetConcurrency 获取 concurrency
func (s *Stats) GetConcurrency() uint32 {
	return atomic.LoadUint32(&s.concurrency)
}

// AddRequestStats 设置性能统计
func (s *Stats) AddRequestStats(status, use int) {
	spdy := 0
	for i, v := range spdyList {
		if use > v {
			spdy = i + 1
		}
	}
	switch spdy {
	case 0:
		atomic.AddUint64(&s.spdy0Count, 1)
	case 1:
		atomic.AddUint64(&s.spdy1Count, 1)
	case 2:
		atomic.AddUint64(&s.spdy2Count, 1)
	case 3:
		atomic.AddUint64(&s.spdy3Count, 1)
	case 4:
		atomic.AddUint64(&s.spdy4Count, 1)
	}
	switch status / 100 {
	case 1:
		atomic.AddUint64(&s.status1Count, 1)
	case 2:
		atomic.AddUint64(&s.status2Count, 1)
	case 3:
		atomic.AddUint64(&s.status3Count, 1)
	case 4:
		atomic.AddUint64(&s.status4Count, 1)
	case 5:
		atomic.AddUint64(&s.status5Count, 1)
	}
}

// GetInfo 获取系统的使用
func (s *Stats) GetInfo() *Info {
	var mb uint64 = 1024 * 1024
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)

	stats := &Info{
		Status: map[string]uint64{
			"1": atomic.LoadUint64(&s.status1Count),
			"2": atomic.LoadUint64(&s.status2Count),
			"3": atomic.LoadUint64(&s.status3Count),
			"4": atomic.LoadUint64(&s.status4Count),
			"5": atomic.LoadUint64(&s.status5Count),
		},
		Spdy: map[string]uint64{
			"0": atomic.LoadUint64(&s.spdy0Count),
			"1": atomic.LoadUint64(&s.spdy1Count),
			"2": atomic.LoadUint64(&s.spdy2Count),
			"3": atomic.LoadUint64(&s.spdy3Count),
			"4": atomic.LoadUint64(&s.spdy4Count),
		},

		GoMaxProcs:   runtime.GOMAXPROCS(0),
		Concurrency:  s.GetConcurrency(),
		Sys:          int(m.Sys / mb),
		HeapSys:      int(m.HeapSys / mb),
		HeapInuse:    int(m.HeapInuse / mb),
		StartedAt:    df.StartedAt,
		RoutineCount: runtime.NumGoroutine(),
		RequestCount: atomic.LoadUint64(&s.requestCount),
		Version:      df.Version,
		BuildedAt:    df.BuildedAt,
		CommitID:     df.CommitID,
		GoVersion:    runtime.Version(),
	}
	return stats
}
