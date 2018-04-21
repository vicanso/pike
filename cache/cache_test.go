package cache

import (
	"encoding/binary"
	"sync/atomic"
	"testing"
	"time"
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
	header := "http header"
	body := "raw body"
	gzipBody := "gzip body"
	brBody := "br body"
	now := uint32(time.Now().Unix())

	t.Run("save response", func(t *testing.T) {
		resp := &Response{
			CreatedAt:  now,
			StatusCode: 200,
			TTL:        600,
			Header:     []byte(header),
			Body:       []byte(body),
			GzipBody:   []byte(gzipBody),
			BrBody:     []byte(brBody),
		}
		err := c.SaveResponse(key, resp)
		if err != nil {
			t.Fatalf("save response fail, %v", err)
		}

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

		if string(resp.Header) != header {
			t.Fatalf("response header is wrong")
		}

		if string(resp.Body) != body {
			t.Fatalf("response body is wrong")
		}

		if string(resp.GzipBody) != gzipBody {
			t.Fatalf("response gzip body is wrong")
		}

		if string(resp.BrBody) != brBody {
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
	c.ClearExpired()

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
