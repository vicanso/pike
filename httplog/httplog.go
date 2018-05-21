package httplog

import (
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo"

	"github.com/vicanso/pike/util"
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
	sizeHuman      = "size-human"
	status         = "status"
	latency        = "latency"
	latencyMs      = "latency-ms"
	cookie         = "cookie"
	payloadSize    = "payload-size"
	requestHeader  = "requestHeader"
	responseHeader = "responseHeader"
	httpProto      = "HTTP"
	httpsProto     = "HTTPS"
)

// Tag log tag
type Tag struct {
	category string
	data     string
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
	m        sync.RWMutex
	date     string
	file     string
}

func (w *FileWriter) checkDate() {
	time.Sleep(10 * time.Second)
	now := time.Now()
	date := now.Format("2006-01-02")
	// 如果日期有变化
	if w.date != date {
		w.m.Lock()
		w.fd.Close()
		w.fd = nil
		w.m.Unlock()
	} else {
		w.checkDate()
	}
}

func (w *FileWriter) initFd() error {
	if w.fd != nil {
		return nil
	}
	w.m.Lock()
	defer w.m.Unlock()
	// 如果有并发的处理已生成fd，直接返回
	if w.fd != nil {
		return nil
	}
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
	fd, err := os.OpenFile(w.file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	w.fd = fd
	// 如果是以按天生成日志，增加定时检测
	if w.Category == Date {
		go w.checkDate()
	}
	return nil
}

// Write 写日志
func (w *FileWriter) Write(buf []byte) error {
	err := w.initFd()
	if err == nil {
		w.m.RLock()
		w.fd.Write(append(buf, '\n'))
		w.m.RUnlock()
	}
	return err
}

// Close 关闭写文件
func (w *FileWriter) Close() error {
	w.m.Lock()
	defer w.m.Unlock()
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
	w.m.Lock()
	defer w.m.Unlock()
	if w.conn != nil {
		return w.conn.Close()
	}
	return nil
}

// Parse 转换日志的输出格式
func Parse(desc []byte) []*Tag {
	reg := regexp.MustCompile(`\{[\S]+?\}`)

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
				data:     string(desc[index:start]),
			})
		}
		k := desc[start+1 : end-1]
		switch k[0] {
		case byte('~'):
			arr = append(arr, &Tag{
				category: cookie,
				data:     string(k[1:]),
			})
		case byte('>'):
			arr = append(arr, &Tag{
				category: requestHeader,
				data:     string(k[1:]),
			})
		case byte('<'):
			arr = append(arr, &Tag{
				category: responseHeader,
				data:     string(k[1:]),
			})
		default:
			arr = append(arr, &Tag{
				category: string(k),
				data:     "",
			})
		}
		index = result[1] + index
	}
	if index < len(desc) {
		arr = append(arr, &Tag{
			category: fillCategory,
			data:     string(desc[index:]),
		})
	}
	return arr
}

// Format 格式化访问日志信息
func Format(c echo.Context, tags []*Tag, startedAt time.Time) string {
	fn := func(tag *Tag) string {
		switch tag.category {
		case host:
			return c.Request().Host
		case method:
			return c.Request().Method
		case path:
			p := c.Request().URL.Path
			if p == "" {
				p = "/"
			}
			return p
		case proto:
			return c.Request().Proto
		case query:
			return c.QueryString()
		case remote:
			return c.Request().RemoteAddr
		case clientIP:
			return c.RealIP()
		case scheme:
			if c.IsTLS() {
				return httpsProto
			}
			return httpProto
		case uri:
			return c.Request().RequestURI
		case cookie:
			cookie, err := c.Cookie(tag.data)
			if err != nil {
				return ""
			}
			return cookie.Value
		case requestHeader:
			return c.Request().Header.Get(tag.data)
		case responseHeader:
			return c.Response().Header().Get(tag.data)
		case referer:
			return c.Request().Referer()
		case userAgent:
			return c.Request().UserAgent()
		case when:
			return time.Now().Format(time.RFC1123Z)
		case whenISO:
			return time.Now().UTC().Format(time.RFC3339)
		case whenISOMs:
			return time.Now().UTC().Format("2006-01-02T15:04:05.999Z07:00")
		case whenUnix:
			return strconv.FormatInt(time.Now().Unix(), 10)
		case status:
			return strconv.Itoa(c.Response().Status)
		// case payloadSize:
		// 	return []byte(strconv.Itoa(len(ctx.Request.Body())))
		case size:
			return strconv.FormatInt(c.Response().Size, 10)
		case sizeHuman:
			return util.GetHumanReadableSize(c.Response().Size)
		case latency:
			return time.Since(startedAt).String()
		case latencyMs:
			ms := util.GetTimeConsuming(startedAt)
			return strconv.Itoa(ms)
		default:
			return tag.data
		}
	}

	arr := make([]string, 0, len(tags))
	for _, tag := range tags {
		arr = append(arr, fn(tag))
	}
	return strings.Join(arr, "")
}
