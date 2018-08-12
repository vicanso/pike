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
		Header        []string
		RequestHeader []string
		Concurrency   int
	}
)

func genHeader(header []string) map[string]string {
	m := make(map[string]string)
	// 将自定义的http response header格式化
	for _, v := range header {
		arr := strings.Split(v, ":")
		if len(arr) != 2 {
			continue
		}
		value := arr[1]
		v := util.CheckAndGetValueFromEnv(value)
		if len(v) != 0 {
			value = v
		}
		m[arr[0]] = value
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

// Initialization 相关一些初始化的操作
func Initialization(config InitializationConfig) pike.Middleware {
	customHeader := genHeader(config.Header)
	customReqHeader := genHeader(config.RequestHeader)

	// 获取限制并发请求数
	concurrency := uint32(defaultConcurrency)
	if config.Concurrency > 0 {
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

		reqHeader := c.Request.Header
		for k, v := range customReqHeader {
			reqHeader.Add(k, v)
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
