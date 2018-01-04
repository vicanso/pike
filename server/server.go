package server

import (
	"time"

	"../cache"
	"../director"
	"../dispatch"
	"../proxy"
	"../util"
	"../vars"
	"github.com/valyala/fasthttp"
)

var hitForPassTTL uint32 = 300

// PikeConfig 程序配置
type PikeConfig struct {
	Name                 string
	Listen               string
	DB                   string
	DisableKeepalive     bool `yaml:"disableKeepalive"`
	Concurrency          int
	HitForPass           int           `yaml:"hitForPass"`
	ReadBufferSize       int           `yaml:"readBufferSize"`
	WriteBufferSize      int           `yaml:"writeBufferSize"`
	ReadTimeout          time.Duration `yaml:"readTimeout"`
	WriteTimeout         time.Duration `yaml:"writeTimeout"`
	MaxConnsPerIP        int           `yaml:"maxConnsPerIP"`
	MaxKeepaliveDuration time.Duration `yaml:"maxKeepaliveDuration"`
	MaxRequestBodySize   int           `yaml:"maxRequestBodySize"`
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

func doProxy(ctx *fasthttp.RequestCtx, us *proxy.Upstream) (*fasthttp.Response, []byte, []byte, error) {
	resp, err := proxy.Do(ctx, us)
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

func handler(ctx *fasthttp.RequestCtx, directorList director.DirectorSlice) {
	host := ctx.Request.Host()
	uri := ctx.RequestURI()
	found := getDirector(host, uri, directorList)
	if found == nil {
		// 没有可用的配置（）
		dispatch.ErrorHandler(ctx, vars.ErrDirectorUnavailable)
		return
	}
	us := found.Upstream
	// 判断该请求是否直接pass到backend
	isPass := util.Pass(ctx, found.Passes)
	status := vars.Pass
	var key []byte
	if !isPass {
		key = util.GenRequestKey(ctx)
		s, c := cache.GetRequestStatus(key)
		status = s
		if c != nil {
			status = <-c
		}
	}
	switch status {
	case vars.Pass:
		resp, header, body, err := doProxy(ctx, us)
		if err != nil {
			dispatch.ErrorHandler(ctx, err)
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
		dispatch.Response(ctx, respData)
	case vars.Fetching, vars.HitForPass:
		resp, header, body, err := doProxy(ctx, us)
		if err != nil {
			cache.HitForPass(key, hitForPassTTL)
			dispatch.ErrorHandler(ctx, err)
			return
		}
		statusCode := uint16(resp.StatusCode())
		cacheAge := util.GetCacheAge(&resp.Header)
		compressType := vars.RawData
		// 可以缓存的数据，则将数据先压缩
		// 不可缓存的数据，`dispatch.Response`函数会根据客户端来决定是否压缩
		if cacheAge > 0 {
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
		dispatch.Response(ctx, respData)

		if cacheAge <= 0 {
			cache.HitForPass(key, hitForPassTTL)
		} else {
			bucket := []byte(found.Name)
			err = cache.SaveResponseData(bucket, key, respData)
			if err != nil {
				cache.HitForPass(key, hitForPassTTL)
			} else {
				cache.Cacheable(key, cacheAge)
			}
		}
	case vars.Cacheable:
		bucket := []byte(found.Name)
		respData, err := cache.GetResponse(bucket, key)
		if err != nil {
			dispatch.ErrorHandler(ctx, err)
			return
		}
		dispatch.Response(ctx, respData)
	}
}

// Start 启动服务器
func Start(conf *PikeConfig, directorList director.DirectorSlice) *fasthttp.Server {
	listen := conf.Listen
	if len(listen) == 0 {
		listen = ":3015"
	}
	if conf.HitForPass > 0 {
		hitForPassTTL = uint32(conf.HitForPass)
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
			handler(ctx, directorList)
		},
	}
	s.ListenAndServe(listen)
	return s
}
