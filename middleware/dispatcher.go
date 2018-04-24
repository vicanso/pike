package customMiddleware

import (
	"strconv"
	"time"

	"github.com/vicanso/dash"

	"github.com/labstack/echo"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/vars"
)

var ignoreKeys = []string{
	"Date",
}

// Dispatcher 对响应数据做缓存，复制等处理
func Dispatcher(client *cache.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			status := c.Get(vars.Status).(int)
			cr := c.Get(vars.Response).(*cache.Response)
			resp := c.Response()

			h := resp.Header()
			for k, values := range cr.Header {
				if dash.IncludesString(ignoreKeys, k) {
					continue
				}
				for _, v := range values {
					h.Add(k, v)
				}
			}
			// 如果数据不是从缓存中读取，都需要判断是否需要写入缓存
			// pass的都是不可能缓存
			if status != cache.Cacheable && status != cache.Pass {
				go func() {
					identity := c.Get(vars.Identity).([]byte)
					if cr.TTL == 0 {
						client.HitForPass(identity, vars.HitForPassTTL)
					} else {
						client.SaveResponse(identity, cr)
						client.Cacheable(identity, cr.TTL)
					}
				}()
			} else if status == cache.Cacheable {
				// 如果数据是读取缓存，有需要设置Age
				age := uint32(time.Now().Unix()) - cr.CreatedAt
				h.Set(vars.Age, strconv.Itoa(int(age)))
			}
			xStatus := ""
			switch status {
			case cache.Pass:
				xStatus = vars.Pass
			case cache.Fetching:
				xStatus = vars.Fetching
			case cache.HitForPass:
				xStatus = vars.HitForPass
			default:
				xStatus = vars.Cacheable
			}
			h.Set(vars.XStatus, xStatus)
			resp.WriteHeader(int(cr.StatusCode))
			_, err := resp.Write(cr.Body)
			return err
		}
	}
}
