package server

import (
	"bytes"
	"log"
	"strconv"
	"time"

	"../cache"
	"../director"
	"../dispatch"
	"../httplog"
	"../performance"
	"../proxy"
	"../util"
	"../vars"
	"github.com/valyala/fasthttp"
)

var hitForPassTTL uint32 = 300

// PikeConfig 程序配置
type PikeConfig struct {
	Name                 string
	Cpus                 int
	Listen               string
	DB                   string
	AdminPath            string `yaml:"adminPath"`
	AdminToken           string `yaml:"adminToken"`
	DisableKeepalive     bool   `yaml:"disableKeepalive"`
	Concurrency          int
	HitForPass           int           `yaml:"hitForPass"`
	ReadBufferSize       int           `yaml:"readBufferSize"`
	WriteBufferSize      int           `yaml:"writeBufferSize"`
	ConnectTimeout       time.Duration `yaml:"connectTimeout"`
	ReadTimeout          time.Duration `yaml:"readTimeout"`
	WriteTimeout         time.Duration `yaml:"writeTimeout"`
	MaxConnsPerIP        int           `yaml:"maxConnsPerIP"`
	MaxKeepaliveDuration time.Duration `yaml:"maxKeepaliveDuration"`
	MaxRequestBodySize   int           `yaml:"maxRequestBodySize"`
	ExpiredClearInterval time.Duration `yaml:"expiredClearInterval"`
	LogFormat            string        `yaml:"logFormat"`
	UDPLog               string        `yaml:"udpLog"`
	AccessLog            string        `yaml:"accessLog"`
	LogType              string        `yaml:"logType"`
	Directors            []*director.Config
}

// getDirector 获取director
func getDirector(host, uri []byte, directorList director.DirectorSlice) *director.Director {
	var found *director.Director
	// 查找可用的director
	for _, d := range directorList {
		if found == nil && d.Match(host, uri) {
			found = d
		}
	}
	return found
}

// 转发处理，返回响应头与响应数据
func doProxy(ctx *fasthttp.RequestCtx, us *proxy.Upstream, timeout time.Duration) (*fasthttp.Response, []byte, []byte, error) {
	resp, err := proxy.Do(ctx, us, timeout)
	if err != nil {
		return nil, nil, nil, err
	}
	body, err := dispatch.GetResponseBody(resp)
	if err != nil {
		return nil, nil, nil, err
	}
	header := dispatch.GetResponseHeader(resp)
	return resp, header, body, nil
}

// 设置响应的 Server-Timing
func setServerTiming(ctx *fasthttp.RequestCtx, startedAt time.Time) {
	v := startedAt.UnixNano()
	now := time.Now().UnixNano()
	use := int((now - v) / 1000000)
	desc := []byte("0=" + strconv.Itoa(use) + ";" + string(vars.Name))
	header := &ctx.Response.Header

	serverTiming := header.PeekBytes(vars.ServerTiming)
	if len(serverTiming) == 0 {
		header.SetCanonical(vars.ServerTiming, desc)
	} else {
		header.SetCanonical(vars.ServerTiming, bytes.Join([][]byte{
			desc,
			serverTiming,
		}, []byte(",")))
	}
}

func handler(ctx *fasthttp.RequestCtx, directorList director.DirectorSlice, conf *PikeConfig) {
	host := ctx.Request.Host()
	uri := ctx.RequestURI()
	found := getDirector(host, uri, directorList)
	// 出错处理
	errorHandler := func(err error) {
		dispatch.ErrorHandler(ctx, err)
	}
	// 正常的响应
	responseHandler := func(data *cache.ResponseData) {
		dispatch.Response(ctx, data)
	}
	if found == nil {
		// 没有可用的配置（）
		errorHandler(vars.ErrDirectorUnavailable)
		return
	}
	us := found.Upstream
	// 判断该请求是否直接pass到backend
	isPass := util.Pass(ctx, found.Passes)
	status := vars.Pass
	var key []byte
	// 如果不是pass的请求，则获取该请求对应的状态
	if !isPass {
		key = util.GenRequestKey(ctx)
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
		resp, header, body, err := doProxy(ctx, us, conf.ConnectTimeout)
		if err != nil {
			errorHandler(err)
			return
		}
		respData := &cache.ResponseData{
			CreatedAt:  util.GetSeconds(),
			StatusCode: uint16(resp.StatusCode()),
			Compress:   vars.RawData,
			TTL:        0,
			Header:     header,
			Body:       body,
		}
		responseHandler(respData)
	case vars.Fetching, vars.HitForPass:
		respHeadr.SetCanonical(vars.XCache, vars.XCacheMiss)
		//feacthing或hitforpass的请求转至upstream
		// 并根据返回的数据是否可以缓存设置缓存
		resp, header, body, err := doProxy(ctx, us, conf.ConnectTimeout)
		if err != nil {
			cache.HitForPass(key, hitForPassTTL)
			errorHandler(err)
			return
		}
		statusCode := uint16(resp.StatusCode())
		cacheAge := util.GetCacheAge(&resp.Header)
		compressType := vars.RawData
		contentType := resp.Header.PeekBytes(vars.ContentType)
		shouldCompress := util.ShouldCompress(contentType)
		// 可以缓存的数据，则将数据先压缩
		// 不可缓存的数据，`dispatch.Response`函数会根据客户端来决定是否压缩
		if shouldCompress && cacheAge > 0 && len(body) > vars.CompressMinLength {
			gzipData, err := util.Gzip(body)
			if err == nil {
				body = gzipData
				compressType = vars.GzipData
			}
		}
		respData := &cache.ResponseData{
			CreatedAt:  util.GetSeconds(),
			StatusCode: statusCode,
			Compress:   uint16(compressType),
			TTL:        cacheAge,
			Header:     header,
			Body:       body,
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

// Start 启动服务器
func Start(conf *PikeConfig, directorList director.DirectorSlice) error {
	listen := conf.Listen
	if len(listen) == 0 {
		listen = ":3015"
	}
	if conf.HitForPass > 0 {
		hitForPassTTL = uint32(conf.HitForPass)
	}

	var blockIP = &BlockIP{}
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
			clientIP := util.GetClientIP(ctx)
			if blockIP.FindIndex(clientIP) != -1 {
				dispatch.ErrorHandler(ctx, vars.ErrAccessIsNotAlloed)
				return
			}
			path := ctx.Path()
			// health check
			if bytes.Compare(path, vars.PingURL) == 0 {
				ctx.SetBody([]byte("pong"))
				return
			}
			// 管理界面相关接口
			if len(adminPath) != 0 && bytes.HasPrefix(path, adminPath) {
				adminHandler(ctx, directorList, blockIP, conf)
				return
			}
			performance.IncreaseRequestCount()
			performance.IncreaseConcurrency()
			defer performance.DecreaseConcurrency()
			startedAt := time.Now()

			defer setServerTiming(ctx, startedAt)
			if enableAccessLog {
				defer func() {
					logBuf := httplog.Format(ctx, tags, startedAt)
					go logWriter.Write(logBuf)
				}()
			}
			handler(ctx, directorList, conf)
		},
	}
	log.Printf("the server will listen on " + listen)
	return s.ListenAndServe(listen)
}
