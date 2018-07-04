package custommiddleware

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/util"
	"github.com/vicanso/pike/vars"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// 根据echo proxy middleware 调整而来

type (
	// ProxyConfig defines the config for Proxy middleware.
	ProxyConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper middleware.Skipper

		// Rewrites defines URL path rewrite rules. The values captured in asterisk can be
		// retrieved by index e.g. $1, $2 and so on.
		// Examples:
		// "/old":              "/new",
		// "/api/*":            "/$1",
		// "/js/*":             "/public/javascripts/$1",
		// "/users/*/orders/*": "/user/$1/order/$2",
		Rewrites []string

		// ETag 是否生成ETag
		ETag bool

		rewriteRegexp map[*regexp.Regexp]string
		// Timeout proxy的连接超时
		Timeout time.Duration
	}
)

const (
	defaultTimeout = 10 * time.Second
)

var (
	noCacheReg = regexp.MustCompile(`no-cache|no-store|private`)
	sMaxAgeReg = regexp.MustCompile(`s-maxage=(\d+)`)
	maxAgeReg  = regexp.MustCompile(`max-age=(\d+)`)
)

// genETag 获取数据对应的ETag
func genETag(buf []byte) string {
	size := len(buf)
	if size == 0 {
		return "\"0-2jmj7l5rSw0yVb_vlWAYkK_YBwk=\""
	}
	h := sha1.New()
	h.Write(buf)
	hash := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("\"%x-%s\"", size, hash)
}

func captureTokens(pattern *regexp.Regexp, input string) *strings.Replacer {
	groups := pattern.FindAllStringSubmatch(input, -1)
	if groups == nil {
		return nil
	}
	values := groups[0][1:]
	replace := make([]string, 2*len(values))
	for i, v := range values {
		j := 2 * i
		replace[j] = "$" + strconv.Itoa(i+1)
		replace[j+1] = v
	}
	return strings.NewReplacer(replace...)
}

func rewrite(rewriteRegexp map[*regexp.Regexp]string, req *http.Request) {
	for k, v := range rewriteRegexp {
		replacer := captureTokens(k, req.URL.Path)
		if replacer != nil {
			req.URL.Path = replacer.Replace(v)
		}
	}
}

func proxyHTTP(t *ProxyTarget, transport *http.Transport) http.Handler {
	p := httputil.NewSingleHostReverseProxy(t.URL)
	if transport != nil {
		p.Transport = transport
	}
	return p
}

// 根据Cache-Control的信息，获取s-maxage或者max-age的值
func getCacheAge(header http.Header) uint16 {
	// 如果有设置cookie，则为不可缓存
	if len(header[vars.SetCookie]) != 0 {
		return 0
	}
	// 如果没有设置cache-control，则不可缓存
	if len(header[vars.CacheControl]) == 0 {
		return 0
	}

	cacheControl := []byte(header[vars.CacheControl][0])
	// 如果为空，则不可缓存
	if len(cacheControl) == 0 {
		return 0
	}
	// 如果设置不可缓存，返回0
	match := noCacheReg.Match(cacheControl)
	if match {
		return 0
	}

	// 优先从s-maxage中获取
	result := sMaxAgeReg.FindSubmatch(cacheControl)
	if len(result) == 2 {
		maxAge, _ := strconv.Atoi(string(result[1]))
		return uint16(maxAge)
	}

	// 从max-age中获取缓存时间
	result = maxAgeReg.FindSubmatch(cacheControl)
	if len(result) != 2 {
		return 0
	}
	maxAge, _ := strconv.Atoi(string(result[1]))
	return uint16(maxAge)
}

// Proxy returns a Proxy middleware with config.
func Proxy(config ProxyConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = middleware.DefaultSkipper
	}
	config.rewriteRegexp = util.GetRewriteRegexp(config.Rewrites)
	timeout := defaultTimeout
	if config.Timeout > 0 {
		timeout = config.Timeout
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if config.Skipper(c) {
				return next(c)
			}
			pc := c.(*Context)
			done := pc.serverTiming.Start(ServerTimingProxy)
			// 如果已获取到数据，则不需要proxy获取(已从cache中获取)
			if pc.resp != nil {
				done()
				return next(pc)
			}
			// 获取director
			director := pc.director
			if director == nil {
				done()
				return vars.ErrDirectorNotFound
			}
			// 从director中选择可用的backend
			backend := director.Select(pc)
			if len(backend) == 0 {
				done()
				return vars.ErrNoBackendAvaliable
			}

			req := pc.Request()

			// Rewrite
			rewrite(config.rewriteRegexp, req)
			if director.RewriteRegexp != nil {
				rewrite(director.RewriteRegexp, req)
			}

			reqHeader := req.Header
			targetURL, err := director.GetTargetURL(&backend)
			if err != nil {
				done()
				return err
			}

			// Proxy
			writer := NewBodyDumpResponseWriter()
			defer ReleaseBodyDumpResponseWriter(writer)

			// proxy时为了避免304的出现，因此调用时临时删除header
			ifModifiedSince := reqHeader.Get(echo.HeaderIfModifiedSince)
			ifNoneMatch := reqHeader.Get(vars.IfNoneMatch)
			if len(ifModifiedSince) != 0 {
				reqHeader.Del(echo.HeaderIfModifiedSince)
			}
			if len(ifNoneMatch) != 0 {
				reqHeader.Del(vars.IfNoneMatch)
			}
			proxyDone := make(chan bool)

			go func() {
				// 在proxy http之后则立即release
				tgt := NewProxyTarget()
				tgt.Name = director.Name
				tgt.URL = targetURL
				proxyHTTP(tgt, director.Transport).ServeHTTP(writer, req)
				ReleaseProxyTarget(tgt)
				proxyDone <- true
			}()
			select {
			case <-proxyDone:
				close(proxyDone)
			case <-time.After(timeout):
				return vars.ErrGatewayTimeout
			}

			if len(ifModifiedSince) != 0 {
				reqHeader.Set(echo.HeaderIfModifiedSince, ifModifiedSince)
			}
			if len(ifNoneMatch) != 0 {
				reqHeader.Set(vars.IfNoneMatch, ifNoneMatch)
			}
			headers := writer.headers
			ttl := getCacheAge(headers)
			body := writer.body.Bytes()
			if config.ETag {
				eTagValue := util.GetHeaderValue(headers, vars.ETag)
				if len(eTagValue) == 0 {
					headers[vars.ETag] = []string{
						genETag(body),
					}
				}
			}
			cr := &cache.Response{
				CreatedAt:  uint32(time.Now().Unix()),
				TTL:        ttl,
				StatusCode: uint16(writer.code),
				Header:     headers,
			}
			contentEncoding := headers[echo.HeaderContentEncoding]
			if len(contentEncoding) == 0 {
				cr.Body = body
			} else {
				switch contentEncoding[0] {
				case vars.GzipEncoding:
					cr.GzipBody = body
				case vars.BrEncoding:
					cr.BrBody = body
				default:
					return vars.ErrContentEncodingNotSupport
				}
			}
			pc.resp = cr
			done()
			return next(pc)
		}
	}
}
