package httplog

import (
	"bytes"
	"net"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/vicanso/pike/util"

	"github.com/valyala/fasthttp"
)

const (
	host           = "host"
	method         = "method"
	path           = "path"
	proto          = "proto"
	query          = "query"
	remote         = "remote"
	clientIP       = "client-ip"
	scheme         = "scheme"
	uri            = "uri"
	referer        = "referer"
	userAgent      = "userAgent"
	when           = "when"
	whenISO        = "when-iso"
	whenUnix       = "when-unix"
	whenISOMs      = "when-iso-ms"
	size           = "size"
	status         = "status"
	latency        = "latency"
	latencyMs      = "latency-ms"
	cookie         = "cookie"
	requestHeader  = "requestHeader"
	responseHeader = "responseHeader"
)

var (
	http11 = []byte("HTTP/1.1")
	http10 = []byte("HTTP/1.0")
	http   = []byte("HTTP")
	https  = []byte("HTTPS")
)

// Tag log tag
type Tag struct {
	category string
	data     []byte
}

const (
	// Normal 普通模式（所有日志写到同一个文件）
	Normal = iota
	// Date 日期分割（按天分割日志）
	Date
)

// Writer the writer interface
type Writer interface {
	Write(buf []byte) error
	Close() error
}

// FileWriter 以文件形式写日志
type FileWriter struct {
	Path     string
	Category int
	fd       *os.File
	m        sync.Mutex
	date     string
	file     string
}

// Write 写日志
func (w *FileWriter) Write(buf []byte) error {
	w.m.Lock()
	defer w.m.Unlock()
	// 如果是按日期分割
	if w.Category == Date {
		now := time.Now()
		date := now.Format("2006-01-02")
		// 如果日期有变化
		if w.date != date {
			w.date = date
			// 关闭当前的file
			if w.fd != nil {
				w.fd.Close()
			}
			w.fd = nil
			w.file = w.Path + "/" + w.date
		}
	} else {
		w.file = w.Path
	}
	// 如果文件未打开，打开文件
	if w.fd == nil {
		fd, err := os.OpenFile(w.file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			return err
		}
		w.fd = fd
	}
	_, err := w.fd.Write(append(buf, '\n'))
	return err
}

// Close 关闭写文件
func (w *FileWriter) Close() error {
	if w.fd != nil {
		return w.fd.Close()
	}
	return nil
}

// UDPWriter 以UDP的形式写日志
type UDPWriter struct {
	URI  string
	conn net.Conn
	m    sync.Mutex
}

// Write 写日志
func (w *UDPWriter) Write(buf []byte) error {
	w.m.Lock()
	defer w.m.Unlock()
	if w.conn == nil {
		conn, err := net.Dial("udp", w.URI)
		if err != nil {
			return err
		}
		w.conn = conn
	}
	_, err := w.conn.Write(buf)
	return err
}

// Close 关闭udp连接
func (w *UDPWriter) Close() error {
	if w.conn != nil {
		return w.conn.Close()
	}
	return nil
}

// Parse 转换日志的输出格式
func Parse(desc []byte) []*Tag {
	reg := regexp.MustCompile(`\{[\S]+\}`)

	index := 0
	arr := make([]*Tag, 0)
	fillCategory := "fill"
	for {
		result := reg.FindIndex(desc[index:])
		if result == nil {
			break
		}
		start := index + result[0]
		end := index + result[1]
		if start != index {
			arr = append(arr, &Tag{
				category: fillCategory,
				data:     desc[index:start],
			})
		}
		k := desc[start+1 : end-1]
		switch k[0] {
		case byte('~'):
			arr = append(arr, &Tag{
				category: cookie,
				data:     k[1:],
			})
		case byte('>'):
			arr = append(arr, &Tag{
				category: requestHeader,
				data:     k[1:],
			})
		case byte('<'):
			arr = append(arr, &Tag{
				category: responseHeader,
				data:     k[1:],
			})
		default:
			arr = append(arr, &Tag{
				category: string(k),
				data:     nil,
			})
		}
		index = result[1] + index
	}
	if index < len(desc) {
		arr = append(arr, &Tag{
			category: fillCategory,
			data:     desc[index:],
		})
	}
	return arr
}

// Format 格式化访问日志信息
func Format(ctx *fasthttp.RequestCtx, tags []*Tag, startedAt time.Time) []byte {
	fn := func(tag *Tag) []byte {
		switch tag.category {
		case host:
			return ctx.Host()
		case method:
			return ctx.Method()
		case path:
			return ctx.Path()
		case proto:
			if ctx.Request.Header.IsHTTP11() {
				return http11
			}
			return http10
		case query:
			return ctx.QueryArgs().QueryString()
		case remote:
			return []byte(ctx.RemoteIP().String())
		case clientIP:
			return []byte(util.GetClientIP(ctx))
		case scheme:
			if ctx.IsTLS() {
				return https
			}
			return http
		case uri:
			return ctx.URI().RequestURI()
		case cookie:
			return ctx.Request.Header.CookieBytes(tag.data)
		case requestHeader:
			return ctx.Request.Header.PeekBytes(tag.data)
		case responseHeader:
			return ctx.Response.Header.PeekBytes(tag.data)
		case referer:
			return ctx.Referer()
		case userAgent:
			return ctx.UserAgent()
		case when:
			return []byte(time.Now().Format(time.RFC1123Z))
		case whenISO:
			return []byte(time.Now().UTC().Format(time.RFC3339))
		case whenISOMs:
			return []byte(time.Now().UTC().Format("2006-01-02T15:04:05.999Z07:00"))
		case whenUnix:
			return []byte(strconv.FormatInt(time.Now().Unix(), 10))
		case status:
			return []byte(strconv.Itoa(ctx.Response.StatusCode()))
		case size:
			return []byte(strconv.Itoa(len(ctx.Response.Body())))
		case latency:
			return []byte(time.Since(startedAt).String())
		case latencyMs:
			offset := (time.Now().UnixNano() - startedAt.UnixNano()) / (1000 * 1000)
			return []byte(strconv.FormatInt(offset, 10))
		default:
			return tag.data
		}
	}
	arr := make([][]byte, 0)
	for _, tag := range tags {
		arr = append(arr, fn(tag))
	}
	return bytes.Join(arr, []byte(""))
}
