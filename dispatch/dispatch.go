package dispatch

import (
	"bytes"
	"compress/gzip"

	"../vars"
	"github.com/valyala/fasthttp"
)

// 对数据压缩
func doGzip(buf []byte) ([]byte, error) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	defer w.Close()
	_, err := w.Write(buf)
	if err != nil {
		return nil, err
	}
	w.Flush()
	return b.Bytes(), nil
}

// 根据Content-Encoding做数据解压
func getBody(resp *fasthttp.Response) ([]byte, error) {
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

// 根据Accept-Encoding判断是否压缩数据返回
func responseBytes(ctx *fasthttp.RequestCtx, buf []byte) {
	// 如果数据小于最小压缩长度，不做压缩
	if len(buf) > vars.CompressMinLength {
		if ctx.Request.Header.HasAcceptEncodingBytes(vars.Gzip) {
			gzipBuf, _ := doGzip(buf)
			if len(gzipBuf) != 0 {
				ctx.Response.Header.SetCanonical(vars.ContentEncoding, vars.Gzip)
				buf = gzipBuf
			} else {
				ctx.Response.Header.DelBytes(vars.ContentEncoding)
			}
		} else {
			ctx.Response.Header.DelBytes(vars.ContentEncoding)
		}
	}
	ctx.Response.Header.SetContentLength(len(buf))
	ctx.SetBody(buf)
}

// Response 返回请求响应数据
func Response(ctx *fasthttp.RequestCtx, resp *fasthttp.Response) {
	resp.Header.CopyTo(&ctx.Response.Header)
	buf, err := getBody(resp)
	if err != nil {
		ErrorHandler(ctx, err)
		return
	}
	responseBytes(ctx, buf)
}
