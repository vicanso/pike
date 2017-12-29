package proxy

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
)

func TestUpstreamHost(t *testing.T) {
	uh := &UpstreamHost{
		Host:     "mac-air",
		MaxConns: 1,
	}
	if uh.Conns != 0 {
		t.Fatalf("connection is not 0")
	}
	if uh.MaxConns != 1 {
		t.Fatalf("max connection is not 1")
	}
	if uh.Fails != 0 {
		t.Fatalf("fail is not 0")
	}
	if uh.Successes != 0 {
		t.Fatalf("success is not 0")
	}
	if uh.Healthy != 0 {
		t.Fatalf("healthy is not 0")
	}

	uh.Healthy = 1
	if uh.Available() != true {
		t.Fatalf("upstream should be available")
	}

	uh.Disable()
	if uh.Available() == true {
		t.Fatalf("upstream should be diabled")
	}
	uh.Enable()

	uh.Conns = 1
	if uh.Full() != true {
		t.Fatalf("upstream should be full")
	}

	if uh.Available() != false {
		t.Fatalf("upstream should not be available")
	}
}

func TestUpstream(t *testing.T) {
	port := 5000 + rand.New(rand.NewSource(time.Now().UnixNano())).Intn(1000)
	us := &Upstream{
		Name:   "tiny",
		Policy: &Random{},
	}
	uh := us.AddBackend("127.0.0.1:" + strconv.Itoa(port))
	if len(us.Hosts) != 1 {
		t.Fatalf("add backend fail")
	}

	if uh.Available() == true {
		t.Fatalf("upstream should not be available")
	}

	go func() {
		server := &fasthttp.Server{
			Handler: func(ctx *fasthttp.RequestCtx) {
				ctx.SetBodyString("pong")
			},
		}
		server.ListenAndServe(":" + strconv.Itoa(port))
	}()

	us.StartHealthcheck("/ping", 100*time.Millisecond)
	time.Sleep(time.Second)
	if uh.Available() == false {
		t.Fatalf("upstream shold be avaliable")
	}
}
