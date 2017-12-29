package proxy

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
)

func TestCreateUpstream(t *testing.T) {
	name := "tiny"
	policy := "roundRoubin"
	us := CreateUpstream(name, policy, "")
	if us.Name != name {
		t.Fatalf("create upstream fail, the name field is wrong")
	}
	if len(us.Hosts) != 0 {
		t.Fatalf("create upstream fail, the hosts should be empty")
	}
	if us.Policy == nil {
		t.Fatalf("create upstream fail, the policy shoud not be nil")
	}
}

func testDo(t *testing.T, us *Upstream, uri, data string, status int) {
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI(uri)
	resp, err := Do(ctx, us)
	if err != nil {
		t.Fatalf("do request fail %v", err)
	}
	respStatusCode := resp.StatusCode()
	if respStatusCode != status {
		t.Fatalf("do request fail, status code expect is %d but %d", status, respStatusCode)
	}
	respData := string(resp.Body())
	if respData != data {
		t.Fatalf("do request fail, response data expect is %q but %q", data, respData)
	}
}

func TestDo(t *testing.T) {
	port := 5000 + rand.New(rand.NewSource(time.Now().UnixNano())).Intn(1000)
	go func() {
		server := &fasthttp.Server{
			Handler: func(ctx *fasthttp.RequestCtx) {
				switch string(ctx.Path()) {
				case "/error":
					ctx.SetStatusCode(500)
					ctx.SetBodyString("fail")
				case "/ping":
					ctx.SetBodyString("pong")
				default:
					ctx.SetStatusCode(404)
				}
			},
		}
		server.ListenAndServe(":" + strconv.Itoa(port))
	}()
	us := CreateUpstream("tiny", "roundRoubin", "")
	us.AddBackend("127.0.0.1:" + strconv.Itoa(port))
	us.Hosts[0].Healthy = 1

	testDo(t, us, "/ping", "pong", 200)

	testDo(t, us, "/error", "fail", 500)

	testDo(t, us, "/404", "", 404)

}
