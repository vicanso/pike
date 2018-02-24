package server

import (
	"bytes"
	"encoding/base64"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/director"
	"github.com/vicanso/pike/dispatch"
	"github.com/vicanso/pike/httplog"
	"github.com/vicanso/pike/performance"
	"github.com/vicanso/pike/proxy"
	"github.com/vicanso/pike/util"
	"github.com/vicanso/pike/vars"
)

var hitForPassTTL uint32 = 300
var defaultListen = ":3015"

var defaultTextTypes = [][]byte{
	[]byte("text"),
	[]byte("javascript"),
	[]byte("json"),
}
var compresssTypes = make([][]byte, 0)

// Config the server config
type Config struct {
	Name             string
	Concurrency      int
	DisableKeepalive bool
	ReadBufferSize   int
	WriteBufferSize  int
	ETag             bool
	// 最小压缩字节长度
	CompressMinLength int
	// 压缩级别
	CompressLevel int
	// 设置jpeg的质量
	JpegQuality int
	// 各类超时配置
	ConnectTimeout       time.Duration
	ReadTimeout          time.Duration
	WriteTimeout         time.Duration
	MaxKeepaliveDuration time.Duration

	MaxConnsPerIP      int
	MaxRequestBodySize int
	Listen             string
	HitForPass         int
	AdminPath          string
	AdminToken         string
	TextTypes          []string
	ResponseHeader     []string
	EnableServerTiming bool
	Favicon            string
	// HTTP日志相关
	LogFormat string
	LogType   string
	UDPLog    string
	AccessLog string
	// HTTPS证书相关
	CertFile string
	KeyFile  string
}

// Header HTTP response header
type Header struct {
	Key   []byte
	Value []byte
}

// 需要清除的headers
var ignoreHeaders = [][]byte{
	[]byte("Date:"),
	[]byte("Connection:"),
	[]byte("Server:"),
}

// trimHeader 将无用的头属性删除（如Date Connection等）
func trimHeader(header []byte) []byte {
	arr := bytes.Split(header, vars.LineBreak)
	data := make([][]byte, 0, len(arr))
	for _, item := range arr {
		index := bytes.IndexByte(item, vars.Colon)
		if index == -1 {
			continue
		}
		found := false
		for _, ignore := range ignoreHeaders {
			if found {
				break
			}
			if bytes.Index(item, ignore) == 0 {
				found = true
			}
		}
		// 需要忽略的http头
		if found {
			continue
		}
		data = append(data, item)
	}
	return bytes.Join(data, vars.LineBreak)
}

// getResponseHeader 获取响应的header
func getResponseHeader(resp *fasthttp.Response) []byte {
	return trimHeader(resp.Header.Header())
}

// getResponseBody 获取响应的数据
func getResponseBody(resp *fasthttp.Response) ([]byte, error) {
	enconding := resp.Header.PeekBytes(vars.ContentEncoding)
	if bytes.Compare(enconding, vars.Deflate) == 0 {
		return resp.BodyInflate()
	}
	if bytes.Compare(enconding, vars.Gzip) == 0 {
		return resp.BodyGunzip()
	}
	return resp.Body(), nil
}

// 转发处理，返回响应头与响应数据
func doProxy(ctx *fasthttp.RequestCtx, us *proxy.Upstream, conf *Config, proxyConfig *proxy.Config) (*fasthttp.Response, []byte, []byte, error) {
	resp, err := proxy.Do(ctx, us, proxyConfig)
	if err != nil {
		return nil, nil, nil, err
	}
	body, err := getResponseBody(resp)
	quality := conf.JpegQuality

	contentType := resp.Header.PeekBytes(vars.ContentType)

	if quality > 0 && bytes.Equal(contentType, vars.JPEG) {
		buf, _ := util.CompressJPEG(body, quality)
		if len(buf) > 0 {
			body = buf
		}
	}

	if err != nil {
		return nil, nil, nil, err
	}
	header := getResponseHeader(resp)
	return resp, header, body, nil
}

func getTimingDesc(ms, name string) []byte {
	if len(ms) == 0 {
		return nil
	}
	return []byte("0=" + ms + ";" + name)
}

