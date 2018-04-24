package customMiddleware

import (
	"bytes"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/proxy"
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

		// Rewrite defines URL path rewrite rules. The values captured in asterisk can be
		// retrieved by index e.g. $1, $2 and so on.
		// Examples:
		// "/old":              "/new",
		// "/api/*":            "/$1",
		// "/js/*":             "/public/javascripts/$1",
		// "/users/*/orders/*": "/user/$1/order/$2",
		Rewrite map[string]string

		rewriteRegex map[*regexp.Regexp]string
	}

	// ProxyTarget defines the upstream target.
	ProxyTarget struct {
		Name string
		URL  *url.URL
	}
	bodyDumpResponseWriter struct {
		body    *bytes.Buffer
		headers http.Header
		code    int
		http.ResponseWriter
	}
)

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

func proxyHTTP(t *ProxyTarget) http.Handler {
	return httputil.NewSingleHostReverseProxy(t.URL)
}

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

// ProxyWithConfig returns a Proxy middleware with config.
// See: `Proxy()`
func ProxyWithConfig(config ProxyConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = middleware.DefaultSkipper
	}
	config.rewriteRegex = map[*regexp.Regexp]string{}

	// Initialize
	for k, v := range config.Rewrite {
		k = strings.Replace(k, "*", "(\\S*)", -1)
		config.rewriteRegex[regexp.MustCompile(k)] = v
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if config.Skipper(c) {
				return next(c)
			}

			// 如果已获取到数据，则不需要proxy(从cache中获取)
			if c.Get(vars.Response) != nil {
				return next(c)
			}
			// 选择director
			d := c.Get(vars.Director)
			if d == nil {
				return vars.ErrDirectorNotFound
			}
			director := d.(*proxy.Director)
			// 从director中选择可用的backend
			backend := director.Select(c)
			if len(backend) == 0 {
				return vars.ErrNoBackendAvaliable
			}

			req := c.Request()
			targetURL, _ := url.Parse(backend)
			tgt := &ProxyTarget{
				Name: director.Name,
				URL:  targetURL,
			}

			// Rewrite
			for k, v := range config.rewriteRegex {
				replacer := captureTokens(k, req.URL.Path)
				if replacer != nil {
					req.URL.Path = replacer.Replace(v)
				}
			}

			// Fix header
			if req.Header.Get(echo.HeaderXRealIP) == "" {
				req.Header.Set(echo.HeaderXRealIP, c.RealIP())
			}
			if req.Header.Get(echo.HeaderXForwardedProto) == "" {
				req.Header.Set(echo.HeaderXForwardedProto, c.Scheme())
			}

			// Proxy
			writer := &bodyDumpResponseWriter{
				body:    new(bytes.Buffer),
				headers: make(http.Header),
			}
			proxyHTTP(tgt).ServeHTTP(writer, req)
			cacheControl := writer.headers[vars.CacheControl]
			var ttl uint16
			if len(cacheControl) != 0 {
				// cache control 只会有一个http header
				ttl = getCacheAge([]byte(cacheControl[0]))
			}
			cr := &cache.Response{
				CreatedAt:  uint32(time.Now().Unix()),
				TTL:        ttl,
				StatusCode: uint16(writer.code),
				Header:     writer.headers,
				Body:       writer.body.Bytes(),
			}
			c.Set(vars.Response, cr)
			return next(c)
		}
	}
}

func (w *bodyDumpResponseWriter) WriteHeader(code int) {
	w.code = code
}

func (w *bodyDumpResponseWriter) Header() http.Header {
	return w.headers
}

func (w *bodyDumpResponseWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}
