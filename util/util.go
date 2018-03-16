package util

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"expvar"
	"fmt"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"net"
	"strconv"
	"time"

	"github.com/google/brotli/go/cbrotli"
	"github.com/valyala/fasthttp"
	"github.com/vicanso/pike/vars"
	"github.com/visionmedia/go-debug"
)

// Debug debug日志输出
var Debug = debug.Debug(string(vars.Name))

// Gzip 对数据压缩
func Gzip(buf []byte, level int) ([]byte, error) {
	var b bytes.Buffer
	if level <= 0 {
		level = gzip.DefaultCompression
	}
	w, _ := gzip.NewWriterLevel(&b, level)
	_, err := w.Write(buf)
	if err != nil {
		return nil, err
	}
	w.Close()
	return b.Bytes(), nil
}

// Brotli brotli压缩
func Brotli(buf []byte, quality int) ([]byte, error) {
	if quality == 0 {
		quality = 9
	}
	return cbrotli.Encode(buf, cbrotli.WriterOptions{
		Quality: quality,
		LGWin:   0,
	})
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

// CompressJPEG 压缩jpeg图片
func CompressJPEG(buf []byte, quality int) ([]byte, error) {
	if quality <= 0 {
		quality = 70
	}
	img, err := jpeg.Decode(bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}
	newBuf := bytes.NewBuffer(nil) //开辟一个新的空buff
	err = jpeg.Encode(newBuf, img, &jpeg.Options{
		Quality: quality,
	})
	return newBuf.Bytes(), err
}

// CompressPNG 压缩png图片
func CompressPNG(buf []byte) ([]byte, error) {
	img, err := png.Decode(bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}
	newBuf := bytes.NewBuffer(nil) //开辟一个新的空buff
	err = png.Encode(newBuf, img)
	if err != nil {
		return nil, err
	}
	return newBuf.Bytes(), nil
}

// GetClientIP 获取客户端IP
func GetClientIP(ctx *fasthttp.RequestCtx) string {
	xFor := ctx.Request.Header.PeekBytes(vars.XForwardedFor)
	ip := ctx.RemoteIP()
	if len(xFor) != 0 {
		arr := bytes.Split(xFor, []byte(","))
		address := net.ParseIP(string(arr[0]))
		xIP := address.To4()
		if xIP == nil {
			xIP = address.To16()
		}
		if xIP != nil {
			ip = xIP
		}
	}
	return ip.String()
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