// 设置响应的 Server-Timing
func setServerTiming(ctx *fasthttp.RequestCtx, startedAt time.Time) {
	header := &ctx.Response.Header
	reqHeader := &ctx.Request.Header
	ms := util.GetTimeConsuming(startedAt)
	totalDesc := getTimingDesc(strconv.Itoa(ms), string(vars.Name))
	fetchDesc := getTimingDesc(string(reqHeader.PeekBytes(vars.TimingFetch)), string(vars.Name)+"-fetch")
	gzipDesc := getTimingDesc(string(reqHeader.PeekBytes(vars.TimingGzip)), string(vars.Name)+"-gzip")

	timing := [][]byte{
		totalDesc,
	}
	if len(fetchDesc) != 0 {
		timing = append(timing, fetchDesc)
	}
	if len(gzipDesc) != 0 {
		timing = append(timing, gzipDesc)
	}

	serverTiming := header.PeekBytes(vars.ServerTiming)
	if len(serverTiming) != 0 {
		timing = append(timing, serverTiming)
	}
	header.SetCanonical(vars.ServerTiming, bytes.Join(timing, []byte(",")))
}

// 增加额外的 Response Header
func addExtraHeader(ctx *fasthttp.RequestCtx, headers []*Header) {
	for _, header := range headers {
		ctx.Response.Header.SetCanonical(header.Key, header.Value)
	}
}

// isPass 判断该请求是否直接pass（不可缓存）
func isPass(ctx *fasthttp.RequestCtx, passList [][]byte) bool {
	method := ctx.Method()
	if bytes.Compare(method, vars.Get) != 0 && bytes.Compare(method, vars.Head) != 0 {
		return true
	}
	pass := false
	uri := ctx.RequestURI()
	for _, item := range passList {
		if bytes.Contains(uri, item) {
			pass = true
			break
		}
	}
	return pass
}

// 根据Cache-Control的信息，获取s-maxage或者max-age的值
func getCacheAge(header *fasthttp.ResponseHeader) uint32 {
	cacheControl := header.PeekBytes(vars.CacheControl)
	if len(cacheControl) == 0 {
		return 0
	}
	// 如果设置不可缓存，返回0
	reg, _ := regexp.Compile(`no-cache|no-store|private`)
	match := reg.Match(cacheControl)
	if match {
		return 0
	}

	// 优先从s-maxage中获取
	reg, _ = regexp.Compile(`s-maxage=(\d+)`)
	result := reg.FindSubmatch(cacheControl)
	if len(result) == 2 {
		maxAge, _ := strconv.Atoi(string(result[1]))
		return uint32(maxAge)
	}

	// 从max-age中获取缓存时间
	reg, _ = regexp.Compile(`max-age=(\d+)`)
	result = reg.FindSubmatch(cacheControl)
	if len(result) != 2 {
		return 0
	}
	maxAge, _ := strconv.Atoi(string(result[1]))
	return uint32(maxAge)
}

// genRequestKey 生成请求的key: Method + host + request uri
func genRequestKey(ctx *fasthttp.RequestCtx) []byte {
	uri := ctx.URI()
	// 对于http https，只是与客户端的数据做加密，缓存的数据一致
	return bytes.Join([][]byte{
		ctx.Method(),
		uri.Host(),
		uri.RequestURI(),
	}, []byte(" "))
}

// shouldCompress 判断该响应数据是否应该压缩(针对文本类压缩)
func shouldCompress(header *fasthttp.ResponseHeader) bool {
	types := compresssTypes
	if len(compresssTypes) == 0 {
		types = defaultTextTypes
	}
	contentType := header.PeekBytes(vars.ContentType)
	found := false
	for _, v := range types {
		if found {
			break
		}
		if bytes.Contains(contentType, v) {
			found = true
		}
	}
	return found
}

