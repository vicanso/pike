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
		value := gjson.Get(body, "ip")
		blackIP.Add(value.String())
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

func statisHandler(ctx *fasthttp.RequestCtx, assetPath string) {
	path := string(ctx.Path())
	file := path[len(assetPath):]
	data, err := Asset("assets/dist/" + file)
	if err != nil {
		ctx.NotFound()
		return
	}
	if strings.HasSuffix(file, ".html") {
		ctx.SetContentType("text/html; charset=utf-8")
	} else {
		ctx.SetContentType("application/javascript")
	}
	ctx.SetBody(data)
}

func adminHandler(ctx *fasthttp.RequestCtx, directorList director.DirectorSlice, blackIP *BlackIP) {
	ctx.Response.Header.SetCanonical(vars.CacheControl, vars.NoCache)
	path := string(ctx.Path())
	assetPath := "/pike/admin/"
	if strings.HasPrefix(path, assetPath) {
		statisHandler(ctx, assetPath)
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
