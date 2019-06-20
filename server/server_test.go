package server

import (
	"bytes"
	"net"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/vicanso/cod"

	"github.com/stretchr/testify/assert"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/df"
	"github.com/vicanso/pike/stats"
	"github.com/vicanso/pike/upstream"
)

func newTestServer() (ln net.Listener, err error) {
	ln, err = net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		return
	}

	d := cod.New()

	inc := func(p *int32) *bytes.Buffer {
		v := atomic.AddInt32(p, 1)
		return bytes.NewBufferString(strconv.Itoa(int(v)))
	}

	genBuffer := func(size int) *bytes.Buffer {
		buf := new(bytes.Buffer)
		for i := 0; i < size; i++ {
			buf.WriteString("a")
		}
		return buf
	}

	// 响应未压缩
	notCompressHandler := func(c *cod.Context) error {
		c.SetHeader("Content-Type", "text/html")
		c.BodyBuffer = genBuffer(4096)
		return nil
	}
	// 响应数据已压缩
	compressHandler := func(c *cod.Context) error {
		c.SetHeader("Content-Type", "text/html")
		c.SetHeader("Content-Encoding", "gzip")
		buf, _ := cache.Gzip(genBuffer(4096).Bytes())

		c.BodyBuffer = bytes.NewBuffer(buf)
		return nil
	}

	setCacheNext := func(c *cod.Context) error {
		c.CacheMaxAge("10s")
		return c.Next()
	}

	d.GET("/ping", func(c *cod.Context) error {
		c.BodyBuffer = bytes.NewBufferString("pong")
		return nil
	})

	var postResponseID int32
	d.POST("/post", func(c *cod.Context) error {
		c.BodyBuffer = inc(&postResponseID)
		return nil
	})

	// 非文本类数据
	d.GET("/image-cache", func(c *cod.Context) error {
		c.CacheMaxAge("10s")
		c.SetHeader("Content-Type", "image/png")
		c.BodyBuffer = genBuffer(4096)
		return nil
	})

	d.POST("/post-not-compress", notCompressHandler)
	d.GET("/get-not-compress", notCompressHandler)

	d.POST("/post-compress", compressHandler)
	d.POST("/get-compress", compressHandler)

	d.GET("/get-cache-not-compress", setCacheNext, notCompressHandler)
	d.GET("/get-cache-compress", setCacheNext, compressHandler)

	d.GET("/get-without-etag", notCompressHandler)

	d.GET("/get-with-etag", func(c *cod.Context) error {
		c.SetHeader("ETag", `"123"`)
		return notCompressHandler(c)
	})

	var noCacheResponseID int32
	d.GET("/no-cache", func(c *cod.Context) error {
		c.BodyBuffer = inc(&noCacheResponseID)
		return nil
	})

	go d.Serve(ln)

	return
}