func handler(ctx *fasthttp.RequestCtx, conf *Config, proxyConfig *proxy.Config) {
	host := ctx.Request.Host()
	uri := ctx.RequestURI()
	found := director.GetMatch(host, uri)
	compressLevel := conf.CompressLevel
	compressMinLength := conf.CompressMinLength
	if compressMinLength == 0 {
		compressMinLength = vars.CompressMinLength
	}

	// 出错处理
	errorHandler := func(err error) {
		dispatch.ErrorHandler(ctx, err)
	}
	// 正常的响应
	responseHandler := func(data *cache.ResponseData) {
		dispatch.Response(ctx, data, compressMinLength, compressLevel)
	}
	if found == nil {
		// 没有可用的配置（）
		errorHandler(vars.ErrDirectorUnavailable)
		return
	}

	us := proxy.GetUpStream(found.Name)
	// 判断该请求是否直接pass到backend
	pass := isPass(ctx, found.Passes)
	status := vars.Pass
	var key []byte
	// 如果不是pass的请求，则获取该请求对应的状态
	if !pass {
		key = genRequestKey(ctx)
		// 如果已经有相同的key在处理，则会返回c(chan int)
		s, c := cache.GetRequestStatus(key)
		status = s
		// 如果有chan，等待chan返回的状态
		if c != nil {
			status = <-c
		}
	}
	respHeadr := &ctx.Response.Header
	switch status {
	case vars.Pass:
		respHeadr.SetCanonical(vars.XCache, vars.XCacheMiss)
		// pass的请求直接转发至upstream
		resp, header, body, err := doProxy(ctx, us, conf, proxyConfig)
		if err != nil {
			errorHandler(err)
			return
		}

		respData := &cache.ResponseData{
			CreatedAt:      uint32(time.Now().Unix()),
			StatusCode:     uint16(resp.StatusCode()),
			Compress:       vars.RawData,
			ShouldCompress: shouldCompress(&resp.Header),
			TTL:            0,
			Header:         header,
			Body:           body,
		}
		responseHandler(respData)
	case vars.Fetching, vars.HitForPass:
		respHeadr.SetCanonical(vars.XCache, vars.XCacheMiss)
		//feacthing或hitforpass的请求转至upstream
		// 并根据返回的数据是否可以缓存设置缓存
		resp, header, body, err := doProxy(ctx, us, conf, proxyConfig)
		if err != nil {
			cache.HitForPass(key, hitForPassTTL)
			errorHandler(err)
			return
		}
		statusCode := uint16(resp.StatusCode())
		cacheAge := getCacheAge(&resp.Header)
		compressType := vars.RawData
		// 可以缓存的数据，则将数据先压缩
		// 不可缓存的数据，`dispatch.Response`函数会根据客户端来决定是否压缩
		shouldDoCompress := shouldCompress(&resp.Header)
		if shouldDoCompress && cacheAge > 0 && len(body) > compressMinLength {
			gzipStartedAt := time.Now()
			gzipData, err := util.Gzip(body, compressLevel)
			util.SetTimingConsumingHeader(gzipStartedAt, &ctx.Request.Header, vars.TimingGzip)
			if err == nil {
				body = gzipData
				compressType = vars.GzipData
			}
		}
		respData := &cache.ResponseData{
			CreatedAt:      uint32(time.Now().Unix()),
			StatusCode:     statusCode,
			Compress:       uint8(compressType),
			ShouldCompress: shouldDoCompress,
			TTL:            cacheAge,
			Header:         header,
			Body:           body,
		}
		responseHandler(respData)

		if cacheAge <= 0 {
			// 如果原来的状态不是hitForPass，则设置状态
			if status != vars.HitForPass {
				cache.HitForPass(key, hitForPassTTL)
			}
		} else {
			err = cache.SaveResponseData(key, respData)
			if err != nil {
				// 如果保存数据失败，则设置hit for pass
				cache.HitForPass(key, hitForPassTTL)
			} else {
				// 如果保存数据成功，则设置为cacheable
				cache.Cacheable(key, cacheAge)
			}
		}
	case vars.Cacheable:
		respHeadr.SetCanonical(vars.XCache, vars.XCacheHit)
		respData, err := cache.GetResponse(key)
		if err != nil {
			errorHandler(err)
			return
		}
		responseHandler(respData)
	}
}

func pingHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetBody([]byte("pong"))
}

func faviconHandler(ctx *fasthttp.RequestCtx, conf *Config) {
	buf, _ := base64.StdEncoding.DecodeString(conf.Favicon)
	ctx.SetContentType("image/x-icon")
	ctx.SetBody(buf)
}

