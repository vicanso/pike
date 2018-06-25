package util

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/brotli/go/cbrotli"
	"github.com/vicanso/pike/vars"
)

const (
	kbytes = 1024
	mbytes = 1024 * 1024
)

func noop() {}

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

// BrotliEncode brotli压缩
func BrotliEncode(buf []byte, quality int) ([]byte, error) {
	if quality == 0 {
		quality = 9
	}
	return cbrotli.Encode(buf, cbrotli.WriterOptions{
		Quality: quality,
		LGWin:   0,
	})
}

// BrotliDecode brotli解压
func BrotliDecode(buf []byte) ([]byte, error) {
	return cbrotli.Decode(buf)
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

// GetHeaderValue 获取 http header的值
func GetHeaderValue(header http.Header, name string) (value []string) {
	n := strings.ToLower(name)
	for k, v := range header {
		if strings.ToLower(k) == n {
			value = v
			return
		}
	}
	return
}

// GetTimeConsuming 获取使用耗时(ms)
func GetTimeConsuming(startedAt time.Time) int {
	v := startedAt.UnixNano()
	now := time.Now().UnixNano()
	return int((now - v) / 1000000)
}

func cut(str string) string {
	l := len(str)
	if l == 0 {
		return str
	}
	ch := str[l-1]
	if ch == '0' || ch == '.' {
		return cut(str[0 : l-1])
	}
	return str
}

// GetHumanReadableSize 获取便于阅读的数据大小
func GetHumanReadableSize(size int64) string {
	if size < kbytes {
		return fmt.Sprintf("%dB", size)
	}
	fSize := float64(size)
	if size < mbytes {
		s := cut(fmt.Sprintf("%.2f", (fSize / kbytes)))
		return s + "KB"
	}
	s := cut(fmt.Sprintf("%.2f", (fSize / mbytes)))
	return s + "MB"
}

// GetRewriteRegexp 获取rewrite的正式匹配表
func GetRewriteRegexp(rewrites []string) map[*regexp.Regexp]string {
	rewriteRegexp := make(map[*regexp.Regexp]string)
	for _, value := range rewrites {
		arr := strings.Split(value, ":")
		if len(arr) != 2 {
			continue
		}
		k := arr[0]
		v := arr[1]
		k = strings.Replace(k, "*", "(\\S*)", -1)
		rewriteRegexp[regexp.MustCompile(k)] = v
	}
	return rewriteRegexp
}

// GetIdentity 获取该请求对应的标识
func GetIdentity(req *http.Request) []byte {
	methodLen := len(req.Method)
	hostLen := len(req.Host)
	uriLen := len(req.RequestURI)
	buffer := make([]byte, methodLen+hostLen+uriLen+2)
	len := 0

	copy(buffer[len:], req.Method)
	len += methodLen

	buffer[len] = vars.SpaceByte
	len++

	copy(buffer[len:], req.Host)
	len += hostLen

	buffer[len] = vars.SpaceByte
	len++

	copy(buffer[len:], req.RequestURI)
	return buffer
}
