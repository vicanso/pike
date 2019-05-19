package upstream

import (
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/vicanso/cod"

	gock "gopkg.in/h2non/gock.v1"
)

func TestUpstreams(t *testing.T) {
	us := make(Upstreams, 0)
	us = append(us, &Upstream{
		Priority: 9,
	})

	us = append(us, &Upstream{
		Priority: 1,
	})
	sort.Sort(us)
	if us[0].Priority != 1 {
		t.Fatalf("sort upstream fail")
	}
}

func TestHash(t *testing.T) {
	if hash("abc") != 440920331 {
		t.Fatalf("hash fail")
	}
}

func TestNewDirector(t *testing.T) {
	d := new(Director)
	d.Fetch()
	d.StartHealthCheck()
	d.ClearUpstreams()
}

func TestCreateTargetPicker(t *testing.T) {
	backends := []string{
		"http://127.0.0.1:7001",
		"http://127.0.0.1:7002",
		"http://127.0.0.1:7003",
	}
	create := func() *Upstream {
		us := createUpstreamFromBackend(Backend{
			Backends: backends,
		})
		for _, item := range us.Server.GetUpstreamList() {
			item.Healthy()
		}
		return us
	}

	t.Run("convert http header", func(t *testing.T) {
		us := createUpstreamFromBackend(Backend{
			Backends: []string{
				"http://127.0.0.1:7001|backup",
				"http://127.0.0.1:7002",
				"http://127.0.0.1:7003",
			},
			Header: []string{
				"X-Server:test",
			},
			RequestHeader: []string{
				"X-Request-ID:123",
			},
		})
		if us.Header.Get("X-Server") != "test" ||
			us.RequestHeader.Get("X-Request-ID") != "123" {
			t.Fatalf("set header fail")
		}
	})

	t.Run("first", func(t *testing.T) {
		us := create()
		us.Policy = policyFirst
		fn := createTargetPicker(us)
		for i := 0; i < 100; i++ {
			info, err := fn(nil)
			if err != nil {
				t.Fatalf("first policy fail, %v", err)
			}
			if info.String() != backends[0] {
				t.Fatalf("first policy should be always get the first backend")
			}
		}
	})

	t.Run("first with backup", func(t *testing.T) {
		us := createUpstreamFromBackend(Backend{
			Backends: []string{
				"http://127.0.0.1:7001|backup",
				"http://127.0.0.1:7002",
			},
		})
		for _, item := range us.Server.GetUpstreamList() {
			item.Healthy()
		}
		us.Policy = policyFirst
		fn := createTargetPicker(us)
		for i := 0; i < 100; i++ {
			info, err := fn(nil)
			if err != nil {
				t.Fatalf("first policy fail, %v", err)
			}
			if info.String() != "http://127.0.0.1:7002" {
				t.Fatalf("first policy(with backup and all healthy) should be always get the first backend")
			}
		}
	})

	t.Run("first with backup", func(t *testing.T) {
		us := createUpstreamFromBackend(Backend{
			Backends: []string{
				"http://127.0.0.1:7001|backup",
				"http://127.0.0.1:7002",
			},
		})
		us.Server.GetUpstreamList()[0].Healthy()
		us.Policy = policyFirst
		fn := createTargetPicker(us)
		for i := 0; i < 100; i++ {
			info, err := fn(nil)
			if err != nil {
				t.Fatalf("first policy fail, %v", err)
			}
			if info.String() != "http://127.0.0.1:7001" {
				t.Fatalf("first policy(with backup and only backup healthy) should be always get the first backend")
			}
		}
	})

	t.Run("random", func(t *testing.T) {
		us := create()
		us.Policy = policyRandom
		firstUP := us.Server.GetUpstreamList()[0]
		firstUP.Sick()
		defer firstUP.Healthy()
		fn := createTargetPicker(us)
		for i := 0; i < 100; i++ {
			info, err := fn(nil)
			if err != nil || info == nil || info.String() == backends[0] {
				t.Fatalf("random policy fail, %v", err)
			}
		}
	})

	t.Run("round robin", func(t *testing.T) {
		us := create()
		us.Policy = policyRoundRobin
		fn := createTargetPicker(us)
		for i := 0; i < 100; i++ {
			info, err := fn(nil)
			index := (i + 1) % len(backends)
			if err != nil || info == nil || info.String() != backends[index] {
				t.Fatalf("round robin policy fail, %v", err)
			}
		}
	})

	t.Run("least conn", func(t *testing.T) {
		us := create()
		us.Policy = policyLeastconn
		fn := createTargetPicker(us)
		c := cod.NewContext(nil, nil)
		for i := 0; i < 100; i++ {
			info, err := fn(c)
			index := i % len(backends)
			if err != nil || info == nil || info.String() != backends[index] {
				t.Fatalf("least conn policy fail, %v", err)
			}
		}
	})

	t.Run("ip hash", func(t *testing.T) {
		us := create()
		us.Policy = policyIPHash
		fn := createTargetPicker(us)
		req := httptest.NewRequest("GET", "/", nil)
		c := cod.NewContext(nil, req)
		c.SetRequestHeader(cod.HeaderXForwardedFor, "1.1.1.1")
		index := hash(c.RealIP()) % uint32(len(backends))
		for i := 0; i < 100; i++ {
			info, err := fn(c)
			if err != nil || info == nil || info.String() != backends[index] {
				t.Fatalf("ip hash policy fail, %v", err)
			}
		}
	})

	t.Run("header field", func(t *testing.T) {
		us := create()
		key := "X-Request-ID"
		us.Policy = headerHashPrefix + key
		fn := createTargetPicker(us)
		req := httptest.NewRequest("GET", "/", nil)
		c := cod.NewContext(nil, req)
		c.SetRequestHeader(key, "123")
		index := hash(c.GetRequestHeader(key)) % uint32(len(backends))
		for i := 0; i < 100; i++ {
			info, err := fn(c)
			if err != nil || info == nil || info.String() != backends[index] {
				t.Fatalf("header field policy fail, %v", err)
			}
		}
	})

	t.Run("cookie field", func(t *testing.T) {
		us := create()
		key := "jt"
		us.Policy = cookieHashPrefix + key
		fn := createTargetPicker(us)
		req := httptest.NewRequest("GET", "/", nil)
		c := cod.NewContext(nil, req)
		c.SetRequestHeader("Cookie", key+"="+"123")
		cookie, err := c.Cookie(key)
		if err != nil || cookie == nil || cookie.Value == "" {
			t.Fatalf("get cookie fail, %v", err)
		}
		index := hash(cookie.Value) % uint32(len(backends))
		for i := 0; i < 100; i++ {
			info, err := fn(c)
			if err != nil || info == nil || info.String() != backends[index] {
				t.Fatalf("cookie field policy fail, %v", err)
			}
		}
	})

	t.Run("no available upstream", func(t *testing.T) {
		us := create()
		for _, item := range us.Server.GetUpstreamList() {
			item.Sick()
		}
		fn := createTargetPicker(us)
		req := httptest.NewRequest("GET", "/", nil)
		c := cod.NewContext(nil, req)
		_, err := fn(c)
		if err != errNoAvailableUpstream {
			t.Fatalf("should return no available upstream")
		}
	})

}

