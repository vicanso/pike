package pike

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/vicanso/pike/cache"
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

const (
	// ServerTimingPike pike
	ServerTimingPike = iota
	// ServerTimingInitialization init中间件
	ServerTimingInitialization
	// ServerTimingIdentifier identifier中间件
	ServerTimingIdentifier
	// ServerTimingDirectorPicker director picker中间件
	ServerTimingDirectorPicker
	// ServerTimingCacheFetcher cache fetcher中间件
	ServerTimingCacheFetcher
	// ServerTimingProxy proxy中间件
	ServerTimingProxy
	// ServerTimingHeaderSetter header setter中间件
	ServerTimingHeaderSetter
	// ServerTimingFreshChecker fresh checker中间件
	ServerTimingFreshChecker
	// ServerTimingDispatcher dispatcher中间件
	ServerTimingDispatcher
	// ServerTimingEnd server timing end
	ServerTimingEnd
)

var (
	// serverTimingDesList server timing的描述
	serverTimingDesList = []string{
		"0;dur=%s;desc=\"pike\"",
		"1;dur=%s;desc=\"init\"",
		"2;dur=%s;desc=\"identifier\"",
		"3;dur=%s;desc=\"director picker\"",
		"4;dur=%s;desc=\"cache fetcher\"",
		"5;dur=%s;desc=\"proxy\"",
		"6;dur=%s;desc=\"header setter\"",
		"7;dur=%s;desc=\"fresh checker\"",
		"7;dur=%s;desc=\"dispatcher\"",
	}
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
	// Context context
	Context struct {
		Request        *http.Request
		Response       *Response
		ResponseWriter http.ResponseWriter
		ServerTiming   *ServerTiming
		// Status 该请求的状态 fetching pass等
		Status int
		// Identity 该请求的标记
		Identity []byte
		// Director 该请求对应的director
		Director *Director
		// Resp 该请求的响应数据
		Resp *cache.Response
		// Fresh 是否fresh
		Fresh bool
		// CreatedAt 创建时间
		CreatedAt time.Time
	}
	// Middleware middleware function
	Middleware func(*Context, Next) error
	// Next next function
	Next func() error
	// ErrorHandler error handle function
	ErrorHandler func(error, *Context)
	// ServerTiming server timing
	ServerTiming struct {
		disabled      bool
		startedAt     int64
		startedAtList []int64
		useList       []int64
	}
	// Response http response
	Response struct {
		body      *bytes.Buffer
		headers   http.Header
		code      int
		Committed bool
	}
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

// NewContext 创新新的Context并重置相应的属性
func NewContext(req *http.Request) (c *Context) {
	c = contextPool.Get().(*Context)
	if c.ServerTiming == nil {
		c.ServerTiming = &ServerTiming{
			startedAtList: make([]int64, ServerTimingEnd),
			useList:       make([]int64, ServerTimingEnd),
			startedAt:     time.Now().UnixNano(),
		}
	} else {
		c.ServerTiming.Reset()
	}
	if c.Response == nil {
		c.Response = &Response{
			body:    new(bytes.Buffer),
			headers: make(http.Header),
			code:    http.StatusNotFound,
		}
	} else {
		c.Response.Reset()
	}
	c.Request = req
	c.Reset()
	return
}

// Reset 重置context
func (c *Context) Reset() {
	c.Status = 0
	c.Identity = nil
	c.Director = nil
	c.Resp = nil
	c.Fresh = false
	c.CreatedAt = time.Now()
}

// RealIP 客户端真实IP
func (c *Context) RealIP() string {
	ra := c.Request.RemoteAddr
	if ip := c.Request.Header.Get(HeaderXForwardedFor); ip != "" {
		ra = strings.Split(ip, ", ")[0]
	} else if ip := c.Request.Header.Get(HeaderXRealIP); ip != "" {
		ra = ip
	} else {
		ra, _, _ = net.SplitHostPort(ra)
	}
	return ra
}

// JSON 返回json
func (c *Context) JSON(data interface{}, status int) error {
	buf, err := json.Marshal(data)
	if err != nil {
		return err
	}
	resp := c.Response
	header := resp.Header()
	header.Set(HeaderContentType, JSONContent)
	header.Set(HeaderContentLength, strconv.Itoa(len(buf)))
	resp.WriteHeader(status)
	_, err = resp.Write(buf)
	return err
}

// Reset 重置
func (st *ServerTiming) Reset() {
	useList := st.useList
	for i := range useList {
		useList[i] = 0
	}
	st.startedAt = time.Now().UnixNano()
}

// Start 开始server timing的记录
func (st *ServerTiming) Start(index int) func() {
	if st.disabled || index <= ServerTimingPike || index >= ServerTimingEnd {
		return noop
	}
	startedAt := time.Now().UnixNano()
	return func() {
		st.useList[index] = time.Now().UnixNano() - startedAt
	}
}

// String 获取server timing的http header string
func (st *ServerTiming) String() string {
	if st.disabled {
		return ""
	}
	desList := []string{}
	ms := float64(time.Millisecond)
	// use := st.use
	appendDesc := func(v int64, str string) {
		desc := fmt.Sprintf(str, strconv.FormatFloat(float64(v)/ms, 'f', -1, 64))
		desList = append(desList, desc)
	}
	useList := st.useList
	useList[0] = time.Now().UnixNano() - st.startedAt

	for i, v := range st.useList {
		if v != 0 {
			appendDesc(v, serverTimingDesList[i])
		}
	}
	return strings.Join(desList, ",")
}

// WriteHeader write header
func (w *Response) WriteHeader(code int) {
	w.code = code
}

// Header get header
func (w *Response) Header() http.Header {
	return w.headers
}

// Write write buffer
func (w *Response) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

// Status get the response status
func (w *Response) Status() int {
	return w.code
}

// Size get the response size
func (w *Response) Size() int {
	return w.body.Len()
}

// Reset reset the response sturct
func (w *Response) Reset() {
	w.body.Reset()
	w.code = http.StatusNotFound
	w.Committed = false
	for k := range w.headers {
		delete(w.headers, k)
	}
}

// Bytes get the bytes of response
func (w *Response) Bytes() []byte {
	return w.body.Bytes()
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
