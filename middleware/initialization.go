package custommiddleware

import (
	"strings"

	"github.com/vicanso/pike/vars"

	"github.com/labstack/echo"
	"github.com/vicanso/pike/performance"
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
func Initialization(config InitializationConfig) echo.MiddlewareFunc {
	customHeader := make(map[string]string)
	for _, v := range config.Header {
		arr := strings.Split(v, ":")
		if len(arr) != 2 {
			continue
		}
		customHeader[arr[0]] = arr[1]
	}
	concurrency := uint32(defaultConcurrency)
	if config.Concurrency != 0 {
		concurrency = uint32(config.Concurrency)
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			defer performance.DecreaseConcurrency()
			performance.IncreaseRequestCount()
			v := performance.IncreaseConcurrency()
			if v > concurrency {
				return vars.ErrTooManyRequst
			}
			resHeader := c.Response().Header()
			for k, v := range customHeader {
				resHeader.Add(k, v)
			}
			return next(c)
		}
	}
}