func TestCreateProxyHandler(t *testing.T) {
	defer gock.Off()
	gock.New("http://127.0.0.1:7001").
		Get("/").
		Reply(200)

	us := createUpstreamFromBackend(Backend{
		Policy: policyLeastconn,
		Ping:   "/ping",
		Hosts: []string{
			"aslant.site",
		},
		Prefixs: []string{
			"/api",
		},
		Rewrites: []string{
			"/api/*:/$1",
		},
		Backends: []string{
			"http://127.0.0.1:7001",
		},
	})
	for _, item := range us.Server.GetUpstreamList() {
		item.Healthy()
	}
	fn := createProxyHandler(us, nil)
	req := httptest.NewRequest("GET", "http://aslant.site/", nil)
	resp := httptest.NewRecorder()
	c := cod.NewContext(resp, req)
	c.Next = func() error {
		return nil
	}
	err := fn(c)
	if err != nil {
		t.Fatalf("proxy fail, %v", err)
	}
	if c.StatusCode != 200 {
		t.Fatalf("proxy response invalid")
	}
}

func TestUpstream(t *testing.T) {
	up := Upstream{
		Hosts: []string{
			"aslant.site",
		},
		Prefixs: []string{
			"/api",
		},
	}

	req := httptest.NewRequest("GET", "http://aslant.site/api/users/me", nil)
	c := cod.NewContext(nil, req)
	if !up.Match(c) {
		t.Fatalf("should match upstream")
	}

	req = httptest.NewRequest("GET", "http://aslant.site/", nil)
	c = cod.NewContext(nil, req)
	if up.Match(c) {
		t.Fatalf("should not match upstream, prefix is not match")
	}

	req = httptest.NewRequest("GET", "http://127.0.0.1/api/users/me", nil)
	c = cod.NewContext(nil, req)
	if up.Match(c) {
		t.Fatalf("should not match upstream, host is not match")
	}
}

func TestProxy(t *testing.T) {
	upstreams := make(Upstreams, 0)
	us := New(Backend{
		Backends: []string{
			"http://127.0.0.1:7001",
		},
		Hosts: []string{
			"aslant.site",
		},
		Header: []string{
			"X-Server:test",
		},
		RequestHeader: []string{
			"X-Request-ID:123",
		},
	}, nil)
	for _, item := range us.Server.GetUpstreamList() {
		item.Healthy()
	}
	upstreams = append(upstreams, us)

	d := Director{
		Upstreams: upstreams,
	}

	req := httptest.NewRequest("GET", "/", nil)
	resp := httptest.NewRecorder()
	c := cod.NewContext(resp, req)
	err := d.Proxy(c)
	if err != errNoMatchUpstream {
		t.Fatalf("should return no match upstream")
	}

	defer gock.Off()
	gock.New("http://127.0.0.1:7001").
		Get("/").
		Reply(200).
		JSON(map[string]string{
			"foo": "bar",
		})

	req = httptest.NewRequest("GET", "http://aslant.site/", nil)
	resp = httptest.NewRecorder()
	c = cod.NewContext(resp, req)
	c.Next = func() error {
		return nil
	}
	err = d.Proxy(c)
	if err != nil ||
		c.StatusCode != http.StatusOK ||
		strings.TrimSpace(c.BodyBuffer.String()) != `{"foo":"bar"}` {
		t.Fatalf("proxy fail, %v", err)
	}
}

func TestGetDirectorStats(t *testing.T) {
	us := createUpstreamFromBackend(Backend{
		Backends: []string{
			"http://127.0.0.1:7001|backup",
			"http://127.0.0.1:7002",
		},
	})
	for _, item := range us.Server.GetUpstreamList() {
		item.Healthy()
	}
	usList := make(Upstreams, 0)
	usList = append(usList, us)
	director := Director{
		Upstreams: usList,
	}
	infoList := director.GetUpstreamInfos()
	if len(infoList) == 0 {
		t.Fatalf("get update stream in")
	}
}
