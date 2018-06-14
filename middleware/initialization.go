package custommiddleware

import (
	"math/rand"
	"strings"
	"time"

	servertiming "github.com/mitchellh/go-server-timing"
	"github.com/oklog/ulid"
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
	seed := time.Now()
	entropy := rand.New(rand.NewSource(seed.UnixNano()))
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if config.Skipper(c) {
				return next(c)
			}
			timing := &servertiming.Header{}
			pikeMetric := timing.NewMetric(vars.PikeMetric)
			pikeMetric.WithDesc("pike handle time").Start()
			c.Set(vars.Timing, timing)
			rid := ulid.MustNew(ulid.Timestamp(seed), entropy).String()
			c.Set(vars.RID, rid)
			startedAt := time.Now()
			defer func() {
				performance.DecreaseConcurrency()
				resp := c.Response()
				status := resp.Status
				use := util.GetTimeConsuming(startedAt)
				performance.AddRequestStats(status, use)
				c.Logger().Debug(rid, " request done")
			}()
			resHeader := c.Response().Header()
			for k, v := range customHeader {
				resHeader.Add(k, v)
			}
			performance.IncreaseRequestCount()
			v := performance.IncreaseConcurrency()
			if v > concurrency {
				return vars.ErrTooManyRequst
			}
			return next(c)
		}
	}
}
