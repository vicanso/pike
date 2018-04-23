package proxy

import (
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	"github.com/labstack/echo"

	"github.com/h2non/gock"
	"github.com/vicanso/dash"
)

func TestDirector(t *testing.T) {
	t.Run("create director", func(t *testing.T) {
		d := &Director{
			Name:   "test",
			Policy: "random",
			Ping:   "/ping",
			Backends: []string{
				"http://127.0.0.1:5001",
			},
		}
		backend := "http://127.0.0.1:5002"
		d.AddBackend(backend)
		if !dash.IncludesString(d.Backends, backend) {
			t.Fatalf("add backend fail")
		}
		d.RemoveBackend(backend)
		if dash.IncludesString(d.Backends, backend) {
			t.Fatalf("remove backend fail")
		}

	})

	t.Run("match", func(t *testing.T) {
		d := &Director{
			Name:   "test",
			Policy: "random",
			Ping:   "/ping",
			Backends: []string{
				"http://127.0.0.1:5001",
			},
		}
		aslant := "(www.)?aslant.site"
		tiny := "tiny.site"
		d.AddHost(aslant)
		if !d.Match("aslant.site", "/") {
			t.Fatalf("match result should be true")
		}
		d.RemoveHost(aslant)
		if !d.Match(tiny, "/") {
			t.Fatalf("match result should be true")
		}

		d.AddHost(tiny)
		if d.Match("aslant.site", "/") {
			t.Fatalf("match result should be false")
		}

		d.AddHost(aslant)
		d.AddPrefix("/api")
		d.RefreshPriority()
		if d.Priority != 2 {
			t.Fatalf("the director priority should be 2")
		}
		if !d.Match(tiny, "/api/users/me") {
			t.Fatalf("match result should be true")
		}
		d.RemovePrefix("/api")
		d.AddPrefix("/rest")
		if d.Match(tiny, "/api/users/me") {
			t.Fatalf("match result should be false")
		}
	})

	t.Run("directors", func(t *testing.T) {
		ds := make(Directors, 0)
		ds = append(ds, &Director{
			Name:   "tiny",
			Policy: "random",
			Ping:   "/ping",
			Backends: []string{
				"http://127.0.0.1:5001",
			},
		})
		ds = append(ds, &Director{
			Name: "aslant",
			Hosts: []string{
				"aslant.site",
			},
			Policy: "random",
			Ping:   "/ping",
			Backends: []string{
				"http://127.0.0.1:5002",
			},
		})
		for _, d := range ds {
			d.RefreshPriority()
		}
		sort.Sort(ds)
		if ds[0].Name != "aslant" {
			t.Fatalf("the directors sort fail")
		}
	})
}

func TestHealthCheck(t *testing.T) {
	defer gock.Off()

	backends := []string{
		"http://127.0.0.1:5001",
		"http://127.0.0.1:5002",
	}

	// health check 需要测试5次，最少三次成功
	for i := 0; i < 3; i++ {
		for _, backend := range backends {
			gock.New(backend).
				Get("/ping").
				Reply(200).
				BodyString("pong")
		}
	}

	d := &Director{
		Name:     "tiny",
		Policy:   "random",
		Ping:     "/ping",
		Backends: backends,
	}
	d.HealthCheck()
	time.Sleep(5 * time.Millisecond)
	availableBackends := d.GetAvailableBackends()
	if len(availableBackends) != 2 {
		t.Fatalf("the health check fail")
	}
	// 此次测试由于没有mock，因此全部失败
	d.HealthCheck()
	time.Sleep(5 * time.Millisecond)
	availableBackends = d.GetAvailableBackends()
	if len(availableBackends) != 0 {
		t.Fatalf("the health check fail(all backend should be down)")
	}
}

func TestSelect(t *testing.T) {
	backends := []string{
		"http://127.0.0.1:5001",
		"http://127.0.0.1:5002",
	}
	d := &Director{
		Name:     "tiny",
		Policy:   "first",
		Ping:     "/ping",
		Backends: backends,
	}
	d.AddAvailableBackend("http://127.0.0.1:5001")
	d.AddAvailableBackend("http://127.0.0.1:5002")
	e := echo.New()
	for i := 0; i < 10; i++ {
		backend := d.Select(e.NewContext(nil, nil))
		if backend != backends[0] {
			t.Fatalf("first policy fail")
		}
	}

	d.Policy = "roundRobin"
	e = echo.New()
	for i := 0; i < 10; i++ {
		backend := d.Select(e.NewContext(nil, nil))
		if backend != backends[(i+1)%2] {
			t.Fatalf("roundRobin policy fail")
		}
	}

	d.Policy = "random"
	e = echo.New()
	for i := 0; i < 10; i++ {
		backend := d.Select(e.NewContext(nil, nil))
		if backend == "" {
			t.Fatalf("random policy fail")
		}
	}

	d.Policy = "ipHash"
	e = echo.New()
	req := httptest.NewRequest("GET", "/", nil)
	for i := 0; i < 10; i++ {
		backend := d.Select(e.NewContext(req, nil))
		if backend != backends[0] {
			t.Fatalf("ipHash policy fail")
		}
	}

	// custom select function
	cusPolicy := "header:token"
	AddSelectByHeader(cusPolicy, "token")
	d.Policy = cusPolicy
	e = echo.New()
	req = httptest.NewRequest("GET", "/", nil)
	req.Header.Set("token", "ABCD")
	for i := 0; i < 10; i++ {
		backend := d.Select(e.NewContext(req, nil))
		if backend != backends[1] {
			t.Fatalf("custom policy fail")
		}
	}

	d.Policy = "roundRobin"
	d.RemoveAvailableBackend("http://127.0.0.1:5002")
	e = echo.New()
	for i := 0; i < 10; i++ {
		backend := d.Select(e.NewContext(nil, nil))
		if backend != backends[0] {
			t.Fatalf("roundRobin policy fail(one backend avaliable)")
		}
	}
}
