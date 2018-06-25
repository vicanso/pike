package custommiddleware

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/proxy"
)

type (
	// Context custom context for pike
	Context struct {
		// server timing
		serverTiming *ServerTiming
		echo.Context
		// status 该请求的状态 fetching pass等
		status int
		// identity 该请求的标记
		identity []byte
		// director 该请求对应的director
		director *proxy.Director
		// resp 该请求的响应数据
		resp *cache.Response
		// fresh 是否fresh
		fresh bool
		// createdAt 创建时间
		createdAt time.Time
		// Skip 是否跳过中间件
		Skip bool
	}
	// BodyDumpResponseWriter dump writer
	BodyDumpResponseWriter struct {
		body    *bytes.Buffer
		headers http.Header
		code    int
	}
	// ServerTiming server timing
	ServerTiming struct {
		disabled           bool
		startedAt          int64
		use                int64
		getRequestStatsAt  int64
		getRequestStatsUse int64
		cacheFetchAt       int64
		cacheFetchUse      int64
		proxyAt            int64
		proxyUse           int64
	}
	// ProxyTarget defines the upstream target.
	ProxyTarget struct {
		Name string
		URL  *url.URL
	}
)

// Init 对Context重置
func (c *Context) Init() {
	c.status = 0
	c.identity = nil
	c.director = nil
	c.resp = nil
	c.fresh = false
	c.Skip = false
	c.createdAt = time.Now()
}

// DisableServerTiming 禁用server timing
func (c *Context) DisableServerTiming() {
	c.serverTiming.disabled = true
}

// WriteHeader write header
func (w *BodyDumpResponseWriter) WriteHeader(code int) {
	w.code = code
}

// Header get header
func (w *BodyDumpResponseWriter) Header() http.Header {
	return w.headers
}

// Write write buffer
func (w *BodyDumpResponseWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

var contextPool = sync.Pool{
	New: func() interface{} {
		return &Context{}
	},
}

var writerPool = sync.Pool{
	New: func() interface{} {
		return &BodyDumpResponseWriter{
			body:    new(bytes.Buffer),
			headers: make(http.Header),
		}
	},
}

var proxyTargetPool = sync.Pool{
	New: func() interface{} {
		return &ProxyTarget{}
	},
}

// NewContext 获取新的context
func NewContext(c echo.Context) *Context {
	pc := contextPool.Get().(*Context)
	pc.Init()
	pc.Context = c
	if pc.serverTiming == nil {
		pc.serverTiming = &ServerTiming{}
	}
	pc.serverTiming.Init()
	return pc
}

// ReleaseContext 释放Context
func ReleaseContext(pc *Context) {
	contextPool.Put(pc)
}

// NewBodyDumpResponseWriter 获取新的body dump responser write
func NewBodyDumpResponseWriter() *BodyDumpResponseWriter {
	w := writerPool.Get().(*BodyDumpResponseWriter)
	w.body.Reset()
	for k := range w.headers {
		delete(w.headers, k)
	}
	return w
}

// ReleaseBodyDumpResponseWriter 释放writer
func ReleaseBodyDumpResponseWriter(w *BodyDumpResponseWriter) {
	writerPool.Put(w)
}

// NewProxyTarget 获取新的proxy target
func NewProxyTarget() *ProxyTarget {
	return proxyTargetPool.Get().(*ProxyTarget)
}

// ReleaseProxyTarget 释放proxy target
func ReleaseProxyTarget(target *ProxyTarget) {
	proxyTargetPool.Put(target)
}

// Init 初始化server timing
func (st *ServerTiming) Init() {
	st.getRequestStatsAt = 0
	st.getRequestStatsUse = 0
	st.cacheFetchAt = 0
	st.cacheFetchUse = 0
	st.proxyAt = 0
	st.proxyUse = 0
	st.startedAt = time.Now().UnixNano()
}

// End 结束
func (st *ServerTiming) End() {
	if st.disabled {
		return
	}
	st.use = time.Now().UnixNano() - st.startedAt
}

// GetRequestStatusStart 开始获取get request status
func (st *ServerTiming) GetRequestStatusStart() {
	if st.disabled {
		return
	}
	st.getRequestStatsAt = time.Now().UnixNano()
}

// GetRequestStatusEnd 结束获取get request status
func (st *ServerTiming) GetRequestStatusEnd() {
	if st.disabled {
		return
	}
	st.getRequestStatsUse = time.Now().UnixNano() - st.getRequestStatsAt
}

// CacheFetchStart 开始获取缓存
func (st *ServerTiming) CacheFetchStart() {
	if st.disabled {
		return
	}
	st.cacheFetchAt = time.Now().UnixNano()
}

// CacheFetchEnd 获取缓存结束
func (st *ServerTiming) CacheFetchEnd() {
	if st.disabled {
		return
	}
	st.cacheFetchUse = time.Now().UnixNano() - st.cacheFetchAt
}

// ProxyStart 开始转发至backend
func (st *ServerTiming) ProxyStart() {
	if st.disabled {
		return
	}
	st.proxyAt = time.Now().UnixNano()
}

// ProxyEnd 转发处理完成
func (st *ServerTiming) ProxyEnd() {
	if st.disabled {
		return
	}
	st.proxyUse = time.Now().UnixNano() - st.proxyAt
}

// String 获取server timing的http header string
func (st *ServerTiming) String() string {
	if st.disabled {
		return ""
	}
	desList := []string{}
	ms := float64(time.Millisecond)
	use := st.use
	appendDesc := func(v int64, str string) {
		desc := fmt.Sprintf(str, strconv.FormatFloat(float64(v)/ms, 'f', -1, 64))
		desList = append(desList, desc)
	}
	if use != 0 {
		appendDesc(use, "0;dur=%s;desc=\"pike\"")
	}

	use = st.getRequestStatsUse
	if use != 0 {
		appendDesc(use, "1;dur=%s;desc=\"get request status\"")
	}

	use = st.cacheFetchUse
	if use != 0 {
		appendDesc(use, "2;dur=%s;desc=\"fetch cache\"")
	}

	use = st.proxyUse
	if use != 0 {
		appendDesc(use, "3;dur=%s;desc=\"fetch data from backend\"")
	}
	return strings.Join(desList, ",")
}
