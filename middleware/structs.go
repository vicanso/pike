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
		// Debug 是否debug
		Debug bool
	}
	// BodyDumpResponseWriter dump writer
	BodyDumpResponseWriter struct {
		body    *bytes.Buffer
		headers http.Header
		code    int
	}
	// ServerTiming server timing
	ServerTiming struct {
		disabled      bool
		startedAt     int64
		startedAtList []int64
		useList       []int64
	}
	// ProxyTarget defines the upstream target.
	ProxyTarget struct {
		Name string
		URL  *url.URL
	}
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
)

// Init 对Context重置
func (c *Context) Init() {
	c.status = 0
	c.identity = nil
	c.director = nil
	c.resp = nil
	c.fresh = false
	c.Skip = false
	c.Debug = false
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
		pc.serverTiming = NewServerTiming()
	} else {
		useList := pc.serverTiming.useList
		for i := range useList {
			useList[i] = 0
		}
		pc.serverTiming.startedAt = time.Now().UnixNano()
	}
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

// NewServerTiming 创建新的server timing
func NewServerTiming() *ServerTiming {
	return &ServerTiming{
		startedAtList: make([]int64, ServerTimingEnd),
		useList:       make([]int64, ServerTimingEnd),
		startedAt:     time.Now().UnixNano(),
	}
}

// Start 开始
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
