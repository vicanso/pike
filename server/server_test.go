package server

import (
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/director"
	"github.com/vicanso/pike/proxy"
	"github.com/vicanso/pike/util"
	"github.com/vicanso/pike/vars"

	"github.com/valyala/fasthttp"
)

func testPass(t *testing.T, uri, method string, resultExpected bool) {
	passList := [][]byte{
		[]byte("cache-control=no-cache"),
	}
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI(uri)
	ctx.Request.Header.SetMethod(method)
	result := isPass(ctx, passList)
	if result != resultExpected {
		t.Fatalf("unexpected result in Pass %q %q: %v. Expecting %v", method, uri, result, resultExpected)
	}
}

func TestTrimHeader(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	header := &ctx.Request.Header
	header.SetCanonical([]byte("Content-Type"), []byte("application/json; charset=utf-8"))
	header.SetCanonical([]byte("X-Response-Id"), []byte("BJJRAyf4f"))
	header.SetCanonical([]byte("Cache-Control"), []byte("no-cache, max-age=0"))
	header.SetCanonical([]byte("Connection"), []byte("keep-alive"))
	header.SetCanonical([]byte("Date"), []byte("Tue, 09 Jan 2018 12:27:02 GMT"))
	str := "User-Agent: fasthttp\r\nContent-Type: application/json; charset=utf-8\r\nX-Response-Id: BJJRAyf4f\r\nCache-Control: no-cache, max-age=0"
	data := string(trimHeader(header.Header()))
	if data != str {
		t.Fatalf("trim header fail expect %v but %v", str, data)
	}
}
func TestGetResponseHeader(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	data := []byte("hello world")
	ctx.Response.SetBody(data)
	ctx.Response.Header.SetContentLength(len(data))
	ctx.Response.Header.SetCanonical(vars.CacheControl, []byte("public, max-age=30"))

	header := getResponseHeader(&ctx.Response)
	if len(header) != 94 {
		t.Fatalf("get the header from response fail")
	}
}

func TestGetResponseBody(t *testing.T) {
	helloWorld := "hello world"
	data, _ := util.Gzip([]byte(helloWorld), 0)
	ctx := &fasthttp.RequestCtx{}
	ctx.Response.Header.SetCanonical(vars.ContentEncoding, vars.Gzip)
	ctx.SetBody(data)

	body, err := getResponseBody(&ctx.Response)
	if err != nil {
		t.Fatalf("get the response body fail, %v", err)
	}
	if string(body) != helloWorld {
		t.Fatalf("get the response body fail")
	}
}

func TestPass(t *testing.T) {
	testPass(t, "http://127.0.0.1/", "GET", false)
	testPass(t, "http://127.0.0.1/", "HEAD", false)

	testPass(t, "http://127.0.0.1/?cache-control=no-cache", "GET", true)

	testPass(t, "http://127.0.0.1/i18ns", "POST", true)
}

func testGetCacheAge(t *testing.T, cacheControl []byte, resultExpected int) {
	ctx := &fasthttp.RequestCtx{}
	if cacheControl != nil {
		ctx.Response.Header.SetCanonical(vars.CacheControl, cacheControl)
	}
	result := getCacheAge(&ctx.Response.Header)
	if result != resultExpected {
		t.Fatalf("unexpected result in GetCacheAge %q: %v. Expecting %v", cacheControl, result, resultExpected)
	}
}

func TestGetCacheAge(t *testing.T) {
	testGetCacheAge(t, nil, 0)
	testGetCacheAge(t, []byte("max-age=30"), 30)
	testGetCacheAge(t, []byte("private,max-age=30"), 0)
	testGetCacheAge(t, []byte("no-store"), 0)
	testGetCacheAge(t, []byte("no-cache"), 0)
	testGetCacheAge(t, []byte("max-age=0"), 0)
	testGetCacheAge(t, []byte("s-maxage=10, max-age=30"), 10)
}

func TestGenRequestKey(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("http://127.0.0.1:5018/users/me?a=1")
	key := string(genRequestKey(ctx))
	if key != "GET 127.0.0.1:5018 /users/me?a=1" {
		t.Fatalf("gen request key fail, %q", key)
	}
}

func TestShouldCompress(t *testing.T) {
	ctx := &fasthttp.RequestCtx{}
	ctx.Response.Header.SetContentType("text/css; charset=UTF-8")
	if shouldCompress(&ctx.Response.Header) != true {
		t.Fatalf("the css should be compress")
	}

	ctx.Response.Header.SetContentType("application/javascript; charset=UTF-8")
	if shouldCompress(&ctx.Response.Header) != true {
		t.Fatalf("the js should be compress")
	}

	ctx.Response.Header.SetContentType("image/png")
	if shouldCompress(&ctx.Response.Header) != false {
		t.Fatalf("the image shouldn't be compress")
	}

}

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

	conf := &config.Config{
		Listen: ":3015",
		DB:     "/tmp/pike.db",
		Directors: []*config.Director{
			&config.Director{
				Name: "test",
				Type: "first",
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

	for _, d := range conf.Directors {
		director.Append(&director.Config{
			Name:     d.Name,
			Policy:   d.Type,
			Ping:     d.Ping,
			Pass:     d.Pass,
			Prefix:   d.Prefix,
			Host:     d.Host,
			Backends: d.Backends,
		})
		proxy.AppendUpstream(&proxy.UpstreamConfig{
			Name:     d.Name,
			Policy:   d.Type,
			Ping:     d.Ping,
			Backends: d.Backends,
		})
	}

	go Start(&Config{
		EnableServerTiming: true,
		Listen:             conf.Listen,
	})
	time.Sleep(5 * time.Second)
	testCachable(t, "http://127.0.0.1:3015/cacheable")
	testHitForPass(t, "http://127.0.0.1:3015/hit-for-pass")
}
