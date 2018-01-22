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
	"strings"

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