// Start 启动服务
func Start(conf *Config) error {
	listen := conf.Listen
	if len(listen) == 0 {
		listen = defaultListen
	}
	if conf.HitForPass > 0 {
		hitForPassTTL = uint32(conf.HitForPass)
	}

	var blockIP = &BlockIP{
		m: &sync.RWMutex{},
	}
	blockIP.InitFromCache()
	tags := httplog.Parse([]byte(conf.LogFormat))
	writeCategory := httplog.Normal
	if conf.LogType == "date" {
		writeCategory = httplog.Date
	}
	var logWriter httplog.Writer
	if len(conf.UDPLog) != 0 {
		logWriter = &httplog.UDPWriter{
			URI: conf.UDPLog,
		}
	} else if len(conf.AccessLog) != 0 {
		logWriter = &httplog.FileWriter{
			Path:     conf.AccessLog,
			Category: writeCategory,
		}
	}
	if logWriter != nil {
		defer logWriter.Close()
	}
	enableAccessLog := len(tags) != 0 && logWriter != nil
	adminPath := []byte(conf.AdminPath)
	for _, str := range conf.TextTypes {
		compresssTypes = append(compresssTypes, []byte(str))
	}
	extraHeaders := make([]*Header, 0)
	if len(conf.ResponseHeader) != 0 {
		for _, str := range conf.ResponseHeader {
			arr := strings.Split(str, ":")
			if len(arr) != 2 {
				continue
			}
			h := &Header{
				Key:   []byte(arr[0]),
				Value: []byte(arr[1]),
			}
			extraHeaders = append(extraHeaders, h)
		}
	}
	proxyConfig := &proxy.Config{
		Timeout: conf.ConnectTimeout,
		ETag:    conf.ETag,
	}
	s := &fasthttp.Server{
		Name:                 conf.Name,
		Concurrency:          conf.Concurrency,
		DisableKeepalive:     conf.DisableKeepalive,
		ReadBufferSize:       conf.ReadBufferSize,
		WriteBufferSize:      conf.WriteBufferSize,
		ReadTimeout:          conf.ReadTimeout,
		WriteTimeout:         conf.WriteTimeout,
		MaxConnsPerIP:        conf.MaxConnsPerIP,
		MaxKeepaliveDuration: conf.MaxKeepaliveDuration,
		MaxRequestBodySize:   conf.MaxRequestBodySize,
		Handler: func(ctx *fasthttp.RequestCtx) {
			if len(extraHeaders) != 0 {
				defer addExtraHeader(ctx, extraHeaders)
			}
			clientIP := util.GetClientIP(ctx)
			if blockIP.FindIndex(clientIP) != -1 {
				dispatch.ErrorHandler(ctx, vars.ErrAccessIsNotAlloed)
				return
			}
			path := ctx.Path()
			// health check
			if bytes.Compare(path, vars.PingURL) == 0 {
				pingHandler(ctx)
				return
			}
			// favicon
			if bytes.Compare(path, vars.FaviconURL) == 0 {
				faviconHandler(ctx, conf)
				return
			}
			// 管理界面相关接口
			if len(adminPath) != 0 && bytes.HasPrefix(path, adminPath) {
				adminHandler(ctx, conf, blockIP)
				return
			}
			performance.IncreaseRequestCount()
			performance.IncreaseConcurrency()
			defer performance.DecreaseConcurrency()
			startedAt := time.Now()
			if conf.EnableServerTiming {
				defer setServerTiming(ctx, startedAt)
			}
			if enableAccessLog {
				defer func() {
					logBuf := httplog.Format(ctx, tags, startedAt)
					go logWriter.Write(logBuf)
				}()
			}
			handler(ctx, conf, proxyConfig)
			ctx.Response.Header.SetServer(conf.Name)
		},
	}
	log.Printf("the server will listen on " + listen)
	if len(conf.CertFile) != 0 && len(conf.KeyFile) != 0 {
		return s.ListenAndServeTLS(listen, conf.CertFile, conf.KeyFile)
	}
	return s.ListenAndServe(listen)
}
