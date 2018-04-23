package customMiddleware

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"

	"github.com/vicanso/dash"

	"github.com/labstack/echo"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/vars"
)

var ignoreKeys = []string{
	"Date",
}

const hitForPassTTL = 600

// 根据Cache-Control的信息，获取s-maxage或者max-age的值
func getCacheAge(cacheControl []byte) uint16 {
	// cacheControl := header.PeekBytes(vars.CacheControl)
	if len(cacheControl) == 0 {
		return 0
	}
	// 如果设置不可缓存，返回0
	reg, _ := regexp.Compile(`no-cache|no-store|private`)
	match := reg.Match(cacheControl)
	if match {
		return 0
	}

	// 优先从s-maxage中获取
	reg, _ = regexp.Compile(`s-maxage=(\d+)`)
	result := reg.FindSubmatch(cacheControl)
	if len(result) == 2 {
		maxAge, _ := strconv.Atoi(string(result[1]))
		return uint16(maxAge)
	}

	// 从max-age中获取缓存时间
	reg, _ = regexp.Compile(`max-age=(\d+)`)
	result = reg.FindSubmatch(cacheControl)
	if len(result) != 2 {
		return 0
	}
	maxAge, _ := strconv.Atoi(string(result[1]))
	return uint16(maxAge)
}

// Dispatcher 对响应数据做缓存，复制等处理
func Dispatcher(client *cache.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			status := c.Get(vars.Status).(int)
			identity := c.Get(vars.Identity).([]byte)
			code := c.Get(vars.Code).(int)
			body := c.Get(vars.Body).([]byte)
			header := c.Get(vars.Header).(http.Header)
			resp := c.Response()

			h := resp.Header()
			for k, values := range header {
				if dash.IncludesString(ignoreKeys, k) {
					continue
				}
				for _, v := range values {
					h.Add(k, v)
				}
			}
			if status == cache.Fetching {
				cacheControl := h.Get(vars.CacheControl)
				cacheAge := getCacheAge([]byte(cacheControl))
				go func() {
					if cacheAge == 0 {
						client.HitForPass(identity, hitForPassTTL)
					} else {
						buf, err := json.Marshal(header)
						if err != nil {
							client.HitForPass(identity, hitForPassTTL)
							return
						}
						cr := &cache.Response{
							StatusCode: uint16(code),
							TTL:        cacheAge,
							Header:     buf,
							Body:       body,
						}
						client.SaveResponse(identity, cr)
						client.Cacheable(identity, cacheAge)
					}
				}()
			}
			resp.WriteHeader(code)
			_, err := resp.Write(body)
			return err
		}
	}
}
