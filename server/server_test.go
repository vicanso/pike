package server

import (
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"../cache"
	"../director"

	"github.com/valyala/fasthttp"
)

func get(url string) (*fasthttp.Response, error) {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	resp := fasthttp.AcquireResponse()
	client := &fasthttp.Client{}
	err := client.Do(req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func testCachable(t *testing.T, url string) {
	mutex := sync.Mutex{}
	count := 10
	result := make([]string, 0, count)
	for index := 0; index < count; index++ {
		go func() {
			resp, err := get(url)
			if err != nil {
				t.Fatalf("get cacheable request fail, %v", err)
			}
			data := string(resp.Body())
			mutex.Lock()
			result = append(result, data)
			mutex.Unlock()
		}()
	}
	time.Sleep(3 * time.Second)
	first := result[0]
	for _, item := range result {
		if item != first {
			t.Fatalf("the cacheable request result should be the same")
		}
	}
}

func testHitForPass(t *testing.T, url string) {
	mutex := sync.Mutex{}
	count := 10
	result := make([]string, 0)
	for index := 0; index < count; index++ {
		go func() {
			resp, err := get(url)
			if err != nil {
				t.Fatalf("get hit for pass request fail, %v", err)
			}
			data := string(resp.Body())
			mutex.Lock()
			result = append(result, data)
			mutex.Unlock()
		}()
	}
	time.Sleep(3 * time.Second)
	for index, item := range result {
		for subIndex, subItem := range result {
			if index != subIndex && item == subItem {
				t.Fatalf("the hit for pass request result shouldn't be the same")
			}
		}
	}
}

func TestServerStart(t *testing.T) {
	port := 5000 + rand.New(rand.NewSource(time.Now().UnixNano())).Intn(1000)
	go func() {
		var requestNo int32
		server := &fasthttp.Server{
			Handler: func(ctx *fasthttp.RequestCtx) {
				switch string(ctx.Path()) {
				case "/error":
					ctx.SetStatusCode(500)
					ctx.SetBodyString("fail")
				case "/ping":
					ctx.SetBodyString("pong")
				case "/hit-for-pass":
					v := int(atomic.AddInt32(&requestNo, 1))
					ctx.SetBodyString(strconv.Itoa(v))
				case "/cacheable":
					ctx.Response.Header.SetCanonical([]byte("Cache-Control"), []byte("public, max-age=60"))
					v := int(atomic.AddInt32(&requestNo, 1))
					ctx.SetBodyString(strconv.Itoa(v))
				default:
					ctx.SetStatusCode(404)
				}
			},
		}
		server.ListenAndServe(":" + strconv.Itoa(port))
	}()

	conf := &PikeConfig{
		Listen: ":3015",
		DB:     "/tmp/pike.db",
		Directors: []*director.Config{
			&director.Config{
				Name: "test",
				Type: "first",
				Ping: "/ping",
				Backends: []string{
					"127.0.0.1:" + strconv.Itoa(port),
				},
			},
		},
	}

	_, err := cache.InitDB(conf.DB)
	if err != nil {
		t.Fatalf("init database fail, %v", err)
	}

	directorList := director.GetDirectors(conf.Directors)
	go Start(conf, directorList)
	time.Sleep(5 * time.Second)
	testCachable(t, "http://127.0.0.1:3015/cacheable")
	testHitForPass(t, "http://127.0.0.1:3015/hit-for-pass")
}
