package cache

import (
	"bytes"
	"log"
	"strconv"
	"testing"
	"time"

	"../util"
	"../vars"

	"github.com/valyala/fasthttp"
)

func TestDB(t *testing.T) {
	_, err := InitDB("/tmp/pike")
	if err != nil {
		t.Fatalf("open db fail, %v", err)
	}
	key := []byte("/users/me")
	data := []byte("vicanso")
	err = Save(key, data, 200)
	if err != nil {
		t.Fatalf("save data fail %v", err)
	}
	buf, err := Get(key)
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

	saveRespData := &ResponseData{
		CreatedAt:  util.GetSeconds(),
		StatusCode: 200,
		Compress:   vars.GzipData,
		TTL:        ttl,
		Header:     resHeader,
		Body:       resBody,
	}

	SaveResponseData(key, saveRespData)
	respData, err := GetResponse(key)
	if err != nil {
		t.Fatalf("get the response fail, %v", err)
	}
	if uint32(time.Now().Unix())-respData.CreatedAt > 1 {
		t.Fatalf("get the create time fail")
	}
	if respData.TTL != 30 {
		t.Fatalf("get the ttle fail")
	}
	checkHeader := []byte("Server: fasthttp\r\nCache-Control: public, max-age=30")
	if bytes.Compare(respData.Header, checkHeader) != 0 {
		t.Fatalf("the response header fail")
	}
	if bytes.Compare(respData.Body, data) != 0 {
		t.Fatalf("the response body fail")
	}
	if respData.Compress != vars.GzipData {
		t.Fatalf("the data should be gzip compress")
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
	size := Size()
	if size != 2 {
		t.Fatalf("the cache size expect 2 but %v", size)
	}
	fetchingCount, waitingCount, cacheableCount, hitForPassCount := Stats()
	if fetchingCount != 0 || waitingCount != 0 {
		t.Fatalf("the fecthing and wating count is wrong")
	}
	if cacheableCount != 1 {
		t.Fatalf("the cacheable couunt expect 1 but %v", cacheableCount)
	}
	if hitForPassCount != 1 {
		t.Fatalf("the hit for pass couunt expect 1 but %v", hitForPassCount)
	}
	time.Sleep(time.Second)
}

func TestResponseCache(t *testing.T) {
	// 测试生成插入多条记录，将对过期数据删除
	startedAt := time.Now()
	count := 10 * 1024
	for index := 0; index < count; index++ {
		key := []byte("test-" + strconv.Itoa(index))
		SaveResponseData(key, &ResponseData{
			CreatedAt:  util.GetSeconds(),
			StatusCode: 200,
			Compress:   vars.RawData,
			TTL:        10,
			Header:     make([]byte, 3*1024),
			Body:       make([]byte, 50*1024),
		})
	}
	log.Printf("create %v use %v", count, time.Since(startedAt))
	ClearExpired()
	Close()
}
