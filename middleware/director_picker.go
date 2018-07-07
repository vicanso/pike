package middleware

import (
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/pike"
)

type (
	// DirectorPickerConfig director配置
	DirectorPickerConfig struct {
	}
)

// DirectorPicker 根据请求的参数获取相应的director
// 判断director是否符合是顺序查询，因此需要将directors先根据优先级排好序
func DirectorPicker(config DirectorPickerConfig, directors pike.Directors) pike.Middleware {
	return func(c *pike.Context, next pike.Next) error {
		done := c.ServerTiming.Start(pike.ServerTimingDirectorPicker)
		// 如果缓存数据，不需要获取director
		if c.Status == cache.Cacheable {
			done()
			return next()
		}
		req := c.Request
		host := req.Host
		uri := req.RequestURI
		found := false

		for _, d := range directors {
			if d.Match(host, uri) {
				c.Director = d
				found = true
				break
			}
		}
		if !found {
			done()
			return ErrDirectorNotFound
		}
		done()
		return next()
	}
}
