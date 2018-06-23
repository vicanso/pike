package custommiddleware

import (
	"strings"

	"github.com/vicanso/pike/util"
	"github.com/vicanso/pike/vars"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/vicanso/pike/performance"
)

const (
	defaultConcurrency = 256 * 1000
)

type (
	// InitializationConfig 初始化配置
	InitializationConfig struct {
		Skipper     middleware.Skipper
		Header      []string
		Concurrency int
	}
)

// Initialization 相关一些初始化的操作
func Initialization(config InitializationConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = middleware.DefaultSkipper
	}
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
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			c.Logger().Debug("init middleware")
			if config.Skipper(c) {
				return next(c)
			}
			pc := c.(*Context)
			defer func() {
				performance.DecreaseConcurrency()
				resp := pc.Response()
				status := resp.Status
				use := util.GetTimeConsuming(pc.createdAt)
				performance.AddRequestStats(status, use)
			}()
			resHeader := pc.Response().Header()
			for k, v := range customHeader {
				resHeader.Add(k, v)
			}
			performance.IncreaseRequestCount()
			v := performance.IncreaseConcurrency()
			if v > concurrency {
				return vars.ErrTooManyRequest
			}
			return next(pc)
		}
	}
}
