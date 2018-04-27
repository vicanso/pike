package cache

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vicanso/pike/vars"

	"github.com/vicanso/pike/util"
)

const (
	dbPath = "/tmp/test.cache"
)

func TestCacheClient(t *testing.T) {

	t.Run("init", func(t *testing.T) {
		c := Client{
			Path: dbPath,
		}

		err := c.Init()

		if err != nil {
			t.Fatalf("cache init fail, %v", err)
		}
		c.Close()
	})
}
func TestTypeConvert(t *testing.T) {
	t.Run("covert between uint16 and bytes", func(t *testing.T) {
		i := uint16(time.Now().Unix())
		buf := uint16ToBytes(i)
		if bytesToUint16(buf) != i {
			t.Fatalf("covert between uint16 and bytes fail")
		}
	})

	t.Run("covert between uint32 and bytes", func(t *testing.T) {
		i := uint32(time.Now().Unix())
		buf := uint32ToBytes(i)
		if bytesToUint32(buf) != i {
			t.Fatalf("covert between uint32 and bytes fail")
		}
	})
}

func TestIsExpired(t *testing.T) {
	now := uint32(time.Now().Unix())
	ttl := uint16(30)
	rs := &RequestStatus{
		ttl:       ttl,
		createdAt: now,
	}
	if isExpired(rs) {
		t.Fatalf("is expired function fail, it should not expired")
	}
	rs.createdAt = now - uint32(ttl) - 1
	if !isExpired(rs) {
		t.Fatalf("is expired function fail, it should expired")
	}
}

