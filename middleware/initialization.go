package middleware

import (
	"strings"

	"github.com/vicanso/pike/performance"
	"github.com/vicanso/pike/pike"
	"github.com/vicanso/pike/util"
)

const (
	defaultConcurrency = 256 * 1000
)

type (
	// InitializationConfig 初始化配置
	InitializationConfig struct {
		Header      []string
		Concurrency int
	}
)

// Initialization 相关一些初始化的操作
func Initialization(config InitializationConfig) pike.Middleware {
	customHeader := make(map[string]string)
	// 将自定义的http response header格式化
	for _, v := range config.Header {
		arr := strings.Split(v, ":")
		if len(arr) != 2 {
			continue
		}
		customHeader[arr[0]] = arr[1]
	}

	// 获取限制并发请求数
	concurrency := uint32(defaultConcurrency)
	if config.Concurrency != 0 {
		concurrency = uint32(config.Concurrency)
	}

	return func(c *pike.Context, next pike.Next) error {
		done := c.ServerTiming.Start(pike.ServerTimingInitialization)
		performance.IncreaseRequestCount()

		defer func() {
			performance.DecreaseConcurrency()
			status := c.Response.Status()
			use := util.GetTimeConsuming(c.CreatedAt)
			performance.AddRequestStats(status, use)
		}()

		resHeader := c.Response.Header()
		for k, v := range customHeader {
			resHeader.Add(k, v)
		}

		v := performance.IncreaseConcurrency()
		if v > concurrency {
			done()
			return ErrTooManyRequest
		}
		done()
		return next()
	}
}
