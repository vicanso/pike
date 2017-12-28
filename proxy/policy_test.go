package proxy

import (
	"testing"

	"github.com/valyala/fasthttp"
)

const testCount = 10

func getUpstreamHostPool() UpstreamHostPool {
	return UpstreamHostPool{
		&UpstreamHost{
			Host:    "host1",
			Healthy: 1,
		},
		&UpstreamHost{
			Host:    "host2",
			Healthy: 1,
		},
	}
}

func TestRoundRobin(t *testing.T) {
	up := getUpstreamHostPool()
	rb := &RoundRobin{}
	hostCount := len(up)
	for index := 0; index < testCount; index++ {
		uh := rb.Select(up, nil)
		if uh.Healthy != 1 || uh.Host != up[(index+1)%hostCount].Host {
			t.Fatalf("round robin policy fail")
		}
	}
}

func TestRandom(t *testing.T) {
	up := getUpstreamHostPool()
	rd := &Random{}
	for index := 0; index < testCount; index++ {
		uh := rd.Select(up, nil)
		if uh.Healthy != 1 {
			t.Fatalf("random policy fail")
		}
	}
	up[0].Healthy = 0
	for index := 0; index < testCount; index++ {
		uh := rd.Select(up, nil)
		if uh.Healthy != 1 || uh.Host == "host1" {
			t.Fatalf("random policy fail")
		}
	}
}

func TestLeastConn(t *testing.T) {
	up := getUpstreamHostPool()
	lc := &LeastConn{}
	for index := 0; index < testCount; index++ {
		uh := lc.Select(up, nil)
		if uh.Healthy != 1 {
			t.Fatalf("least conn policy fail")
		}
	}

	up[0].Conns = 1
	for index := 0; index < testCount; index++ {
		uh := lc.Select(up, nil)
		if uh.Healthy != 1 || uh.Host != "host2" {
			t.Fatalf("least conn policy fail")
		}
	}
}

func TestURIHash(t *testing.T) {
	up := getUpstreamHostPool()
	uriHash := &URIHash{}
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.SetRequestURI("http://127.0.0.1/users/me")
	for index := 0; index < testCount; index++ {
		uh := uriHash.Select(up, ctx)
		if uh.Healthy != 1 || uh.Host != "host1" {
			t.Fatalf("uri hash policy fail")
		}
	}

	ctx.Request.SetRequestURI("http://127.0.0.1/users")
	for index := 0; index < testCount; index++ {
		uh := uriHash.Select(up, ctx)
		if uh.Healthy != 1 || uh.Host != "host2" {
			t.Fatalf("uri hash policy fail")
		}
	}
}

func TestFirst(t *testing.T) {
	up := getUpstreamHostPool()
	first := &First{}
	for index := 0; index < testCount; index++ {
		uh := first.Select(up, nil)
		if uh.Healthy != 1 || uh.Host != "host1" {
			t.Fatalf("first policy fail")
		}
	}
	up[0].Healthy = 0
	for index := 0; index < testCount; index++ {
		uh := first.Select(up, nil)
		if uh.Healthy != 1 || uh.Host != "host2" {
			t.Fatalf("first policy fail")
		}
	}
}

func TestHeader(t *testing.T) {
	up := getUpstreamHostPool()
	header := &Header{
		Name: "X-Token",
	}
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.Header.SetCanonical([]byte("X-Token"), []byte("AB"))
	for index := 0; index < testCount; index++ {
		uh := header.Select(up, ctx)
		if uh.Healthy != 1 || uh.Host != "host1" {
			t.Fatalf("header policy fail")
		}
	}
	ctx.Request.Header.SetCanonical([]byte("X-Token"), []byte("ABC"))
	for index := 0; index < testCount; index++ {
		uh := header.Select(up, ctx)
		if uh.Healthy != 1 || uh.Host != "host2" {
			t.Fatalf("header policy fail")
		}
	}
}
