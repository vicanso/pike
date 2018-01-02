package dispatch

import (
	"bytes"

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

// ResponseBytes 数据以字节返回
func ResponseBytes(ctx *fasthttp.RequestCtx, header, buf []byte) {
	arr := bytes.Split(header, []byte("\n"))
	divide := []byte(":")[0]
	space := []byte(" ")[0]
	for _, item := range arr {
		index := bytes.IndexByte(item, divide)
		if index == -1 {
			continue
		}
		k := item[:index]
		index++
		if item[index] == space {
			index++
		}
		v := item[index : len(item)-1]
		ctx.Response.Header.SetCanonical(k, v)
	}
	ctx.Response.Header.SetContentLength(len(buf))
	ctx.SetBody(buf)
}

// ResponseGzipBytes 数据以压缩后的字节返回
func ResponseGzipBytes(ctx *fasthttp.RequestCtx, header, buf []byte) {
	ctx.Response.Header.SetCanonical(vars.ContentEncoding, vars.Gzip)
	ResponseBytes(ctx, header, buf)
}
