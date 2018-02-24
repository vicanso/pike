package server

import (
	"encoding/json"
	"strings"

	"github.com/gobuffalo/packr"
	"github.com/tidwall/gjson"
	"github.com/valyala/fasthttp"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/director"
	"github.com/vicanso/pike/dispatch"
	"github.com/vicanso/pike/performance"
	"github.com/vicanso/pike/proxy"
	"github.com/vicanso/pike/util"
	"github.com/vicanso/pike/vars"
)

// responseJSON 返回json数据
func responseJSON(ctx *fasthttp.RequestCtx, data []byte) {
	ctx.SetContentTypeBytes(vars.JSON)
	if len(data) > vars.CompressMinLength {
		rawData, err := util.Gzip(data, 0)
		if err == nil {
			data = rawData
			ctx.Response.Header.SetCanonical(vars.ContentEncoding, vars.Gzip)
		}
	}
	ctx.SetBody(data)
}

// blockIPHandler 黑名单IP的配置处理
func blockIPHandler(ctx *fasthttp.RequestCtx, blockIP *BlockIP) {
	method := string(ctx.Method())
	switch method {
	case "GET":
		data, err := json.Marshal(blockIP)
		if err != nil {
			dispatch.ErrorHandler(ctx, err)
		}
		responseJSON(ctx, data)
	case "POST":
		body := string(ctx.Request.Body())
		value := gjson.Get(body, "ip").String()
		if len(value) != 0 {
			blockIP.Add(value)
		}
		ctx.SetStatusCode(fasthttp.StatusCreated)
	case "DELETE":
		body := string(ctx.Request.Body())
		value := gjson.Get(body, "ip")
		blockIP.Remove(value.String())
		ctx.SetStatusCode(fasthttp.StatusNoContent)
	default:
		ctx.NotFound()
	}
}

// cachedHandler cache的处理
func cachedHandler(ctx *fasthttp.RequestCtx) {
	method := string(ctx.Method())
	switch method {
	case "DELETE":
		body := string(ctx.Request.Body())
		value := gjson.Get(body, "key").String()
		cache.Expire([]byte(value))
		ctx.SetStatusCode(fasthttp.StatusNoContent)
	default:
		data := cache.GetCachedList()
		responseJSON(ctx, data)
	}
}

// statisHandler 静态文件处理
func statisHandler(ctx *fasthttp.RequestCtx, assetPath string) {
	path := string(ctx.Path())
	file := path[len(assetPath):]
	box := packr.NewBox("../assets/dist")
	data, err := box.MustBytes(file)
	if err != nil {
		ctx.NotFound()
		return
	}
	gzipData, err := util.Gzip(data, 0)
	if err == nil {
		ctx.Response.Header.SetCanonical(vars.ContentEncoding, vars.Gzip)
		data = gzipData
	}
	if strings.HasSuffix(file, ".html") {
		ctx.SetContentType("text/html; charset=utf-8")
	} else {
		ctx.SetContentType("application/javascript")
	}
	ctx.SetBody(data)
}

// adminHandler 管理员相关接口处理
func adminHandler(ctx *fasthttp.RequestCtx, conf *Config, blockIP *BlockIP) {
	ctx.Response.Header.SetCanonical(vars.CacheControl, vars.NoCache)
	path := string(ctx.Path())
	adminPath := conf.AdminPath
	assetPath := adminPath + "/admin/"
	if strings.HasPrefix(path, assetPath) {
		statisHandler(ctx, assetPath)
		return
	}
	// 对token校验
	if string(ctx.Request.Header.Peek("X-Admin-Token")) != conf.AdminToken {
		ctx.SetBody([]byte("{\"error\": \"The token is invalid\"}"))
		ctx.SetContentTypeBytes(vars.JSON)
		ctx.SetStatusCode(fasthttp.StatusUnauthorized)
		return
	}
	switch path {
	case adminPath + "/stats":
		stats, err := json.Marshal(performance.GetStats())
		if err != nil {
			dispatch.ErrorHandler(ctx, err)
			return
		}
		responseJSON(ctx, stats)
	case adminPath + "/directors":
		data, err := json.Marshal(director.List())
		if err != nil {
			dispatch.ErrorHandler(ctx, err)
			return
		}
		responseJSON(ctx, data)
	case adminPath + "/upstreams":
		data, err := json.Marshal(proxy.List())
		if err != nil {
			dispatch.ErrorHandler(ctx, err)
			return
		}
		responseJSON(ctx, data)
	case adminPath + "/block-ips":
		blockIPHandler(ctx, blockIP)
	case adminPath + "/cacheds":
		cachedHandler(ctx)
	default:
		ctx.NotFound()
	}
}
