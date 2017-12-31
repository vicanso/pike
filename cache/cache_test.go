package cache

import (
	"bytes"
	"testing"
	"time"

	"../util"
	"../vars"
	"github.com/valyala/fasthttp"
)

func TestStatus(t *testing.T) {
	key := "aslant.site/users/me"
	status := GetStatus(key)
	none := vars.None
	if status != none {
		t.Fatalf("the status should be none")
	}
	status = GetStatus(key)
	if status == none {
		t.Fatalf("the status should not be none")
	}
	DeleteStatus(key)
	status = GetStatus(key)
	if status != none {
		t.Fatalf("the status should be none")
	}
	SetHitForPass(key)
	status = GetStatus(key)
	if status != vars.HitForPass {
		t.Fatalf("the status should be hit for pass")
	}
}

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
