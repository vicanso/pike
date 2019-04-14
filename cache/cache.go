package cache

import (
	"bytes"
	"crypto/sha1"
	"net/http"
	"regexp"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/golang/groupcache/lru"
	"github.com/vicanso/cod"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/df"
	"github.com/vicanso/pike/util"
)

var (
	dispatcherList    []*Dispatcher
	dispatcherListLen int
	hitForPassTTL     int64
	compressLevel     int
	compressMinLength int
	textFilter        *regexp.Regexp

	ignoreHeaderKeys = []string{
		df.HeaderAge,
		cod.HeaderContentEncoding,
		df.HeaderContentLength,
	}
)

const (
	// Fetch fetch
	Fetch = iota
	// Pass request status: pass
	Pass
	// HitForPass request status: hitForPass
	HitForPass
	// Cacheable request status: cacheable
	Cacheable
)

var (
	statusDescMap = map[int]string{
		Fetch:      "fetch",
		Pass:       "pass",
		HitForPass: "hitForPass",
		Cacheable:  "cacheable",
	}
)

type (
	// Dispatcher dispatcher
	Dispatcher struct {
		// 用于保证每个dispatcher的操作，避免同时操作lru cache
		mu       sync.Mutex
		lruCache *lru.Cache
	}
	// HTTPCache http cache
	HTTPCache struct {
		rw sync.RWMutex
		// wg sync.WaitGroup
		// Status cache's status
		Status int
		// CreatedAt create time
		CreatedAt int64
		// MaxAge max-age
		MaxAge int
		// ExpiredAt expired time
		ExpiredAt int64
		// Headers http response's header
		Headers http.Header
		// StatusCode http status code
		StatusCode int
		// Body http response's body
		Body *bytes.Buffer
		// GzipBody http gzip response's body
		GzipBody *bytes.Buffer
		// BrBody http br response's body
		BrBody *bytes.Buffer
	}
)

func init() {
	compressLevel = config.GetCompressLevel()
	compressMinLength = config.GetCompressMinLength()

	textFilter = config.GetTextFilter()

	hitForPassTTL = int64(config.GetHitForPassTTL())
	size := config.GetCacheSize()
	dispatcherList = make([]*Dispatcher, size)
	zoneSize := config.GetCacheZoneSize()
	for i := 0; i < size; i++ {
		dispatcherList[i] = NewDispatcher(zoneSize)
	}
	dispatcherListLen = len(dispatcherList)
}

// GetHTTPCache get http cache
func GetHTTPCache(key []byte) (hc *HTTPCache) {
	b := sha1.Sum(key)
	index := (int(b[0]) | int(b[1])<<8) % dispatcherListLen
	dsp := dispatcherList[index]
	return dsp.GetHTTPCache(key)
}

// GetStatusDesc get status desc
func GetStatusDesc(status int) string {
	return statusDescMap[status]
}

// byteSliceToString converts a []byte to string without a heap allocation.
func byteSliceToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// NewDispatcher new dispatcher
func NewDispatcher(size int) *Dispatcher {
	return &Dispatcher{
		lruCache: lru.New(size),
	}
}

// GetHTTPCache get http cache
func (dsp *Dispatcher) GetHTTPCache(k []byte) (hc *HTTPCache) {
	key := byteSliceToString(k)
	lruCache := dsp.lruCache
	// 保证lru cache的并发安全
	// 此锁需要快速释放，不能长期占用
	dsp.mu.Lock()
	if c, ok := lruCache.Get(key); ok {
		v, _ := c.(*HTTPCache)
		if v != nil {
			// 如果未过期或刚初始化，则使用此缓存
			expiredAt := atomic.LoadInt64(&v.ExpiredAt)
			if expiredAt == 0 || expiredAt >= time.Now().Unix() {
				hc = v
			}
		}
	}
	// 如果获取到对应缓存，则直接返回
	if hc != nil {
		dsp.mu.Unlock()
		// 后续使用到读数据，调用读锁
		hc.rw.RLock()
		return
	}
	hc = &HTTPCache{
		Status: Fetch,
	}
	// 首次创建的http cache，需要写数据，因此调用写锁
	hc.rw.Lock()

	lruCache.Add(key, hc)
	dsp.mu.Unlock()
	return
}

// Unlock unlock
func (hc *HTTPCache) Unlock() {
	hc.rw.Unlock()
}

// RUnlock r unlock
func (hc *HTTPCache) RUnlock() {
	hc.rw.RUnlock()
}

// HitForPass set status to be hit for pass
func (hc *HTTPCache) HitForPass() {
	// 调用hit for pass函数前应该先获取写锁
	hc.Status = HitForPass
	hc.CreatedAt = time.Now().Unix()
	atomic.StoreInt64(&hc.ExpiredAt, hc.CreatedAt+hitForPassTTL)
}

// Cacheable set status to bo cacheable
func (hc *HTTPCache) Cacheable(maxAge int, c *cod.Context) {
	hc.Status = Cacheable
	hc.CreatedAt = time.Now().Unix()
	atomic.StoreInt64(&hc.ExpiredAt, hc.CreatedAt+int64(maxAge))

	header := c.Header()
	body := c.BodyBuffer.Bytes()
	encoding := header.Get(cod.HeaderContentEncoding)
	if encoding != "" {
		// 如果不是gzip，设置为hit for pass
		if encoding != df.GZIP {
			hc.HitForPass()
			return
		}
		if len(body) != 0 {
			buf, err := doGunzip(body)
			// 如果解压出错，设置为hit for pass
			if err != nil {
				hc.HitForPass()
				return
			}
			body = buf
		}
	}
	// 只针对文本并且大于等于最小尺寸数据压缩
	var gzipBody, brBody []byte
	if len(body) >= compressMinLength &&
		textFilter.MatchString(header.Get(cod.HeaderContentType)) {
		gzipBody, _ = doGzip(body)
		brBody, _ = doBrotli(body)
	}

	h := make(http.Header)
	for key, values := range header {
		if util.ContainString(ignoreHeaderKeys, key) {
			continue
		}
		for _, value := range values {
			h.Add(key, value)
		}
	}

	hc.StatusCode = c.StatusCode

	hc.Headers = h
	// 如果有gzip缓存，则不再缓存原数据
	// 因为绝大部分客户端都支持gzip
	// 不支持的客户端从gzip解压
	if len(gzipBody) == 0 {
		hc.Body = bytes.NewBuffer(body)
	}
	if len(gzipBody) != 0 {
		hc.GzipBody = bytes.NewBuffer(gzipBody)
	}
	if len(brBody) != 0 {
		hc.BrBody = bytes.NewBuffer(brBody)
	}
}
