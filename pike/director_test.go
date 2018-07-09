package pike

import (
	"net/http/httptest"
	"net/url"
	"sort"
	"testing"
	"time"

	funk "github.com/thoas/go-funk"

	"github.com/h2non/gock"
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
		if !funk.ContainsString(d.Backends, backend) {
			t.Fatalf("add backend fail")
		}
		d.RemoveBackend(backend)
		if funk.ContainsString(d.Backends, backend) {
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

	t.Run("select", func(t *testing.T) {
		c := NewContext(nil)
		d := &Director{}
		backend := d.Select(c)
		if backend != "" {
			t.Fatalf("should return no backend")
		}
		d.Policy = "AC"
		backend = d.Select(c)
		if backend != "" {
			t.Fatalf("not support policy should return no backend")
		}
	})

	t.Run("get target url", func(t *testing.T) {
		backend := "http://127.0.0.1:5001"
		d := &Director{
			Name:   "tiny",
			Policy: "random",
			Ping:   "/ping",
			Backends: []string{
				backend,
			},
			TargetURLMap: make(map[string]*url.URL),
		}
		targetHost := "127.0.0.1:5001"
		targetURL, err := d.GetTargetURL(&backend)
		if err != nil {
			t.Fatalf("get url from director fail, %v", err)
		}
		if targetURL.Host != targetHost {
			t.Fatalf("the target host should be" + targetHost)
		}
		if targetURL.Scheme != "http" {
			t.Fatalf("the target schema should be http")
		}
		if len(targetURL.RawQuery) != 0 {
			t.Fatalf("the target raw query should be empty")
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
	time.Sleep(100 * time.Millisecond)
	availableBackends := d.GetAvailableBackends()
	if len(availableBackends) != 2 {
		t.Fatalf("the health check fail")
	}
	// 此次测试由于没有mock，因此全部失败
	d.HealthCheck()
	time.Sleep(100 * time.Millisecond)
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
		Rewrites: []string{
			"/api/*:/$1",
		},
	}
	d.AddAvailableBackend("http://127.0.0.1:5001")
	d.AddAvailableBackend("http://127.0.0.1:5002")
	d.GenRewriteRegexp()
	if len(d.RewriteRegexp) != len(d.Rewrites) {
		t.Fatalf("gen rewrite regexp fail")
	}

	for i := 0; i < 10; i++ {
		backend := d.Select(NewContext(nil))
		if backend != backends[0] {
			t.Fatalf("first policy fail")
		}
	}

	d.Policy = "roundRobin"
	for i := 0; i < 10; i++ {
		backend := d.Select(NewContext(nil))
		if backend != backends[(i+1)%2] {
			t.Fatalf("roundRobin policy fail")
		}
	}

	d.Policy = "random"
	for i := 0; i < 10; i++ {
		backend := d.Select(NewContext(nil))
		if backend == "" {
			t.Fatalf("random policy fail")
		}
	}

	d.Policy = "ipHash"
	c := NewContext(httptest.NewRequest("GET", "/", nil))
	for i := 0; i < 10; i++ {
		backend := d.Select(c)
		if backend != backends[0] {
			t.Fatalf("ipHash policy fail")
		}
	}

	// custom select function
	cusPolicy := "header:token"
	AddSelectByHeader(cusPolicy, "token")
	d.Policy = cusPolicy
	req := httptest.NewRequest("GET", "/", nil)
	c = NewContext(req)
	req.Header.Set("token", "ABCD")
	c.Request = req
	for i := 0; i < 10; i++ {
		backend := d.Select(c)
		if backend != backends[1] {
			t.Fatalf("custom policy fail")
		}
	}

	d.Policy = "roundRobin"
	d.RemoveAvailableBackend("http://127.0.0.1:5002")
	for i := 0; i < 10; i++ {
		backend := d.Select(NewContext(nil))
		if backend != backends[0] {
			t.Fatalf("roundRobin policy fail(one backend avaliable)")
		}
	}
}
