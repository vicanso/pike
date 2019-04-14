package middleware

import (
	"net/http"
	"regexp"
	"strconv"

	"github.com/vicanso/cod"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/df"
	"github.com/vicanso/pike/util"
)

var (
	noCacheReg = regexp.MustCompile(`no-cache|no-store|private`)
	sMaxAgeReg = regexp.MustCompile(`s-maxage=(\d+)`)
	maxAgeReg  = regexp.MustCompile(`max-age=(\d+)`)
)

// 根据Cache-Control的信息，获取s-maxage 或者max-age的值
func getCacheAge(header http.Header) int {
	// 如果有设置cookie，则为不可缓存
	if len(header.Get(cod.HeaderSetCookie)) != 0 {
		return 0
	}
	// 如果没有设置cache-control，则不可缓存
	cc := header.Get(cod.HeaderCacheControl)
	if len(cc) == 0 {
		return 0
	}

	// 如果设置不可缓存，返回0
	match := noCacheReg.MatchString(cc)
	if match {
		return 0
	}
	// 优先从s-maxage中获取
	var maxAge = 0
	result := sMaxAgeReg.FindStringSubmatch(cc)
	if len(result) == 2 {
		maxAge, _ = strconv.Atoi(result[1])
	} else {
		// 从max-age中获取缓存时间
		result = maxAgeReg.FindStringSubmatch(cc)
		if len(result) == 2 {
			maxAge, _ = strconv.Atoi(result[1])
		}
	}

	// 如果有设置了 age 字段，则最大缓存时长减少
	age := header.Get(df.HeaderAge)
	if age != "" {
		v, _ := strconv.Atoi(age)
		maxAge -= v
	}

	return maxAge
}

// NewCacheIdentifier create a cache identifier middleware
func NewCacheIdentifier() cod.Handler {
	identify := config.GetIdentity()
	fn := util.GetIdentity
	if identify != "" {
		fn = util.GenerateGetIdentity(identify)
	}
	return func(c *cod.Context) (err error) {
		// 如果非 GET HEAD请求，直接跳过
		method := c.Request.Method
		if method != http.MethodGet && method != http.MethodHead {
			c.Set(df.Status, cache.Pass)
			return c.Next()
		}

		key := fn(c.Request)
		hc := cache.GetHTTPCache(key)
		status := hc.Status
		// 如果是 fetch 的，则需要写缓存，在获取时会调用写锁
		if status == cache.Fetch {
			defer hc.Unlock()
		} else {
			defer hc.RUnlock()
		}
		c.Set(df.Status, status)
		c.Set(df.Cache, hc)
		err = c.Next()
		// 如果不是初始化状态，直接返回，无须处理后续
		if status != cache.Fetch {
			return
		}

		maxAge := 0
		if err == nil {
			maxAge = getCacheAge(c.Header())
		}
		if maxAge <= 0 {
			hc.HitForPass()
		} else {
			hc.Cacheable(maxAge, c)
			// 如果设置状态成功，则清除现有数据
			// 在responder中间件重新生成
			if hc.Status == cache.Cacheable {
				c.BodyBuffer = nil
				c.StatusCode = 0
				c.ResetHeader()
			}
		}
		return
	}
}
