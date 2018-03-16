package dispatch

import (
	"bytes"
	"log"
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
	case vars.ErrAccessIsNotAlloed:
		ctx.SetStatusCode(fasthttp.StatusForbidden)
	default:
		log.Print("internal server error:", err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
	}
	ctx.Response.Header.SetCanonical(vars.CacheControl, vars.NoCache)
	ctx.SetBodyString(err.Error())
}

func getResponseData(header *fasthttp.RequestHeader, respData *cache.ResponseData) ([]byte, []byte, error) {
	gzipData := respData.GzipBody
	brData := respData.BrBody
	if header.HasAcceptEncodingBytes(vars.Br) && len(brData) != 0 {
		return brData, vars.Br, nil
	}
	if header.HasAcceptEncodingBytes(vars.Gzip) && len(gzipData) != 0 {
		return gzipData, vars.Gzip, nil
	}
	body := respData.Body
	// 如果没有body数据，而且有gzipData
	// 如果body gzipData都没有（为204的情况，正常）
	if len(body) == 0 && len(gzipData) != 0 {
		data, err := util.Gunzip(gzipData)
		if err != nil {
			return nil, nil, err
		}
		body = data
	}
	return body, nil, nil
}

// Response 响应数据
func Response(ctx *fasthttp.RequestCtx, respData *cache.ResponseData) {
	respHeader := &ctx.Response.Header
	reqHeader := &ctx.Request.Header
	header := respData.Header
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
	if bytes.Equal(method, vars.Get) || bytes.Equal(method, vars.Head) {
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
	ctx.Response.SetStatusCode(statusCode)
	if statusCode == fasthttp.StatusNoContent {
		return
	}
	body, encoding, err := getResponseData(reqHeader, respData)
	if err != nil {
		ErrorHandler(ctx, err)
		return
	}
	if len(encoding) != 0 {
		respHeader.SetCanonical(vars.ContentEncoding, encoding)
	}
	respHeader.SetContentLength(len(body))
	ctx.SetBody(body)
}
