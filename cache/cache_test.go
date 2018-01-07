package cache

import (
	"bytes"
	"log"
	"strconv"
	"testing"
	"time"

	"../util"
	"../vars"
	"github.com/boltdb/bolt"
	"github.com/valyala/fasthttp"
)

func TestDB(t *testing.T) {
	_, err := InitDB("/tmp/pike.db")
	if err != nil {
		t.Fatalf("open db fail, %v", err)
	}
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

	saveRespData := &ResponseData{
		CreatedAt:  util.GetSeconds(),
		StatusCode: 200,
		Compress:   vars.GzipData,
		TTL:        ttl,
		Header:     resHeader,
		Body:       resBody,
	}

	SaveResponseData(bucket, key, saveRespData)
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
	time.Sleep(time.Second)
}

func TestResponseCache(t *testing.T) {
	InitDB("/tmp/pike.db")
	bucket := []byte("aslant")
	InitBucket(bucket)
	var ttl uint32 = 60
	// 设置数据保存后则刚过期
	SaveResponseData(bucket, []byte("response-a"), &ResponseData{
		CreatedAt:  util.GetSeconds() - ttl,
		StatusCode: 200,
		Compress:   vars.RawData,
		TTL:        ttl,
		Header:     []byte("response-a-header"),
		Body:       []byte("response-a-body"),
	})

	SaveResponseData(bucket, []byte("response-b"), &ResponseData{
		CreatedAt:  util.GetSeconds() - ttl - 10,
		StatusCode: 200,
		Compress:   vars.RawData,
		TTL:        ttl,
		Header:     []byte("response-b-header"),
		Body:       []byte("response-b-body"),
	})

	err := ClearExpiredResponseData(bucket)
	if err != nil {
		t.Fatalf("clear expired response data fail, %v", err)
	}
	data, err := GetResponse(bucket, []byte("response-b"))
	if data != nil {
		t.Fatalf("clear expired response data fail, the response-b should be remove")
	}

	// 测试生成插入多条记录，将对过期数据删除
	startedAt := time.Now()
	count := 10 * 1024
	for index := 0; index < count; index++ {
		key := []byte("test-" + strconv.Itoa(index))
		offset := (ttl + 10)
		if index%2 == 0 {
			offset = 0
		}
		SaveResponseData(bucket, key, &ResponseData{
			CreatedAt:  util.GetSeconds() - offset,
			StatusCode: 200,
			Compress:   vars.RawData,
			TTL:        ttl,
			Header:     make([]byte, 3*1024),
			Body:       make([]byte, 50*1024),
		})
	}
	log.Printf("create %v use %v", count, time.Since(startedAt))
	getCount := func() int {
		count := 0
		client := GetClient()
		client.View(func(tx *bolt.Tx) error {
			c := tx.Bucket(bucket).Cursor()
			prefix := []byte("test-")
			for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
				if v != nil {
					count++
				}
			}
			return nil
		})
		return count
	}
	startedAt = time.Now()
	currnetCount := getCount()
	if currnetCount != count {
		t.Fatalf("get the test- cache expect %v but %v", count, currnetCount)
	}
	log.Printf("count %v use %v", count, time.Since(startedAt))

	startedAt = time.Now()
	err = ClearExpiredResponseData(bucket)
	if err != nil {
		t.Fatalf("clear expired response data fail, %v", err)
	}
	log.Printf("clear expired data %v use %v", count/2, time.Since(startedAt))
	currnetCount = getCount()
	if currnetCount != count/2 {
		t.Fatalf("get the test- cache expect %v but %v", count/2, currnetCount)
	}
	client.Close()
}
