package cache

import (
	"bytes"
	"testing"
	"time"

	"../util"
	"../vars"
	"github.com/valyala/fasthttp"
)

func TestDB(t *testing.T) {
	db, err := InitDB("/tmp/pike.db")
	if err != nil {
		t.Fatalf("open db fail, %v", err)
	}
	defer db.Close()
	bucket := []byte("aslant")
	err = InitBucket(bucket)
	if err != nil {
		t.Fatalf("init bucket fail, %v", err)
	}
	key := []byte("/users/me")
	data := []byte("vicanso")
	err = Save(bucket, key, data)
	if err != nil {
		t.Fatalf("save data fail %v", err)
	}
	buf, err := Get(bucket, key)
	if err != nil {
		t.Fatalf("get data fail %v", err)
	}
	if bytes.Compare(data, buf) != 0 {
		t.Fatalf("get data fail")
	}

	ctx := &fasthttp.RequestCtx{}
	ctx.Response.SetBody(data)
	ctx.Response.Header.SetCanonical(vars.CacheControl, []byte("public, max-age=30"))

	resBody := ctx.Response.Body()
	resHeader := ctx.Response.Header.Header()
	ttl := util.GetCacheAge(ctx)

	SaveResponseData(bucket, key, resBody, resHeader, ttl)
	respData, err := GetResponse(bucket, key)
	if err != nil {
		t.Fatalf("get the response fail, %v", err)
	}
	if uint32(time.Now().Unix())-respData.CreatedAt > 1 {
		t.Fatalf("get the create time fail")
	}
	if respData.TTL != 30 {
		t.Fatalf("get the ttle fail")
	}
	if bytes.Compare(respData.Header, ctx.Response.Header.Header()) != 0 {
		t.Fatalf("the response header fail")
	}
	if bytes.Compare(respData.Body, data) != 0 {
		t.Fatalf("the response body fail")
	}
}

func TestRequestStatus(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("http://aslant.site/users/me")
	status := GetRequestStatus(ctx)
	// 第一次请求时，状态为fetching
	if status != vars.Fetching {
		t.Fatalf("the first request should be fetching")
	}
	ctx = &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("http://aslant.site/users/me")
	status = GetRequestStatus(ctx)
	// 此时有相同的请求在处理，该请求状态为waiting
	if status != vars.Waiting {
		t.Fatalf("the second request should be wating")
	}
	// 获取等待的请求，数量为一
	list := GetWaitingRequests(util.GenRequestKey(ctx))
	if len(list) != 1 {
		t.Fatalf("the wating list count is wrong")
	}
	if list[0] != ctx {
		t.Fatalf("the wating request is wrong")
	}

	//测试 hit for pass
	ctx = &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("http://aslant.site/users")
	status = GetRequestStatus(ctx)
	// 第一次请求时，状态为fetching
	if status != vars.Fetching {
		t.Fatalf("the first request should be fetching")
	}
	GetWatingRequstAndSetStatus(util.GenRequestKey(ctx), vars.HitForPass, 1)
	ctx = &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("http://aslant.site/users")
	status = GetRequestStatus(ctx)
	if status != vars.HitForPass {
		t.Fatalf("the request shoud be hit for pass")
	}
	time.Sleep(2 * time.Second)

	ctx = &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("http://aslant.site/users")
	status = GetRequestStatus(ctx)
	// hti for pass 已过期，状态为fetching
	if status != vars.Fetching {
		t.Fatalf("the request should be fetching")
	}

	ctx = &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("http://aslant.site/books")
	status = GetRequestStatus(ctx)
	if status != vars.Fetching {
		t.Fatalf("the request should be fetching")
	}

	GetWatingRequstAndSetStatus(util.GenRequestKey(ctx), vars.Cacheable, 100)

	ctx = &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("http://aslant.site/books")
	status = GetRequestStatus(ctx)

}
