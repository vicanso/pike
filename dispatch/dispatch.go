package dispatch

import (
	"bytes"
	"strconv"

	"../cache"
	"../util"
	"../vars"

	"github.com/valyala/fasthttp"
)

// GetResponseHeader 获取响应的header
func GetResponseHeader(resp *fasthttp.Response) []byte {
	newHeader := &fasthttp.ResponseHeader{}
	resp.Header.CopyTo(newHeader)
	newHeader.DelBytes(vars.ContentEncoding)
	newHeader.DelBytes(vars.ContentLength)
	return newHeader.Header()
}

// GetResponseBody 获取响应的数据
func GetResponseBody(resp *fasthttp.Response) ([]byte, error) {
	enconding := resp.Header.PeekBytes(vars.ContentEncoding)

	if bytes.Compare(enconding, vars.Gzip) == 0 {
		return resp.BodyGunzip()
	}
	return resp.Body(), nil
}

// ErrorHandler 出错处理
func ErrorHandler(ctx *fasthttp.RequestCtx, err error) {
	// TODO 出错的处理，504 502等
	switch err {
	case vars.ErrDirectorUnavailable:
		ctx.SetStatusCode(503)
	case vars.ErrServiceUnavailable:
		ctx.SetStatusCode(503)
	default:
		ctx.SetStatusCode(500)
	}
	ctx.SetBodyString(err.Error())
}

// Response 响应数据
func Response(ctx *fasthttp.RequestCtx, respData *cache.ResponseData) {
	respHeader := &ctx.Response.Header
	header := respData.Header
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
		v := item[index : len(item)-1]
		respHeader.SetCanonical(k, v)
	}
	if respData.TTL > 0 {
		age := util.GetSeconds() - respData.CreatedAt
		respHeader.SetCanonical(vars.Age, []byte(strconv.Itoa(int(age))))
	}
	statusCode := int(respData.StatusCode)
	// TODO 304
	// method := ctx.Method()
	// if bytes.Compare(method, vars.Get) == 0 || bytes.Compare(method, vars.Head) == 0 {

	// }

	supportGzip := ctx.Request.Header.HasAcceptEncodingBytes(vars.Gzip)
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
	} else if supportGzip && bodyLength > vars.CompressMinLength {
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
