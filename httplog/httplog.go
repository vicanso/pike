package httplog

import (
	"bytes"
	"regexp"
	"strconv"
	"time"

	"../util"

	"github.com/valyala/fasthttp"
)

const (
	host      = "host"
	method    = "method"
	path      = "path"
	proto     = "proto"
	query     = "query"
	remote    = "remote"
	scheme    = "scheme"
	uri       = "uri"
	referer   = "referer"
	userAgent = "user-agent"
	when      = "when"
	whenISO   = "when_iso"
	whenUnix  = "when_unix"
	size      = "size"
	status    = "status"
	latency   = "latency"
	latencyMs = "latency_ms"
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
		k := string(desc[start+1 : end-1])
		arr = append(arr, &Tag{
			category: k,
			data:     nil,
		})
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
			return []byte(util.GetClientIP(ctx))
		case scheme:
			if ctx.IsTLS() {
				return https
			}
			return http
		case uri:
			return ctx.URI().RequestURI()
		case referer:
			return ctx.Referer()
		case userAgent:
			return ctx.UserAgent()
		case when:
			return []byte(time.Now().Format(time.RFC1123Z))
		case whenISO:
			return []byte(time.Now().UTC().Format(time.RFC3339))
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
