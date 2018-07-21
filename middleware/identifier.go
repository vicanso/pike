package middleware

import (
	"net/http"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/pike"
	"github.com/vicanso/pike/util"
)

type (
	// IdentifierConfig 定义配置
	IdentifierConfig struct {
		Format string
	}
)

// Identifier 对请求的参数校验，生成各类状态值
/*
- 判断请求状态，生成status
- 对于状态非Pass的请求，根据request url 生成identity
*/
func Identifier(config IdentifierConfig, client *cache.Client) pike.Middleware {
	fn := util.GetIdentity
	if config.Format != "" {
		fn = util.GenerateGetIdentity(config.Format)
	}
	return func(c *pike.Context, next pike.Next) error {
		serverTiming := c.ServerTiming
		done := serverTiming.Start(pike.ServerTimingIdentifier)
		req := c.Request
		method := req.Method
		// 只有get与head请求可缓存
		if method != http.MethodGet && method != http.MethodHead {
			c.Status = cache.Pass
			done()
			return next()
		}
		key := fn(req)
		status, ch := client.GetRequestStatus(key)
		// TODO是否应该增加超时机制（proxy中已有超时机制，应该不会有其它流程会卡，因此暂认为无需处理）
		if ch != nil {
			status = <-ch
		}
		c.Status = status
		c.Identity = key
		done()
		return next()
	}
}
