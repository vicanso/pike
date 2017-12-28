package proxy

import "testing"

func TestUpstreamHost(t *testing.T) {
	uh := &UpstreamHost{
		Host: "mac-air",
		MaxConns: 1,
	}
	if uh.Conns != 0 {
		t.Fatalf("connection is not 0")
	}
	if uh.MaxConns != 1{
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

	uh.Conns = 1
	if uh.Full() != true {
		t.Fatalf("upstream should be full")
	}

	if uh.Available() != false {
		t.Fatalf("upstream should not be available")
	}	
}

func TestUpstream(t *testing.T) {
	us := &Upstream{
		Name: "tiny",
		Policy: &Random{},
	}
	us.AddBackend("127.0.0.1:5018")
	if len(us.Hosts) != 1 {
		t.Fatalf("add backend fail")
	}
}
