package custommiddleware

import (
	"bytes"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/proxy"
)

type (
	// Context custom context for pike
	Context struct {
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
	}
	// BodyDumpResponseWriter dump writer
	BodyDumpResponseWriter struct {
		body    *bytes.Buffer
		headers http.Header
		code    int
	}
)

// Init 对Context重置
func (c *Context) Init() {
	c.status = 0
	c.identity = nil
	c.director = nil
	c.resp = nil
	c.fresh = false
	c.createdAt = time.Now()
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

// NewContext 获取新的context
func NewContext(c echo.Context) *Context {
	pc := contextPool.Get().(*Context)
	pc.Init()
	pc.Context = c
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
