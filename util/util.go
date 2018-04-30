package util

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/google/brotli/go/cbrotli"
)

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
