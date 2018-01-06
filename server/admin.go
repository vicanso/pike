package server

import (
	"encoding/json"

	"../director"
	"../dispatch"
	"../performance"
	"../vars"
	"github.com/tidwall/gjson"
	"github.com/valyala/fasthttp"
)

func responseJSON(ctx *fasthttp.RequestCtx, data []byte) {
	ctx.SetContentTypeBytes(vars.JSON)
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

func adminHandler(ctx *fasthttp.RequestCtx, directorList director.DirectorSlice, blackIP *BlackIP) {
	ctx.Response.Header.SetCanonical(vars.CacheControl, vars.NoCache)
	switch string(ctx.Path()) {
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
