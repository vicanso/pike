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
	ttl := util.GetCacheAge(&ctx.Response.Header)

	SaveResponseData(bucket, key, resBody, resHeader, 200, ttl)
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
	key := []byte("GEThttp://aslant.site/users/me")
	status, c := GetRequestStatus(key)
	// 第一次请求时，状态为fetching
	if status != vars.Fetching {
		t.Fatalf("the first request should be fetching")
	}
	if c != nil {
		t.Fatalf("the chan of first request should be nil")
	}

	status, c = GetRequestStatus(key)
	if status != vars.Waiting {
		t.Fatalf("the second request should be wating for the first request result")
	}
	if c == nil {
		t.Fatalf("the chan of second request shouldn't be nil")
	}
	go func(tmp chan int) {
		tmpStatus := <-tmp
		if tmpStatus != vars.HitForPass {
			t.Fatalf("the waiting request should be hit for pass")
		}
	}(c)

	HitForPass(key, 100)
	time.Sleep(time.Second)

	key = []byte("GEThttp://aslant.site/books")
	status, c = GetRequestStatus(key)
	// 第一次请求时，状态为fetching
	if status != vars.Fetching {
		t.Fatalf("the first request should be fetching")
	}
	if c != nil {
		t.Fatalf("the chan of first request should be nil")
	}

	status, c = GetRequestStatus(key)
	if status != vars.Waiting {
		t.Fatalf("the second request should be wating for the first request result")
	}
	if c == nil {
		t.Fatalf("the chan of second request shouldn't be nil")
	}
	go func(tmp chan int) {
		tmpStatus := <-tmp
		if tmpStatus != vars.Cacheable {
			t.Fatalf("the waiting request should be cacheable")
		}
	}(c)

	Cacheable(key, 100)
	time.Sleep(time.Second)
}
