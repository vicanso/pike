package util

import (
	"bytes"
	"regexp"
	"strconv"

	"../vars"
	"github.com/valyala/fasthttp"
)

// Pass 判断该请求是否直接pass（不可缓存）
func Pass(ctx *fasthttp.RequestCtx, passList [][]byte) bool {
	method := ctx.Method()
	if bytes.Compare(method, vars.Get) != 0 && bytes.Compare(method, vars.Head) != 0 {
		return true
	}
	isPass := false
	uri := ctx.RequestURI()
	for _, item := range passList {
		if !isPass && bytes.Contains(uri, item) {
			isPass = true
		}
	}
	return isPass
}

// GetCacheAge 获取s-maxage或者max-age的值
func GetCacheAge(ctx *fasthttp.RequestCtx) int {
	cacheControl := ctx.Response.Header.PeekBytes(vars.CacheControl)
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
		return maxAge
	}

	// 从max-age中获取缓存时间
	reg, _ = regexp.Compile(`max-age=(\d+)`)
	result = reg.FindSubmatch(cacheControl)
	if len(result) != 2 {
		return 0
	}
	maxAge, _ := strconv.Atoi(string(result[1]))
	return maxAge
}

// SupportGzip 判断是否支持gzip
func SupportGzip(ctx *fasthttp.RequestCtx) bool {
	return ctx.Request.Header.HasAcceptEncodingBytes(vars.Gzip)
}

// SupportBr 判断是否支持brotli压缩
func SupportBr(ctx *fasthttp.RequestCtx) bool {
	return ctx.Request.Header.HasAcceptEncodingBytes(vars.Br)
}
