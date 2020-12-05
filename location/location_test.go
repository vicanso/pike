package location

import (
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vicanso/pike/config"
)

func TestLocation(t *testing.T) {
	assert := assert.New(t)

	testHost := "test.com"
	testUrl := "/api/users/me"
	tests := []struct {
		match    bool
		priority int
		host     string
		url      string
		l        *Location
	}{

		// 无host与prefix限制
		{
			match:    true,
			priority: 8,
			host:     testHost,
			url:      testUrl,
			l:        &Location{},
		},
		// 有host限制且匹配
		{
			match:    true,
			priority: 6,
			host:     testHost,
			url:      testUrl,
			l: &Location{
				Hosts: []string{
					"test.com",
				},
			},
		},
		// 有host限制且不匹配
		{
			match:    false,
			priority: 6,
			host:     "test1.com",
			url:      testUrl,
			l: &Location{
				Hosts: []string{
					"test.com",
				},
			},
		},
		// 有prefix限制且匹配
		{
			match:    true,
			priority: 4,
			host:     testHost,
			url:      testUrl,
			l: &Location{
				Prefixes: []string{
					"/api",
				},
			},
		},
		// 有prefix限制且不匹配
		{
			match:    false,
			priority: 4,
			host:     testHost,
			url:      testUrl,
			l: &Location{
				Prefixes: []string{
					"/rest",
				},
			},
		},
		// 有host prefix限制且匹配
		{
			match:    true,
			priority: 2,
			host:     testHost,
			url:      testUrl,
			l: &Location{
				Prefixes: []string{
					"/api",
				},
				Hosts: []string{
					"test.com",
				},
			},
		},
		// 有host prefix限制且不匹配
		{
			match:    false,
			priority: 2,
			host:     testHost,
			url:      testUrl,
			l: &Location{
				Prefixes: []string{
					"/rest",
				},
				Hosts: []string{
					"test.com",
				},
			},
		},
	}
	for _, tt := range tests {
		assert.Equal(tt.match, tt.l.Match(tt.host, tt.url))
		assert.Equal(tt.priority, tt.l.getPriority())
	}
}

func TestLocations(t *testing.T) {
	assert := assert.New(t)
	ls := NewLocations(Location{
		Name: "test",
	})

	l := ls.Get("test.com", "/api/users/me", "test")
	assert.Equal("test", l.Name)

	ls.Set([]Location{
		{
			Name: "test1",
			Hosts: []string{
				"test.com",
			},
		},
		{
			Name: "test2",
			Prefixes: []string{
				"/api",
			},
		},
	})

	l = ls.Get("test.com", "/api/users/me", "test1", "test2")
	assert.Equal("test2", l.Name)
}

func TestConvertConfig(t *testing.T) {
	assert := assert.New(t)
	name := "location-test"
	upstream := "upstream-test"
	prefixes := []string{
		"/api",
	}
	rewrites := []string{
		"/api/*:/$1",
	}
	hosts := []string{
		"test.com",
	}
	querystrings := []string{
		"id:1",
	}
	reqID := strconv.Itoa(rand.Int())
	os.Setenv("__reqID", reqID)
	reqHeaders := []string{
		"X-Req-Id:$__reqID",
	}
	respHeaders := []string{
		"X-Resp-Id:2",
	}
	timeout := 60 * time.Second

	configs := []config.LocationConfig{
		{
			Name:         name,
			Upstream:     upstream,
			QueryStrings: querystrings,
			Prefixes:     prefixes,
			Rewrites:     rewrites,
			Hosts:        hosts,
			ReqHeaders:   reqHeaders,
			RespHeaders:  respHeaders,
			ProxyTimeout: "1m",
		},
	}
	opts := convertConfigs(configs)
	assert.Equal(1, len(opts))
	assert.Equal(name, opts[0].Name)
	assert.Equal(upstream, opts[0].Upstream)
	assert.Equal(prefixes, opts[0].Prefixes)
	query := make(url.Values)
	query.Add("id", "1")
	assert.Equal(query, opts[0].Query)
	assert.Equal(hosts, opts[0].Hosts)
	assert.Equal(timeout, opts[0].ProxyTimeout)
	assert.Equal(http.Header{
		"X-Req-Id": []string{
			reqID,
		},
	}, opts[0].RequestHeader)
	assert.Equal(http.Header{
		"X-Resp-Id": []string{
			"2",
		},
	}, opts[0].ResponseHeader)
}
func TestDefaultLocations(t *testing.T) {
	assert := assert.New(t)
	l := Get("test.com", "/api/users/me", "test1", "test2")
	assert.Nil(l)
	Reset([]config.LocationConfig{
		{
			Name: "test1",
			Hosts: []string{
				"test.com",
			},
		},
		{
			Name: "test2",
			Prefixes: []string{
				"/api",
			},
		},
	})
	l = Get("test.com", "/api/users/me", "test1", "test2")
	assert.Equal("test2", l.Name)
}

func TestURLRewrite(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		req      *http.Request
		rewrites []string
		result   string
	}{
		{
			req: httptest.NewRequest("GET", "/api/users/me", nil),
			rewrites: []string{
				"^/api/*:/$1",
			},
			result: "/users/me",
		},
		// 不匹配（因为有^前置)
		{
			req: httptest.NewRequest("GET", "/api/users/me", nil),
			rewrites: []string{
				"^/users/*:/$1",
			},
			result: "/api/users/me",
		},
		// 匹配
		{
			req: httptest.NewRequest("GET", "/api/users/me", nil),
			rewrites: []string{
				"/users/*:/rest/$1",
			},
			result: "/rest/me",
		},
	}
	for _, tt := range tests {
		fn := generateURLRewriter(tt.rewrites)
		fn(tt.req)
		assert.Equal(tt.result, tt.req.URL.Path)
	}
}

func TestAddHeader(t *testing.T) {
	assert := assert.New(t)

	randomHeader := func() http.Header {
		h := make(http.Header)
		for i := 0; i < 2; i++ {
			k := strconv.Itoa(rand.Intn(10))
			for j := 0; j < 2; j++ {
				v := strconv.Itoa(rand.Intn(10))
				h.Add(k, v)
			}
		}
		return h
	}
	tests := []struct {
		requestHeader  http.Header
		responseHeader http.Header
	}{
		{
			requestHeader:  randomHeader(),
			responseHeader: randomHeader(),
		},
	}
	for _, tt := range tests {
		l := Location{
			RequestHeader:  tt.requestHeader,
			ResponseHeader: tt.responseHeader,
		}
		{

			h := make(http.Header)
			l.AddRequestHeader(h)
			assert.Equal(tt.requestHeader, h)
			assert.NotEqual(0, len(h))
		}
		{
			h := make(http.Header)
			l.AddResponseHeader(h)
			assert.Equal(tt.responseHeader, h)
			assert.NotEqual(0, len(h))
		}
	}
}
