package util

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"encoding/base64"
	"expvar"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"
	"github.com/vicanso/pike/vars"
	"github.com/visionmedia/go-debug"
)

// Debug debug日志输出
var Debug = debug.Debug(string(vars.Name))

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
func GetCacheAge(header *fasthttp.ResponseHeader) uint32 {
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

// SupportGzip 判断是否支持gzip
func SupportGzip(ctx *fasthttp.RequestCtx) bool {
	return ctx.Request.Header.HasAcceptEncodingBytes(vars.Gzip)
}

// SupportBr 判断是否支持brotli压缩
func SupportBr(ctx *fasthttp.RequestCtx) bool {
	return ctx.Request.Header.HasAcceptEncodingBytes(vars.Br)
}

// GenRequestKey 生成请求的key: Method + host + request uri
func GenRequestKey(ctx *fasthttp.RequestCtx) []byte {
	uri := ctx.URI()
	return bytes.Join([][]byte{
		ctx.Method(),
		uri.Host(),
		uri.RequestURI(),
	}, []byte(""))
}

// Gzip 对数据压缩
func Gzip(buf []byte) ([]byte, error) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, err := w.Write(buf)
	if err != nil {
		return nil, err
	}
	w.Close()
	return b.Bytes(), nil
}

// Gunzip 解压数据
func Gunzip(buf []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return ioutil.ReadAll(r)
}

// GetClientIP 获取客户端IP
func GetClientIP(ctx *fasthttp.RequestCtx) string {
	xFor := ctx.Request.Header.PeekBytes(vars.XForwardedFor)
	if len(xFor) == 0 {
		return ctx.RemoteIP().String()
	}
	arr := bytes.Split(xFor, []byte(","))
	return string(arr[0])
}

// ShouldCompress 判断该响应数据是否应该压缩(针对文本类压缩)
func ShouldCompress(contentType []byte) bool {
	// 检测是否为文本
	reg, _ := regexp.Compile(`text|application/javascript|application/x-javascript|application/json`)
	return reg.Match(contentType)
}

// TrimHeader 将无用的头属性删除（如Date Connection等）
func TrimHeader(header []byte) []byte {
	arr := bytes.Split(header, vars.LineBreak)
	data := make([][]byte, 0, len(arr))
	ignoreList := []string{
		"date",
		"connection",
	}
	for _, item := range arr {
		index := bytes.IndexByte(item, vars.Colon)
		if index == -1 {
			continue
		}
		k := strings.ToLower(string(item[:index]))
		found := false
		for _, ignore := range ignoreList {
			if found {
				break
			}
			if k == ignore {
				found = true
			}
		}
		if found {
			continue
		}
		data = append(data, item)
	}
	return bytes.Join(data, vars.LineBreak)
}

// GetDebugVars 获取 debug vars
func GetDebugVars() []byte {
	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	fmt.Fprintf(w, "{\n")
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprintf(w, ",\n")
		}
		first = false
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(w, "\n}\n")
	w.Flush()
	return b.Bytes()
}

// GetETag 获取数据对应的ETag
func GetETag(buf []byte) string {
	size := len(buf)
	if size == 0 {
		return "\"0-2jmj7l5rSw0yVb/vlWAYkK/YBwk\""
	}
	h := sha1.New()
	h.Write(buf)
	hash := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("\"%x-%s\"", size, hash)
}
