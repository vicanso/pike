package cache

import (
	"bytes"
	"crypto/sha1"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/vicanso/hes"

	"github.com/vicanso/cod"
	"github.com/vicanso/pike/df"
	"github.com/vicanso/pike/log"
	"github.com/vicanso/pike/util"

	"go.uber.org/zap"
)

var (
	ignoreHeaderKeys = []string{
		df.HeaderAge,
		cod.HeaderContentEncoding,
		cod.HeaderContentLength,
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
		list []*Cache
		opts *Options
	}
	// Options dispatch options
	Options struct {
		// Size cache list's size
		Size int
		// ZoneSize cache zone's size
		ZoneSize int
		// CompressLevel compress level
		CompressLevel int
		// CompressMinLength compress min length
		CompressMinLength int
		// HitForPassTTL hit for pass ttl
		HitForPassTTL int
		// TextFilter text filter regexp
		TextFilter *regexp.Regexp
	}
	// Cache cache dispatcher
	Cache struct {
		// 用于保证每个dispatcher的操作，避免同时操作lru cache
		mu       sync.Mutex
		lruCache *LRUCache
	}
	// HTTPCache http cache
	HTTPCache struct {
		// 是否写锁(仅首次创建使用)
		writeLock bool
		opts      *Options
		rw        sync.RWMutex
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
	// HTTPCacheInfo http cache info
	HTTPCacheInfo struct {
		Key       string `json:"key,omitempty"`
		MaxAge    int    `json:"maxAge,omitempty"`
		ExpiredAt int64  `json:"expiredAt,omitempty"`
		Status    int    `json:"status,omitempty"`
	}
)

// GetStatusDesc get status desc
func GetStatusDesc(status int) string {
	return statusDescMap[status]
}

// byteSliceToString converts a []byte to string without a heap allocation.
func byteSliceToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// NewDispatcher new dispatcher
func NewDispatcher(opts Options) *Dispatcher {
	size := opts.Size
	if size <= 0 {
		size = 10
	}
	zoneSize := opts.ZoneSize
	if zoneSize <= 0 {
		zoneSize = 1024
	}
	list := make([]*Cache, size)

	for i := 0; i < size; i++ {
		list[i] = &Cache{
			lruCache: NewLRU(zoneSize),
		}
	}
	return &Dispatcher{
		list: list,
		opts: &opts,
	}
}

// GetCacheList 获取缓存数据列表
func (dsp *Dispatcher) GetCacheList() (cacheList []*HTTPCacheInfo) {
	cacheList = make([]*HTTPCacheInfo, 0, 100)
	now := time.Now().Unix()
	for _, cache := range dsp.list {
		cache.mu.Lock()
		cache.lruCache.ForEach(func(key string, value *HTTPCache) {
			if value.ExpiredAt < now {
				return
			}
			cacheList = append(cacheList, &HTTPCacheInfo{
				Key:       key,
				MaxAge:    value.MaxAge,
				ExpiredAt: value.ExpiredAt,
				Status:    value.Status,
			})
		})
		cache.mu.Unlock()
	}
	return
}

func (dsp *Dispatcher) getCache(k []byte) *Cache {
	b := sha1.Sum(k)
	index := (int(b[0]) | int(b[1])<<8) % len(dsp.list)
	return dsp.list[index]
}

// Expire expire the cache
func (dsp *Dispatcher) Expire(k []byte) {
	cache := dsp.getCache(k)

	key := byteSliceToString(k)
	lruCache := cache.lruCache
	// 保证lru cache的并发安全
	// 此锁需要快速释放，不能长期占用
	cache.mu.Lock()
	if v, ok := lruCache.Get(key); ok {
		if v != nil {
			// 设置为过期
			expiredAt := atomic.LoadInt64(&v.ExpiredAt)
			// 0 为fetching，不设置过期
			if expiredAt != 0 {
				atomic.StoreInt64(&v.ExpiredAt, 1)
			}
		}
	}
	cache.mu.Unlock()
}

// GetHTTPCache get http cache
func (dsp *Dispatcher) GetHTTPCache(k []byte) (hc *HTTPCache) {
	cache := dsp.getCache(k)

	key := byteSliceToString(k)
	lruCache := cache.lruCache
	// 保证lru cache的并发安全
	// 此锁需要快速释放，不能长期占用
	cache.mu.Lock()
	if v, ok := lruCache.Get(key); ok {
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
		// 解除cache的锁，此时其它不同类型的key则可获取锁
		// 此条件下cache的解锁必须前置，不然有可能因为此key的读锁获取不到而导致其它的key也无法操作
		cache.mu.Unlock()
		// 后续使用到读数据，调用读锁，如果此缓存有写锁，则需等待
		hc.rw.RLock()
		return
	}
	hc = &HTTPCache{
		writeLock: true,
		Status:    Fetch,
		opts:      dsp.opts,
	}
	// 首次创建的http cache，需要写数据，因此调用写锁
	hc.rw.Lock()

	lruCache.Add(key, hc)
	// hc锁成功之后，再解除cache的锁（此顺序不可调换）
	cache.mu.Unlock()
	return
}

// Done use http cache done
func (hc *HTTPCache) Done() {
	// 如果是首次创建使用的写锁，则设置为false并解锁
	if hc.writeLock {
		hc.writeLock = false
		hc.rw.Unlock()
	} else {
		hc.rw.RUnlock()
	}
}

// HitForPass set status to be hit for pass
func (hc *HTTPCache) HitForPass() {
	// 调用hit for pass函数前应该先获取写锁
	hc.Status = HitForPass
	hc.CreatedAt = time.Now().Unix()
	ttl := 0
	if hc.opts != nil {
		ttl = hc.opts.HitForPassTTL
	}
	hc.MaxAge = ttl
	atomic.StoreInt64(&hc.ExpiredAt, hc.CreatedAt+int64(ttl))
}

// Cacheable set status to bo cacheable
func (hc *HTTPCache) Cacheable(maxAge int, c *cod.Context) {
	hc.Status = Cacheable
	hc.CreatedAt = time.Now().Unix()
	hc.MaxAge = maxAge
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
	var err error
	var textFilter *regexp.Regexp
	opts := hc.opts
	compressMinLength := 0
	if opts != nil {
		textFilter = opts.TextFilter
		compressMinLength = opts.CompressMinLength
	}
	if opts != nil &&
		len(body) >= compressMinLength &&
		textFilter != nil &&
		textFilter.MatchString(header.Get(cod.HeaderContentType)) {
		gzipBody, err = doGzip(body, opts.CompressLevel)
		if err != nil {
			log.Default().Error("gzip fail",
				zap.String("url", c.Request.RequestURI),
				zap.Error(err),
			)
		}
		brBody, err = doBrotli(body, opts.CompressLevel)
		if err != nil {
			log.Default().Error("brotli fail",
				zap.String("url", c.Request.RequestURI),
				zap.Error(err),
			)
		}
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

// Response get the response data
func (hc *HTTPCache) Response(acceptEncoding string) (buf *bytes.Buffer, encoding string, err error) {
	// 如果有br压缩数据，而且客户端接受br
	if hc.BrBody != nil &&
		strings.Contains(acceptEncoding, df.BR) {
		buf = hc.BrBody
		encoding = df.BR
	} else if hc.GzipBody != nil && strings.Contains(acceptEncoding, df.GZIP) {
		// 如果有gzip压缩数据，而且客户端接受gzip
		buf = hc.GzipBody
		encoding = df.GZIP
	} else if hc.GzipBody != nil {
		// 缓存了压缩数据，但是客户端不支持，需要解压
		// 因为如果数据可压缩，缓存中只缓存压缩数据
		rawData, e := Gunzip(hc.GzipBody.Bytes())
		if e != nil {
			err = hes.NewWithError(e)
			return
		}
		buf = bytes.NewBuffer(rawData)
	} else {
		buf = hc.Body
	}
	return
}
