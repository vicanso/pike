package dispatch

import (
	"bytes"
	"strconv"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/vicanso/fresh"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/util"
	"github.com/vicanso/pike/vars"
)

// ErrorHandler 出错处理
func ErrorHandler(ctx *fasthttp.RequestCtx, err error) {
	switch err {
	case vars.ErrDirectorUnavailable, vars.ErrServiceUnavailable:
		ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
	case vars.ErrGatewayTimeout:
		ctx.SetStatusCode(fasthttp.StatusGatewayTimeout)
	default:
		// TODO 非主动抛出的出错，是否需要输出日志等
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
	}
	ctx.Response.Header.SetCanonical(vars.CacheControl, vars.NoCache)
	ctx.SetBodyString(err.Error())
}

// Response 响应数据
func Response(ctx *fasthttp.RequestCtx, respData *cache.ResponseData) {
	respHeader := &ctx.Response.Header
	reqHeader := &ctx.Request.Header
	header := respData.Header
	shouldCompress := respData.ShouldCompress
	body := respData.Body
	bodyLength := len(body)
	arr := bytes.Split(header, vars.LineBreak)
	// 设置响应头
	for _, item := range arr {
		index := bytes.IndexByte(item, vars.Colon)
		if index == -1 {
			continue
		}
		k := item[:index]
		index++
		if item[index] == vars.Space {
			index++
		}
		v := item[index:len(item)]
		respHeader.SetCanonical(k, v)
	}
	// 如果该请求是可缓存的，要设置Age（因为客户端缓存的时间应该是 ttl - age）
	if respData.TTL > 0 {
		now := uint32(time.Now().Unix())
		age := now - respData.CreatedAt
		respHeader.SetCanonical(vars.Age, []byte(strconv.Itoa(int(age))))
	}

	statusCode := int(respData.StatusCode)
	method := ctx.Method()

	// 只对于GET HEAD请求做304的判断
	if bytes.Compare(method, vars.Get) == 0 || bytes.Compare(method, vars.Head) == 0 {
		// 响应的状态码需要为20x或304
		if (statusCode >= 200 && statusCode < 300) || statusCode == 304 {
			requestHeaderData := &fresh.RequestHeader{
				IfModifiedSince: reqHeader.PeekBytes(vars.IfModifiedSince),
				IfNoneMatch:     reqHeader.PeekBytes(vars.IfNoneMatch),
				CacheControl:    reqHeader.PeekBytes(vars.CacheControl),
			}
			respHeaderData := &fresh.ResponseHeader{
				ETag:         respHeader.PeekBytes(vars.ETag),
				LastModified: respHeader.PeekBytes(vars.LastModified),
			}
			// 304
			if fresh.Fresh(requestHeaderData, respHeaderData) {
				ctx.Response.ResetBody()
				ctx.SetStatusCode(fasthttp.StatusNotModified)
				return
			}
		}
	}

	supportGzip := reqHeader.HasAcceptEncodingBytes(vars.Gzip)
	// 如果数据是gzip
	if respData.Compress == vars.GzipData {
		// 如果客户端不支持gzip，则解压
		if !supportGzip {
			rawData, err := util.Gunzip(body)
			if err != nil {
				ErrorHandler(ctx, err)
				return
			}
			body = rawData
			bodyLength = len(body)
		} else {
			// 客户端支持则设置gzip encoding
			respHeader.SetCanonical(vars.ContentEncoding, vars.Gzip)
		}
	} else if supportGzip && shouldCompress && bodyLength > vars.CompressMinLength {
		// 支持gzip，但是数据未压缩，而且数据大于 CompressMinLength
		gzipData, err := util.Gzip(body)
		// 如果压缩失败，直接返回未压缩数据
		if err == nil {
			body = gzipData
			bodyLength = len(body)
			respHeader.SetCanonical(vars.ContentEncoding, vars.Gzip)
		}
	}
	ctx.Response.SetStatusCode(statusCode)
	respHeader.SetContentLength(bodyLength)
	ctx.SetBody(body)
}
