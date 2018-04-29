package custommiddleware

import (
	"strings"

	"github.com/labstack/echo"
	"github.com/vicanso/pike/performance"
)

type (
	// InitializationConfig 初始化配置
	InitializationConfig struct {
		Header []string
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
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			performance.IncreaseRequestCount()
			performance.IncreaseConcurrency()
			defer performance.DecreaseConcurrency()
			resHeader := c.Response().Header()
			for k, v := range customHeader {
				resHeader.Add(k, v)
			}
			return next(c)
		}
	}
}
