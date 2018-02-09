package util

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"expvar"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/vicanso/pike/vars"
	"github.com/visionmedia/go-debug"
)

// Debug debug日志输出
var Debug = debug.Debug(string(vars.Name))

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

// GetTimeConsuming 获取使用耗时(ms)
func GetTimeConsuming(startedAt time.Time) int {
	v := startedAt.UnixNano()
	now := time.Now().UnixNano()
	return int((now - v) / 1000000)
}

// SetTimingConsumingHeader 设置耗时至http头
func SetTimingConsumingHeader(startedAt time.Time, header *fasthttp.RequestHeader, key []byte) {
	ms := GetTimeConsuming(startedAt)
	header.SetCanonical(key, []byte(strconv.Itoa(ms)))
}
