package cache

import (
	"bytes"
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/vicanso/pike/util"
	"github.com/vicanso/pike/vars"

	"github.com/valyala/fasthttp"
)

func TestInitRequestStatus(t *testing.T) {
	rs := initRequestStatus(30)
	if rs.createdAt == 0 {
		t.Fatalf("the created at should be seconds for now")
	}
	if rs.ttl != 30 {
		t.Fatalf("the ttl should be 30s")
	}
}

func TestIsExpired(t *testing.T) {
	rs := initRequestStatus(30)
	if isExpired(rs) != false {
		t.Fatalf("the rs should not be expired")
	}
	rs.createdAt = 0
	if isExpired(rs) != true {
		t.Fatalf("the rs should be expired")
	}
}

func TestByteToUnit(t *testing.T) {
	b16 := uint16ToBytes(100)
	v16 := bytesToUint16(b16)
	if v16 != 100 {
		t.Fatalf("the uint16 to bytes fail")
	}

	b32 := uint32ToBytes(100)
	v32 := bytesToUint32(b32)
	if v32 != 100 {
		t.Fatalf("the uint32 to bytes fail")
	}
}

func TestDB(t *testing.T) {
	_, err := InitDB("/tmp/pike")
	if err != nil {
		t.Fatalf("open db fail, %v", err)
	}
	// 再次初始化直接不会抛错
	_, err = InitDB("/tmp/pike")

	if err != nil {
		t.Fatalf("init again fail, %v", err)
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

	gzipData, err := util.Gzip(data, 0)
	if err != nil {
		t.Fatalf("gzip data fail, %v", err)
	}
	brData, err := util.Brotli(data, 0)
	if err != nil {
		t.Fatalf("brotli data fail, %v", err)
	}

	saveRespData := &ResponseData{
		StatusCode: 200,
		TTL:        30,
		Header:     resHeader,
		Body:       resBody,
		GzipBody:   gzipData,
		BrBody:     brData,
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
	if len(bytes.Split(respData.Header, []byte("\r\n"))) != 6 {
		t.Fatalf("the response header fail")
	}
	if !bytes.Equal(respData.Body, data) {
		t.Fatalf("the response body fail")
	}
	if !bytes.Equal(respData.GzipBody, gzipData) {
		t.Fatalf("the gzip response body fail")
	}
	if !bytes.Equal(respData.BrBody, brData) {
		t.Fatalf("the brotli response body fail")
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

	fetchingCount, waitingCount, _, _ := Stats()
	if fetchingCount != 1 || waitingCount != 1 {
		t.Fatalf("the fetching and wainting count should be 1")
	}
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
	lsm, vLog := DataSize()
	if lsm == -1 || vLog == -1 {
		t.Fatalf("get the data size fail")
	}
	fetchingCount, waitingCount, cacheableCount, hitForPassCount := Stats()
	if fetchingCount != 0 || waitingCount != 0 {
		t.Fatalf("the fecthing and wating count is wrong")
	}
	if cacheableCount != 1 {
		t.Fatalf("the cacheable count expect 1 but %v", cacheableCount)
	}
	if hitForPassCount != 1 {
		t.Fatalf("the hit for pass count expect 1 but %v", hitForPassCount)
	}
	time.Sleep(time.Second)
	buf := GetCachedList()
	if bytes.Index(buf, []byte("GEThttp://aslant.site/books")) == -1 {
		t.Fatalf("the cache list should include GEThttp://aslant.site/books")
	}

	status, _ = GetRequestStatus(key)
	if status != vars.Cacheable {
		t.Fatalf("the %s should be cacheable", key)
	}
	Expire(key)
	status, _ = GetRequestStatus(key)
	if status != vars.Fetching {
		t.Fatalf("the %s should be fetching", key)
	}
}

func TestResponseCache(t *testing.T) {
	// 测试生成插入多条记录，将对过期数据删除
	startedAt := time.Now()
	count := 10 * 1024
	for index := 0; index < count; index++ {
		key := []byte("test-" + strconv.Itoa(index))
		SaveResponseData(key, &ResponseData{
			CreatedAt:  uint32(time.Now().Unix()),
			StatusCode: 200,
			TTL:        10,
			Header:     make([]byte, 3*1024),
			Body:       make([]byte, 50*1024),
		})
	}
	log.Printf("create %v use %v", count, time.Since(startedAt))
	ClearExpired()
	Close()
}

func TestErrorCase(t *testing.T) {
	lsm, vLog := DataSize()
	if lsm != -1 || vLog != -1 {
		t.Fatalf("when client close, database size should be -1")
	}
	_, err := Get([]byte(""))
	if err != vars.ErrDbNotInit {
		t.Fatalf("get function should throw error when database is not init")
	}

	err = Save([]byte(""), []byte(""), 1)
	if err != vars.ErrDbNotInit {
		t.Fatalf("save function should throw error when database is not init")
	}

	err = ClearExpired()
	if err != vars.ErrDbNotInit {
		t.Fatalf("clear expired function should throw error when database is not init")
	}

	err = Close()
	if err != vars.ErrDbNotInit {
		t.Fatalf("close function should throw error when database is not init")
	}
}