func TestNewServer(t *testing.T) {

	assert := assert.New(t)

	ln, err := newTestServer()
	assert.Nil(err)
	defer ln.Close()

	up := upstream.New(config.BackendConfig{
		Name: "test",
		Ping: "/ping",
		Backends: []string{
			"http://" + ln.Addr().String(),
		},
	}, nil)
	up.Server.DoHealthCheck()

	upstreams := make(upstream.Upstreams, 0)
	upstreams = append(upstreams, up)
	director := &upstream.Director{
		Upstreams: upstreams,
	}

	passStatus := cache.GetStatusDesc(cache.Pass)
	cacheableStatus := cache.GetStatusDesc(cache.Cacheable)
	fetchStatus := cache.GetStatusDesc(cache.Fetch)
	hitForPassStatus := cache.GetStatusDesc(cache.HitForPass)

	dsp := cache.NewDispatcher(cache.Options{
		Size:              10,
		ZoneSize:          10,
		CompressMinLength: 1024,
		TextFilter:        regexp.MustCompile("text|javascript|json"),
	})

	basicConfig, err := config.NewBasicConfig("")
	assert.Nil(err)
	err = basicConfig.ReadConfig()
	assert.Nil(err)
	d := NewServer(Options{
		BasicConfig: basicConfig,
		Director:    director,
		Dispatcher:  dsp,
		Stats:       stats.New(),
	})

	newRequest := func(method, url string) *http.Request {
		req := httptest.NewRequest(method, url, nil)
		req.Header.Set(cod.HeaderAcceptEncoding, "gzip")
		return req
	}

	groupHandle := func(handle func()) {
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				handle()
				wg.Done()
			}()
		}
		wg.Wait()
	}

	checkStatusList := func(t *testing.T, statusList []string, onceStatus, otherStatus string) {
		onceCount := 0
		otherCount := 0
		for _, status := range statusList {
			switch status {
			case onceStatus:
				onceCount++
			case otherStatus:
				otherCount++
			default:
				t.Fatalf("unexpected status:%s", status)
			}
		}
		assert.Equal(1, onceCount, onceStatus+" should appear once")
		assert.Equal(len(statusList)-1, otherCount, otherStatus+" should appear "+strconv.Itoa(len(statusList)-1))
	}

	t.Run("post", func(t *testing.T) {
		groupHandle(func() {
			req := newRequest("POST", "/post")
			resp := httptest.NewRecorder()
			d.ServeHTTP(resp, req)
			h := resp.Header()

			assert.Equal(resp.Code, 200)
			assert.Equal(h.Get(df.HeaderStatus), passStatus)
		})
	})

	t.Run("image cache", func(t *testing.T) {
		statusList := make([]string, 0)
		var mu sync.Mutex
		groupHandle(func() {
			req := newRequest("GET", "/image-cache")
			resp := httptest.NewRecorder()
			d.ServeHTTP(resp, req)
			h := resp.Header()

			status := h.Get(df.HeaderStatus)
			mu.Lock()
			defer mu.Unlock()
			statusList = append(statusList, status)

			assert.Equal(resp.Code, 200)
			assert.Equal(resp.Body.Len(), 4096)
			assert.Equal(h.Get(cod.HeaderContentEncoding), "")
			assert.Equal(h.Get(cod.HeaderCacheControl), "public, max-age=10")

		})
		checkStatusList(t, statusList, fetchStatus, cacheableStatus)
	})

	t.Run("post-not-compress", func(t *testing.T) {
		groupHandle(func() {
			req := newRequest("POST", "/post-not-compress")
			resp := httptest.NewRecorder()
			d.ServeHTTP(resp, req)
			h := resp.Header()

			assert.Equal(resp.Code, 200)
			assert.Equal(resp.Body.Len(), 43)
			assert.NotEqual(h.Get(cod.HeaderETag), "")
			assert.Equal(h.Get(cod.HeaderContentEncoding), "gzip")

		})
	})

	t.Run("get not compress", func(t *testing.T) {
		statusList := make([]string, 0)
		var mu sync.Mutex
		groupHandle(func() {
			req := newRequest("GET", "/get-not-compress")
			resp := httptest.NewRecorder()
			d.ServeHTTP(resp, req)
			h := resp.Header()
			mu.Lock()
			defer mu.Unlock()
			status := h.Get(df.HeaderStatus)
			statusList = append(statusList, status)

			assert.Equal(resp.Code, 200)
			assert.Equal(resp.Body.Len(), 43)
			assert.NotEqual(h.Get(cod.HeaderETag), "")
			assert.Equal(h.Get(cod.HeaderContentEncoding), "gzip")
		})
		checkStatusList(t, statusList, fetchStatus, hitForPassStatus)
	})

	t.Run("get cache not compress", func(t *testing.T) {
		statusList := make([]string, 0)
		var mu sync.Mutex
		groupHandle(func() {
			req := newRequest("GET", "/get-cache-not-compress")
			resp := httptest.NewRecorder()
			d.ServeHTTP(resp, req)
			h := resp.Header()
			status := h.Get(df.HeaderStatus)
			mu.Lock()
			defer mu.Unlock()
			statusList = append(statusList, status)

			assert.Equal(resp.Code, 200)
			assert.Equal(resp.Body.Len(), 43)
			assert.NotEqual(h.Get(cod.HeaderETag), "")
			assert.Equal(h.Get(cod.HeaderContentEncoding), "gzip")
		})
		checkStatusList(t, statusList, fetchStatus, cacheableStatus)
	})

	t.Run("get cache compress", func(t *testing.T) {
		statusList := make([]string, 0)
		var mu sync.Mutex
		groupHandle(func() {
			req := newRequest("GET", "/get-cache-compress")
			resp := httptest.NewRecorder()
			d.ServeHTTP(resp, req)
			h := resp.Header()
			status := h.Get(df.HeaderStatus)
			mu.Lock()
			defer mu.Unlock()
			statusList = append(statusList, status)
			assert.Equal(resp.Code, 200)
			assert.Equal(resp.Body.Len(), 43)
			assert.NotEqual(h.Get(cod.HeaderETag), "")
			assert.Equal(h.Get(cod.HeaderContentEncoding), "gzip")
		})
		checkStatusList(t, statusList, fetchStatus, cacheableStatus)
	})

	t.Run("get without etag", func(t *testing.T) {
		groupHandle(func() {
			req := newRequest("GET", "/get-without-etag")
			resp := httptest.NewRecorder()
			d.ServeHTTP(resp, req)
			h := resp.Header()
			assert.Equal(resp.Code, 200)
			assert.NotEqual(h.Get(cod.HeaderETag), "")
		})
	})

	t.Run("get with etag", func(t *testing.T) {
		groupHandle(func() {
			req := newRequest("GET", "/get-with-etag")
			resp := httptest.NewRecorder()
			d.ServeHTTP(resp, req)
			h := resp.Header()
			assert.Equal(resp.Code, 200)
			assert.Equal(h.Get(cod.HeaderETag), `"123"`)
		})
	})

	t.Run("get no cache", func(t *testing.T) {
		statusList := make([]string, 0)
		respList := make([]string, 0)
		var mu sync.Mutex
		groupHandle(func() {
			req := newRequest("GET", "/no-cache")
			resp := httptest.NewRecorder()
			d.ServeHTTP(resp, req)
			h := resp.Header()
			status := h.Get(df.HeaderStatus)
			mu.Lock()
			defer mu.Unlock()
			statusList = append(statusList, status)
			respList = append(respList, resp.Body.String())
			assert.Equal(resp.Code, 200)
		})
		checkStatusList(t, statusList, fetchStatus, hitForPassStatus)
		iRespList := make([]int, 0)
		for _, value := range respList {
			v, err := strconv.Atoi(value)
			assert.Nil(err, "atoi fail")
			iRespList = append(iRespList, v)
		}
		sort.Sort(sort.IntSlice(iRespList))
		for index, value := range iRespList {
			assert.Equal(value, index+1)
		}
	})
}
