package middleware

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/vicanso/pike/pike"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/util"
)

type (
	// ProxyConfig defines the config for Proxy middleware.
	ProxyConfig struct {

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
	// ProxyTarget defines the upstream target.
	ProxyTarget struct {
		Name string
		URL  *url.URL
	}
)

const (
	defaultTimeout = 10 * time.Second
)

var (
	noCacheReg      = regexp.MustCompile(`no-cache|no-store|private`)
	sMaxAgeReg      = regexp.MustCompile(`s-maxage=(\d+)`)
	maxAgeReg       = regexp.MustCompile(`max-age=(\d+)`)
	proxyTargetPool = sync.Pool{
		New: func() interface{} {
			return &ProxyTarget{}
		},
	}
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

// byteSliceToString converts a []byte to string without a heap allocation.
func byteSliceToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// 根据Cache-Control的信息，获取s-maxage或者max-age的值
func getCacheAge(header http.Header) uint16 {
	// 如果有设置cookie，则为不可缓存
	if len(header.Get(pike.HeaderSetCookie)) != 0 {
		return 0
	}
	// 如果没有设置cache-control，则不可缓存
	cc := header.Get(pike.HeaderCacheControl)
	if len(cc) == 0 {
		return 0
	}

	cacheControl := []byte(cc)
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
		maxAge, _ := strconv.Atoi(byteSliceToString(result[1]))
		return uint16(maxAge)
	}

	// 从max-age中获取缓存时间
	result = maxAgeReg.FindSubmatch(cacheControl)
	if len(result) != 2 {
		return 0
	}
	maxAge, _ := strconv.Atoi(byteSliceToString(result[1]))
	return uint16(maxAge)
}

// Proxy returns a Proxy middleware with config.
func Proxy(config ProxyConfig) pike.Middleware {
	config.rewriteRegexp = util.GetRewriteRegexp(config.Rewrites)
	timeout := defaultTimeout
	if config.Timeout > 0 {
		timeout = config.Timeout
	}

	return func(c *pike.Context, next pike.Next) error {
		done := c.ServerTiming.Start(pike.ServerTimingProxy)
		// 如果已获取到数据，则不需要proxy获取(已从cache中获取)
		if c.Resp != nil {
			done()
			return next()
		}
		// 获取director
		director := c.Director
		if director == nil {
			done()
			return ErrDirectorNotFound
		}
		// 从director中选择可用的backend
		backend := director.Select(c)
		if len(backend) == 0 {
			done()
			return ErrNoBackendAvaliable
		}

		req := c.Request

		// Rewrite
		rewrite(config.rewriteRegexp, req)
		if director.RewriteRegexp != nil {
			rewrite(director.RewriteRegexp, req)
		}
		reqHeader := req.Header

		// 添加自定义请求头
		if director.RequestHeaderMap != nil {
			for k, v := range director.RequestHeaderMap {
				reqHeader.Add(k, v)
			}
		}

		targetURL, err := director.GetTargetURL(&backend)
		if err != nil {
			done()
			return err
		}

		// Proxy
		writer := pike.NewResponse()

		// proxy时为了避免304的出现，因此调用时临时删除header
		ifModifiedSince := reqHeader.Get(pike.HeaderIfModifiedSince)
		ifNoneMatch := reqHeader.Get(pike.HeaderIfNoneMatch)
		if len(ifModifiedSince) != 0 {
			reqHeader.Del(pike.HeaderIfModifiedSince)
		}
		if len(ifNoneMatch) != 0 {
			reqHeader.Del(pike.HeaderIfNoneMatch)
		}
		proxyDone := make(chan bool)

		go func() {
			// 在proxy http之后则立即release
			tgt := proxyTargetPool.Get().(*ProxyTarget)
			tgt.Name = director.Name
			tgt.URL = targetURL
			proxyHTTP(tgt, director.Transport).ServeHTTP(writer, req)
			proxyTargetPool.Put(tgt)
			proxyDone <- true
		}()
		select {
		case <-proxyDone:
			close(proxyDone)
		case <-time.After(timeout):
			return ErrGatewayTimeout
		}

		if len(ifModifiedSince) != 0 {
			reqHeader.Set(pike.HeaderIfModifiedSince, ifModifiedSince)
		}
		if len(ifNoneMatch) != 0 {
			reqHeader.Set(pike.HeaderIfNoneMatch, ifNoneMatch)
		}

		headers := writer.Header()
		if director.HeaderMap != nil {
			for k, v := range director.HeaderMap {
				headers.Add(k, v)
			}
		}

		ttl := getCacheAge(headers)
		body := writer.Bytes()
		status := writer.Status()
		// 如果是出错返回，则不需要生成ETag
		// 因为出错的数据都不做缓存，提高性能
		if config.ETag && (status >= http.StatusOK && status < http.StatusBadRequest) {
			eTagValue := util.GetHeaderValue(headers, pike.HeaderETag)
			if len(eTagValue) == 0 {
				headers.Set(pike.HeaderETag, genETag(body))
			}
		}
		cr := &cache.Response{
			CreatedAt:  uint32(time.Now().Unix()),
			TTL:        ttl,
			StatusCode: uint16(status),
			Header:     headers,
		}
		contentEncoding := headers.Get(pike.HeaderContentEncoding)
		if len(contentEncoding) == 0 {
			cr.Body = body
		} else {
			switch contentEncoding {
			case cache.GzipEncoding:
				cr.GzipBody = body
			case cache.BrEncoding:
				cr.BrBody = body
			default:
				return ErrContentEncodingNotSupport
			}
		}
		c.Resp = cr
		done()
		return next()
	}
}