func TestResponse(t *testing.T) {
	c := Client{
		Path: dbPath,
	}
	err := c.Init()
	if err != nil {
		t.Fatalf("cache init fail, %v", err)
	}
	defer c.Close()
	key := []byte("pike.aslant.site /users/me")
	header := make(http.Header)
	header["token"] = []string{
		"A",
	}
	body := []byte("raw body")
	gzipBody, _ := util.Gzip(body, 0)
	brBody, _ := util.Brotli(body, 0)

	now := uint32(time.Now().Unix())

	t.Run("get raw body", func(t *testing.T) {
		r := &Response{
			Body: body,
		}
		rawBody, _ := r.getRawBody()
		if !bytes.Equal(body, rawBody) {
			t.Fatalf("get raw body from body fail")
		}
		r.Body = nil
		r.GzipBody = gzipBody
		rawBody, _ = r.getRawBody()
		if !bytes.Equal(body, rawBody) {
			t.Fatalf("get raw body from gzip body fail")
		}

		r.GzipBody = nil
		r.BrBody = brBody
		rawBody, _ = r.getRawBody()
		if !bytes.Equal(body, rawBody) {
			t.Fatalf("get raw body from br body fail")
		}

		r.BrBody = nil
		_, err = r.getRawBody()
		if err != vars.ErrBodyCotentNotFound {
			t.Fatalf("not found body should return error")
		}
	})

	t.Run("get body", func(t *testing.T) {
		r := &Response{
			StatusCode: http.StatusNoContent,
		}
		// 204的处理
		b, e := r.GetBody("gzip")
		if len(b) != 0 || e != "" {
			t.Fatalf("get no content fail")
		}

		// 200 但是数据比最小压缩长度要小
		r.Body = body
		r.StatusCode = http.StatusOK
		b, e = r.GetBody("gzip")
		if !bytes.Equal(b, body) || e != "" {
			t.Fatalf("get body less then compress min length fail")
		}

		// 数据比最小压缩长度要大，但是客户端不支持压缩
		r.CompressMinLength = 1
		b, e = r.GetBody("")
		if !bytes.Equal(b, body) || e != "" {
			t.Fatalf("get body less then compress min length fail")
		}

		// 数据比最小压缩长度要大，客户端支持gzip，直接读取gzip数据
		r.CompressMinLength = 1
		r.GzipBody = gzipBody
		b, e = r.GetBody("gzip")
		if !bytes.Equal(b, gzipBody) || e != "gzip" {
			t.Fatalf("get gzip body fail")
		}

		// 数据比最小压缩长度要大，客户端支持gzip，
		// 没办法直接读取gzip数据，需要将原数据压缩
		r.GzipBody = nil
		b, e = r.GetBody("gzip")
		if !bytes.Equal(b, gzipBody) || e != "gzip" {
			t.Fatalf("get gzip body fail")
		}

		// 数据比最小压缩长度要大，客户端支持br，直接读取br数据
		r.CompressMinLength = 1
		r.BrBody = brBody
		b, e = r.GetBody("br")
		if !bytes.Equal(b, brBody) || e != "br" {
			t.Fatalf("get br body fail")
		}
		// 数据比最小压缩长度要大，客户端支持br
		// 没办法直接读取br数据，需要将原数据压缩
		r.BrBody = nil
		b, e = r.GetBody("br")
		if !bytes.Equal(b, brBody) || e != "br" {
			t.Fatalf("get br body fail")
		}
	})

	t.Run("save response", func(t *testing.T) {
		resp := &Response{
			CreatedAt:  now,
			StatusCode: 200,
			TTL:        600,
			Header:     header,
			Body:       body,
			GzipBody:   gzipBody,
			BrBody:     brBody,
		}
		err := c.SaveResponse(key, resp)
		if err != nil {
			t.Fatalf("save response fail, %v", err)
		}

		// 如果没有createdAt，则自动创建
		tmpKey := []byte("tmp")
		c.SaveResponse(tmpKey, &Response{})
		resp, _ = c.GetResponse(tmpKey)
		if resp == nil || resp.CreatedAt == 0 {
			t.Fatalf("the response created at should be fill auto")
		}
	})

	t.Run("get response", func(t *testing.T) {
		resp, err := c.GetResponse(key)
		if err != nil {
			t.Fatalf("get response fail, %v", err)
		}
		if resp.CreatedAt != now {
			t.Fatalf("response createat is wrong")
		}

		if resp.StatusCode != 200 {
			t.Fatalf("response status code is wrong")
		}

		if resp.TTL != 600 {
			t.Fatalf("response ttl is wrong")
		}

		buf1, err := json.Marshal(resp.Header)
		if err != nil {
			t.Fatalf("respose header marshal fail")
		}
		buf2, _ := json.Marshal(header)

		if string(buf1) != string(buf2) {
			t.Fatalf("response header is wrong")
		}

		if !bytes.Equal(resp.Body, body) {
			t.Fatalf("response body is wrong")
		}

		if !bytes.Equal(resp.GzipBody, gzipBody) {
			t.Fatalf("response gzip body is wrong")
		}

		if !bytes.Equal(resp.BrBody, brBody) {
			t.Fatalf("response br body is wrong")
		}
	})
}

