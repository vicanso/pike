package server

import (
	"encoding/json"
	"strings"

	"../director"
	"../dispatch"
	"../performance"
	"../util"
	"../vars"
	"github.com/tidwall/gjson"
	"github.com/valyala/fasthttp"
)

// responseJSON 返回json数据
func responseJSON(ctx *fasthttp.RequestCtx, data []byte) {
	ctx.SetContentTypeBytes(vars.JSON)
	if len(data) > vars.CompressMinLength {
		rawData, err := util.Gzip(data)
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
		ctx.SetStatusCode(201)
	case "DELETE":
		body := string(ctx.Request.Body())
		value := gjson.Get(body, "ip")
		blockIP.Remove(value.String())
		ctx.SetStatusCode(204)
	default:
		ctx.NotFound()
	}
}

// statisHandler 静态文件处理
func statisHandler(ctx *fasthttp.RequestCtx, assetPath string) {
	path := string(ctx.Path())
	file := path[len(assetPath):]
	data, err := Asset("assets/dist/" + file)
	if err != nil {
		ctx.NotFound()
		return
	}
	gzipData, err := util.Gzip(data)
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
func adminHandler(ctx *fasthttp.RequestCtx, directorList director.DirectorSlice, blockIP *BlockIP, conf *PikeConfig) {
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
		}
		responseJSON(ctx, stats)
	case adminPath + "/directors":
		data, err := json.Marshal(directorList)
		if err != nil {
			dispatch.ErrorHandler(ctx, err)
		}
		responseJSON(ctx, data)
	case adminPath + "/block-ips":
		blockIPHandler(ctx, blockIP)
	default:
		ctx.NotFound()
	}
}
