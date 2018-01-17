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

// blackIPHandler 黑名单IP的配置处理
func blackIPHandler(ctx *fasthttp.RequestCtx, blackIP *BlackIP) {
	method := string(ctx.Method())
	switch method {
	case "GET":
		data, err := json.Marshal(blackIP)
		if err != nil {
			dispatch.ErrorHandler(ctx, err)
		}
		responseJSON(ctx, data)
	case "POST":
		body := string(ctx.Request.Body())
		value := gjson.Get(body, "ip").String()
		if len(value) != 0 {
			blackIP.Add(value)
		}
		ctx.SetStatusCode(201)
	case "DELETE":
		body := string(ctx.Request.Body())
		value := gjson.Get(body, "ip")
		blackIP.Remove(value.String())
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
func adminHandler(ctx *fasthttp.RequestCtx, directorList director.DirectorSlice, blackIP *BlackIP, conf *PikeConfig) {
	ctx.Response.Header.SetCanonical(vars.CacheControl, vars.NoCache)
	path := string(ctx.Path())
	assetPath := "/pike/admin/"
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
	case "/pike/stats":
		stats, err := json.Marshal(performance.GetStats())
		if err != nil {
			dispatch.ErrorHandler(ctx, err)
		}
		responseJSON(ctx, stats)
	case "/pike/directors":
		data, err := json.Marshal(directorList)
		if err != nil {
			dispatch.ErrorHandler(ctx, err)
		}
		responseJSON(ctx, data)
	case "/pike/black-ips":
		blackIPHandler(ctx, blackIP)
	default:
		ctx.NotFound()
	}
}
