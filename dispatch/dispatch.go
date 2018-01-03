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
	header := respData.Header
	body := respData.Body
	arr := bytes.Split(header, vars.LineBreak)
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
		ctx.Response.Header.SetCanonical(k, v)
	}
	if respData.TTL > 0 {
		age := util.GetSeconds() - int64(respData.CreatedAt)
		ctx.Response.Header.SetCanonical(vars.Age, []byte(strconv.Itoa(int(age))))
	}
	ctx.Response.Header.SetContentLength(len(body))
	ctx.Response.Header.SetCanonical(vars.Server, vars.ServerName)
	ctx.SetBody(body)
}
