package pike

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/vicanso/pike/vars"
)

/*
Package pike for high performance http cache server.
*/

// TODO recover

const (
	defaultReadTimeout  = 10 * time.Second
	defaultWriteTimeout = 10 * time.Second
	// HeaderXForwardedFor x-forwarder-for header
	HeaderXForwardedFor = "X-Forwarded-For"
	// HeaderXRealIP x-real-ip header
	HeaderXRealIP = "X-Real-IP"
	// HeaderIfModifiedSince if-modified-since header
	HeaderIfModifiedSince = "If-Modified-Since"
	// HeaderIfNoneMatch if-none-match header
	HeaderIfNoneMatch = "If-None-Match"
	// HeaderContentEncoding content-encoding header
	HeaderContentEncoding = "Content-Encoding"
	// HeaderETag http response etag
	HeaderETag = "ETag"
	// HeaderLastModified last-modified header
	HeaderLastModified = "Last-Modified"
	// HeaderCacheControl http response cache control header
	HeaderCacheControl = "Cache-Control"
	// HeaderContentLength content-length header
	HeaderContentLength = "Content-Length"
	// HeaderSetCookie http set cookie header
	HeaderSetCookie = "Set-Cookie"
	// HeaderContentType http content-type header
	HeaderContentType = "Content-Type"
	// HeaderServerTiming http response server timing
	HeaderServerTiming = "Server-Timing"
	// HeaderAge http age header
	HeaderAge = "Age"
	// HeaderAcceptEncoding http accept-encoding header
	HeaderAcceptEncoding = "Accept-Encoding"
	// HeaderXStatus http x-status response header
	HeaderXStatus = "X-Status"
	// GzipEncoding gzip encoding
	GzipEncoding = "gzip"
	// BrEncoding br encoding
	BrEncoding = "br"
	// JSONContent json content-type
	JSONContent = "application/json; charset=utf-8"
)

var (
	noop = func() {}
	// NoopNext no op next function
	NoopNext = func() error {
		return nil
	}
	// ErrTargetURLNotInit target url map未初始化
	ErrTargetURLNotInit = NewHTTPError(http.StatusNotImplemented, "target url is not init")
	// ErrParseBackendURLFail 生成backend url失败
	ErrParseBackendURLFail = NewHTTPError(http.StatusNotImplemented, "parse backend url fail")
)

type (
	// Pike app instance of pike
	Pike struct {
		middleware   []Middleware
		server       *http.Server
		ReadTimeout  time.Duration
		WriteTimeout time.Duration
		ErrorHandler ErrorHandler
	}

	// Middleware middleware function
	Middleware func(*Context, Next) error
	// Next next function
	Next func() error
	// ErrorHandler error handle function
	ErrorHandler func(error, *Context)

	// HTTPError represents an error that occurred while handling a request. (copy from echo)
	HTTPError struct {
		Code     int
		Message  interface{}
		Internal error // Stores the error returned by an external dependency
	}
)

var contextPool = sync.Pool{
	New: func() interface{} {
		return &Context{}
	},
}

// NewHTTPError creates a new HTTPError instance.
func NewHTTPError(code int, message ...interface{}) *HTTPError {
	he := &HTTPError{Code: code, Message: http.StatusText(code)}
	if len(message) > 0 {
		he.Message = message[0]
	}
	return he
}

// Error makes it compatible with `error` interface.
func (he *HTTPError) Error() string {
	return fmt.Sprintf("code=%d, message=%v", he.Code, he.Message)
}

// New 创建新的Pike实例
func New() (p *Pike) {
	p = &Pike{
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
	}
	p.ErrorHandler = p.DefaultErrorHanddler
	return
}

// Use add middleware function
func (p *Pike) Use(mids ...Middleware) {
	p.middleware = append(p.middleware, mids...)
}

// handle http server handler function
func (p *Pike) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mids := p.middleware
	c := NewContext(r)
	defer func() {
		c.Request = nil
		c.ResponseWriter = nil
		contextPool.Put(c)
	}()
	c.ResponseWriter = w
	max := len(mids)
	index := -1
	var next Next
	next = func() error {
		index++
		if index >= max {
			return nil
		}
		return mids[index](c, next)
	}
	err := next()
	if err != nil {
		p.ErrorHandler(err, c)
		return
	}
	res := c.Response
	res.Committed = true
	header := w.Header()
	for field, values := range res.Header() {
		for _, value := range values {
			header.Set(field, value)
		}
	}
	body := res.Bytes()
	header.Set(HeaderContentLength, strconv.Itoa(len(body)))
	w.WriteHeader(res.Status())
	w.Write(body)
}

// ListenAndServe the http function ListenAndServe
func (p *Pike) ListenAndServe(addr string) error {
	p.server = &http.Server{
		Addr:           addr,
		Handler:        p,
		ReadTimeout:    p.ReadTimeout,
		WriteTimeout:   p.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}
	log.Infof("pike(%s) will listen on %s", vars.Version, addr)
	return p.server.ListenAndServe()
}

// Close close the http server
func (p *Pike) Close() error {
	if p.server == nil {
		return nil
	}
	return p.server.Close()
}

// DefaultErrorHanddler 默认的出错处理
func (p *Pike) DefaultErrorHanddler(err error, c *Context) {
	if c.Response.Committed {
		return
	}
	code := http.StatusInternalServerError
	msg := err.Error()
	if he, ok := err.(*HTTPError); ok {
		code = he.Code
		// TODO 支持多种类型的出错
		msg = he.Message.(string)
	}
	c.ResponseWriter.WriteHeader(code)
	c.ResponseWriter.Write([]byte(msg))
}
