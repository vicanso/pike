package customMiddleware

import (
	"bytes"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"strings"

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
			c.Set(vars.Body, writer.body.Bytes())
			c.Set(vars.Code, writer.code)
			c.Set(vars.Header, writer.headers)
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
