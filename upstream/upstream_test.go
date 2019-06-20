package upstream

import (
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vicanso/cod"
	"github.com/vicanso/pike/config"

	gock "gopkg.in/h2non/gock.v1"
)

func TestUpstreams(t *testing.T) {
	assert := assert.New(t)

	us := make(Upstreams, 0)
	us = append(us, &Upstream{
		Priority: 9,
	})

	us = append(us, &Upstream{
		Priority: 1,
	})
	sort.Sort(us)
	assert.Equal(1, us[0].Priority, "upstream should be sorted by priority")
}

func TestHash(t *testing.T) {
	assert.Equal(t, uint32(440920331), hash("abc"))
}

func TestNewDirector(t *testing.T) {
	d := new(Director)
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
		us := createUpstreamFromBackend(config.BackendConfig{
			Backends: backends,
		})
		for _, item := range us.Server.GetUpstreamList() {
			item.Healthy()
		}
		return us
	}

	t.Run("convert http header", func(t *testing.T) {
		assert := assert.New(t)
		us := createUpstreamFromBackend(config.BackendConfig{
			Backends: []string{
				"http://127.0.0.1:7001|backup",
				"http://127.0.0.1:7002",
				"http://127.0.0.1:7003",
			},
			ResponseHeader: []string{
				"X-Server:test",
			},
			RequestHeader: []string{
				"X-Request-ID:123",
			},
		})
		assert.Equal("test", us.Header.Get("X-Server"))
		assert.Equal("123", us.RequestHeader.Get("X-Request-ID"))
	})

	t.Run("first", func(t *testing.T) {
		assert := assert.New(t)
		us := create()
		us.Policy = policyFirst
		fn := createTargetPicker(us)
		for i := 0; i < 100; i++ {
			info, _, err := fn(nil)
			assert.Nil(err, "first policy fail")
			assert.Equal(backends[0], info.String(), "first policy should always get the first backend")
		}
	})

	t.Run("first with backup", func(t *testing.T) {
		assert := assert.New(t)
		us := createUpstreamFromBackend(config.BackendConfig{
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
			info, _, err := fn(nil)
			assert.Nil(err, "first policy fail")
			assert.Equal("http://127.0.0.1:7002", info.String(), "first policy(with backup and all healthy) should always get the first backend")
		}
	})

	t.Run("first with backup", func(t *testing.T) {
		assert := assert.New(t)
		us := createUpstreamFromBackend(config.BackendConfig{
			Backends: []string{
				"http://127.0.0.1:7001|backup",
				"http://127.0.0.1:7002",
			},
		})
		us.Server.GetUpstreamList()[0].Healthy()
		us.Policy = policyFirst
		fn := createTargetPicker(us)
		for i := 0; i < 100; i++ {
			info, _, err := fn(nil)
			assert.Nil(err, "first policy fail")
			assert.Equal("http://127.0.0.1:7001", info.String(), "first policy(with backup and only backup healthy) should always get the first backend")
		}
	})

	t.Run("random", func(t *testing.T) {
		assert := assert.New(t)
		us := create()
		us.Policy = policyRandom
		firstUP := us.Server.GetUpstreamList()[0]
		firstUP.Sick()
		defer firstUP.Healthy()
		fn := createTargetPicker(us)
		for i := 0; i < 100; i++ {
			info, _, err := fn(nil)
			assert.Nil(err, "random policy fail")
			assert.NotEqual(backends[0], info.String(), "sick backend shouldn't be got be random policy")
		}
	})

	t.Run("round robin", func(t *testing.T) {
		assert := assert.New(t)
		us := create()
		us.Policy = policyRoundRobin
		fn := createTargetPicker(us)
		for i := 0; i < 100; i++ {
			info, _, err := fn(nil)
			index := (i + 1) % len(backends)
			assert.Nil(err, "round robin fail")
			assert.Equal(backends[index], info.String(), "round robin should get backend round")
		}
	})

	t.Run("least conn", func(t *testing.T) {
		assert := assert.New(t)
		us := create()
		us.Policy = policyLeastconn
		fn := createTargetPicker(us)
		c := cod.NewContext(nil, nil)
		for i := 0; i < 100; i++ {
			info, _, err := fn(c)
			index := i % len(backends)
			assert.Nil(err, "least conn policy fail")
			assert.Equal(backends[index], info.String(), "least conn should always get least conn backend")
		}
	})

	t.Run("ip hash", func(t *testing.T) {
		assert := assert.New(t)
		us := create()
		us.Policy = policyIPHash
		fn := createTargetPicker(us)
		req := httptest.NewRequest("GET", "/", nil)
		c := cod.NewContext(nil, req)
		c.SetRequestHeader(cod.HeaderXForwardedFor, "1.1.1.1")
		index := hash(c.RealIP()) % uint32(len(backends))
		for i := 0; i < 100; i++ {
			info, _, err := fn(c)
			assert.Nil(err, "ip hash policy fail")
			assert.Equal(backends[index], info.String(), "ip hash should get backend by hash ip")
		}
	})

	t.Run("header field", func(t *testing.T) {
		assert := assert.New(t)
		us := create()
		key := "X-Request-ID"
		us.Policy = headerHashPrefix + key
		fn := createTargetPicker(us)
		req := httptest.NewRequest("GET", "/", nil)
		c := cod.NewContext(nil, req)
		c.SetRequestHeader(key, "123")
		index := hash(c.GetRequestHeader(key)) % uint32(len(backends))
		for i := 0; i < 100; i++ {
			info, _, err := fn(c)
			assert.Nil(err, "header field policy fail")
			assert.Equal(backends[index], info.String(), "header field should get backned by hash header field")
		}
	})

	t.Run("cookie field", func(t *testing.T) {
		assert := assert.New(t)
		us := create()
		key := "jt"
		us.Policy = cookieHashPrefix + key
		fn := createTargetPicker(us)
		req := httptest.NewRequest("GET", "/", nil)
		c := cod.NewContext(nil, req)
		c.SetRequestHeader("Cookie", key+"="+"123")
		cookie, err := c.Cookie(key)
		assert.Nil(err, "get cookie fail")
		assert.Equal(cookie.Value, "123")
		index := hash(cookie.Value) % uint32(len(backends))
		for i := 0; i < 100; i++ {
			info, _, err := fn(c)
			assert.Nil(err)
			assert.Equal(backends[index], info.String(), "cookie policy should get backend by hash cookie value")
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
		_, _, err := fn(c)
		assert.Equal(t, errNoAvailableUpstream, err)
	})

}

func TestCreateProxyHandler(t *testing.T) {
	assert := assert.New(t)
	defer gock.Off()
	gock.New("http://127.0.0.1:7001").
		Get("/").
		Reply(200)

	us := createUpstreamFromBackend(config.BackendConfig{
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
	assert.Nil(err, "proxy fail")
	assert.Equal(200, c.StatusCode)
}

func TestUpstream(t *testing.T) {
	assert := assert.New(t)
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
	assert.True(up.Match(c), "should match the upstream")

	req = httptest.NewRequest("GET", "http://aslant.site/", nil)
	c = cod.NewContext(nil, req)
	assert.False(up.Match(c), "should not match upstream, prefix is not match")

	req = httptest.NewRequest("GET", "http://127.0.0.1/api/users/me", nil)
	c = cod.NewContext(nil, req)
	assert.False(up.Match(c), "should not match upstream, host is not match")
}

func TestProxy(t *testing.T) {
	assert := assert.New(t)
	upstreams := make(Upstreams, 0)
	us := New(config.BackendConfig{
		Backends: []string{
			"http://127.0.0.1:7001",
		},
		Hosts: []string{
			"aslant.site",
		},
		ResponseHeader: []string{
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
	assert.Equal(err, errNoMatchUpstream)

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
	assert.Nil(err)
	assert.Equal(http.StatusOK, c.StatusCode)
	assert.Equal(`{"foo":"bar"}`, strings.TrimSpace(c.BodyBuffer.String()))
}

func TestGetDirectorStats(t *testing.T) {
	us := createUpstreamFromBackend(config.BackendConfig{
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
	assert.NotEqual(t, 0, len(infoList))
}