func TestRequestStatus(t *testing.T) {
	c := Client{
		Path: dbPath,
	}
	err := c.Init()
	if err != nil {
		t.Fatalf("cache init fail, %v", err)
	}
	defer c.Close()
	t.Run("get status", func(t *testing.T) {
		key := []byte("test get status")
		status, ch := c.GetRequestStatus(key)
		done := make(chan int)
		max := 20
		var count uint32
		for index := 0; index < max; index++ {
			// 启用多个goroutine去获取（将是waiting状态）
			go func() {
				s1, c1 := c.GetRequestStatus(key)
				if s1 != Waiting {
					t.Fatalf("the next request should be waiting")
				}
				if c1 == nil {
					t.Fatalf("the chan of next request should be chan int")
				}
				// 并没有等待 chan
				n := atomic.AddUint32(&count, 1)
				if int(n) == max {
					done <- 0
				}
			}()
		}
		<-done
		if status != Fetching {
			t.Fatalf("the first request should be fetching")
		}
		if ch != nil {
			t.Fatalf("the chan of first request should be null")
		}
	})

	t.Run("update status", func(t *testing.T) {
		key := []byte("pike.aslant.site /users/me")
		status, ch := c.GetRequestStatus(key)
		if status != Fetching {
			t.Fatalf("the first request should be fetching")
		}
		if ch != nil {
			t.Fatalf("the chan of first request should be null")
		}
		done := make(chan int)
		chDone := make(chan int)
		max := 20
		var count, statusCount uint32
		for index := 0; index < max; index++ {
			go func() {
				s1, c1 := c.GetRequestStatus(key)
				if s1 != Waiting {
					t.Fatalf("the next request should be waiting")
				}
				if c1 == nil {
					t.Fatalf("the chan of next request should be chan int")
				}

				n := atomic.AddUint32(&count, 1)
				if int(n) == max {
					done <- 0
				}
				// 等待 chan 状态
				v := <-c1
				if v != HitForPass {
					t.Fatalf("the chan should be hitforpass")
				}
				n = atomic.AddUint32(&statusCount, 1)
				if int(n) == max {
					chDone <- 0
				}
			}()
		}
		<-done
		c.HitForPass(key, 300)
		<-chDone
		status, _ = c.GetRequestStatus(key)
		if status != HitForPass {
			t.Fatalf("the status should be hit for pass")
		}
	})

	t.Run("expire", func(t *testing.T) {
		key := []byte("test expire")
		expired := isExpired(&RequestStatus{
			createdAt: 1,
			ttl:       10,
		})
		if !expired {
			t.Fatalf("the status should be expired")
		}
		c.GetRequestStatus(key)
		c.Cacheable(key, 1)
		status, _ := c.GetRequestStatus(key)
		if status != Cacheable {
			t.Fatalf("the reqeust status should be cacheable")
		}
		// 等待两秒后缓存过期
		time.Sleep(2 * time.Second)
		status, _ = c.GetRequestStatus(key)
		if status != Fetching {
			t.Fatalf("the reqeust status should be fetching")
		}

	})
}

func TestClearExpired(t *testing.T) {
	c := Client{
		Path: dbPath,
	}
	err := c.Init()
	if err != nil {
		t.Fatalf("cache init fail, %v", err)
	}
	defer c.Close()
	count := 1000
	for index := 0; index < count; index++ {
		bs := make([]byte, 4)
		binary.LittleEndian.PutUint32(bs, uint32(index))
		c.GetRequestStatus(bs)
	}

	for index := 0; index < count; index++ {
		bs := make([]byte, 4)
		binary.LittleEndian.PutUint32(bs, uint32(index))
		c.UpdateRequestStatus(bs, HitForPass, 1)
	}

	time.Sleep(2 * time.Second)
	c.ClearExpired(0)

	size := c.Size()
	if size != 0 {
		t.Fatalf("all cache shoud be expired")
	}
}

func TestGetStats(t *testing.T) {
	c := Client{
		Path: dbPath,
	}
	err := c.Init()
	if err != nil {
		t.Fatalf("cache init fail, %v", err)
	}
	defer c.Close()
	c.GetRequestStatus([]byte("1"))
	c.GetRequestStatus([]byte("1"))
	c.GetRequestStatus([]byte("1"))

	c.GetRequestStatus([]byte("2"))
	c.HitForPass([]byte("2"), 300)

	c.GetRequestStatus([]byte("3"))
	c.Cacheable([]byte("3"), 300)

	t.Run("get stats", func(t *testing.T) {
		stats := c.GetStats()
		if stats.Fetching != 1 {
			t.Fatalf("feching count fail")
		}
		if stats.Waiting != 2 {
			t.Fatalf("waiting count fail")
		}
		if stats.HitForPass != 1 {
			t.Fatalf("hit for pass count fail")
		}
		if stats.Cacheable != 1 {
			t.Fatalf("cacheable count fail")
		}
	})
}
