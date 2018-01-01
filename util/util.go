package util

import (
	"bytes"
	"encoding/binary"
	"regexp"
	"strconv"
	"time"

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
func GetCacheAge(ctx *fasthttp.RequestCtx) uint16 {
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
		return uint16(maxAge)
	}

	// 从max-age中获取缓存时间
	reg, _ = regexp.Compile(`max-age=(\d+)`)
	result = reg.FindSubmatch(cacheControl)
	if len(result) != 2 {
		return 0
	}
	maxAge, _ := strconv.Atoi(string(result[1]))
	return uint16(maxAge)
}

// SupportGzip 判断是否支持gzip
func SupportGzip(ctx *fasthttp.RequestCtx) bool {
	return ctx.Request.Header.HasAcceptEncodingBytes(vars.Gzip)
}

// SupportBr 判断是否支持brotli压缩
func SupportBr(ctx *fasthttp.RequestCtx) bool {
	return ctx.Request.Header.HasAcceptEncodingBytes(vars.Br)
}

// ConvertUint16ToBytes 将uint16转换为[]byte
func ConvertUint16ToBytes(v uint16) []byte {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, v)
	return buf
}

// ConvertBytesToUint16 将[]byte转换为uint16
func ConvertBytesToUint16(buf []byte) uint16 {
	return binary.LittleEndian.Uint16(buf)
}

// ConvertUint32ToBytes 将uint32转换为[]byte
func ConvertUint32ToBytes(v uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, v)
	return buf
}

// ConvertBytesToUint32 将[]byte转换为uint32
func ConvertBytesToUint32(buf []byte) uint32 {
	return binary.LittleEndian.Uint32(buf)
}

// GetSeconds 获取当前时间的时间戳（秒）
func GetSeconds() int64 {
	return time.Now().Unix()
}

// GetNowSecondsBytes 获取当时时间的字节表示(4个字节)
func GetNowSecondsBytes() []byte {
	return ConvertUint32ToBytes(uint32(GetSeconds()))
}

// ConvertToSeconds 将字节保存的秒转换为整数
func ConvertToSeconds(buf []byte) uint32 {
	return binary.LittleEndian.Uint32(buf)
}

// GenRequestKey 生成请求的key: Method + host + request uri
func GenRequestKey(ctx *fasthttp.RequestCtx) []byte {
	return bytes.Join([][]byte{
		ctx.Method(),
		ctx.URI().FullURI(),
	}, []byte(""))
}
