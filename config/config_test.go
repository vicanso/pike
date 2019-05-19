package config

import (
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/vicanso/pike/df"
)

func converHeader(arr []string) string {
	sort.Sort(sort.StringSlice(arr))
	return strings.Join(arr, "\n")
}

func TestFetchConfig(t *testing.T) {
	assert := assert.New(t)
	cfg := New()
	err := cfg.Fetch()
	assert.Nil(err)

	assert.Equal(cfg.GetListenAddress(), "127.0.0.1:3015")
	assert.Equal(cfg.GetIdentity(), "host method path proto scheme uri userAgent query ~jt >X-Token ?id")
	assert.Equal(converHeader(cfg.GetHeader()), "X-Location:GZ\nX-Server:$SERVER")
	assert.Equal(converHeader(cfg.GetRequestHeader()), "X-Location:GD\nX-Server:$CLIENT")
	assert.Equal(cfg.GetConcurrency(), uint32(256000))
	assert.True(cfg.GetEnableServerTiming())
	assert.Equal(cfg.GetCacheZoneSize(), 1024)
	assert.Equal(cfg.GetCacheSize(), 10240)
	assert.Equal(cfg.GetHitForPassTTL(), 600)
	assert.Equal(cfg.GetCompressLevel(), 8)
	assert.Equal(cfg.GetCompressMinLength(), 2000)
	assert.Equal(cfg.GetTextFilter(), "text|javascript|json|xml")
	assert.Equal(cfg.GetIdleConnTimeout(), 60*time.Second)
	assert.Equal(cfg.GetExpectContinueTimeout(), 2*time.Second)
	assert.Equal(cfg.GetResponseHeaderTimeout(), 3*time.Second)
	assert.Equal(cfg.GetConnectTimeout(), 4*time.Second)
	assert.Equal(cfg.GetTLSHandshakeTimeout(), 10*time.Second)
	assert.Equal(cfg.GetAdminPath(), "/pike")
	assert.Equal(cfg.GetAdminUser(), "pike")
	assert.Equal(cfg.GetAdminPassword(), "def")
}

func TestSave(t *testing.T) {
	assert := assert.New(t)
	cfg := NewFileConfig("test-config")
	err := cfg.Fetch()
	assert.Nil(err)
	file := df.ConfigPathList[0] + "/" + cfg.Name + "." + defaultConfigType
	defer os.Remove(file)
	cfg.Viper.Set("name", "vicanso")
	err = cfg.WriteConfig()
	assert.Nil(err)
}

func TestDirectorConfig(t *testing.T) {
	assert := assert.New(t)
	cfg := NewFileConfig("director")
	err := cfg.Fetch()
	assert.Nil(err)
	backends := cfg.GetBackends()
	assert.Equal(len(backends), 2)
	assert.Equal(backends[0].Name, "aslant")
}

func TestSetGetConfig(t *testing.T) {
	assert := assert.New(t)
	cfg := &Config{
		Viper: viper.New(),
	}

	listenAddr := ":1234"
	cfg.SetListenAddress(listenAddr)
	assert.Equal(cfg.GetListenAddress(), listenAddr)

	identity := ":method :url"
	cfg.SetIdentity(identity)
	assert.Equal(cfg.GetIdentity(), identity)

	header := []string{
		"X-Server:abc",
	}
	cfg.SetHeader(header)
	assert.Equal(converHeader(cfg.GetHeader()), converHeader(header))

	requestHeader := []string{
		"X-Request-Id:123",
	}
	cfg.SetRequestHeader(requestHeader)
	assert.Equal(converHeader(cfg.GetRequestHeader()), converHeader(requestHeader))

	var concurrency uint32 = 111
	cfg.SetConcurrency(concurrency)
	assert.Equal(cfg.GetConcurrency(), concurrency)

	enableServerTiming := true
	cfg.SetEnableServerTiming(enableServerTiming)
	assert.Equal(cfg.GetEnableServerTiming(), enableServerTiming)

	cacheZone := 132
	cfg.SetCacheZoneSize(cacheZone)
	assert.Equal(cfg.GetCacheZoneSize(), cacheZone)

	cacheSzie := 103
	cfg.SetCacheSzie(cacheSzie)
	assert.Equal(cfg.GetCacheSize(), cacheSzie)

	hitForPassTTL := 198
	cfg.SetHitForPassTTL(hitForPassTTL)
	assert.Equal(cfg.GetHitForPassTTL(), hitForPassTTL)

	compressLevel := 3
	cfg.SetCompressLevel(compressLevel)
	assert.Equal(cfg.GetCompressLevel(), compressLevel)

	compressMinLength := 2034
	cfg.SetCompressMinLength(compressMinLength)
	assert.Equal(cfg.GetCompressMinLength(), compressMinLength)

	textFilter := "a|b|c"
	cfg.SetTextFilter(textFilter)
	assert.Equal(cfg.GetTextFilter(), textFilter)

	idleConnTimeout := time.Millisecond
	cfg.SetIdleConnTimeout(idleConnTimeout)
	assert.Equal(cfg.GetIdleConnTimeout(), idleConnTimeout)

	expectContinueTimeout := 2 * time.Millisecond
	cfg.SetExpectContinueTimeout(expectContinueTimeout)
	assert.Equal(cfg.GetExpectContinueTimeout(), expectContinueTimeout)

	responseHeaderTimeout := 3 * time.Millisecond
	cfg.SetResponseHeaderTimeout(responseHeaderTimeout)
	assert.Equal(cfg.GetResponseHeaderTimeout(), responseHeaderTimeout)

	connectTimeout := 4 * time.Millisecond
	cfg.SetConnectTimeout(connectTimeout)
	assert.Equal(cfg.GetConnectTimeout(), connectTimeout)

	tlsHandshakeTimeout := 5 * time.Millisecond
	cfg.SetTLSHandshakeTimeout(tlsHandshakeTimeout)
	assert.Equal(cfg.GetTLSHandshakeTimeout(), tlsHandshakeTimeout)

	adminPath := "/admin-path"
	cfg.SetAdminPath(adminPath)
	assert.Equal(cfg.GetAdminPath(), adminPath)

	adminUser := "ghost"
	cfg.SetAdminUser(adminUser)
	assert.Equal(cfg.GetAdminUser(), adminUser)

	adminPassword := "pwd"
	cfg.SetAdminPassword(adminPassword)
	assert.Equal(cfg.GetAdminPassword(), adminPassword)
}

func TestSetBackend(t *testing.T) {
	assert := assert.New(t)
	cfg := &Config{
		Viper: viper.New(),
	}
	backend := Backend{
		Name:   "test",
		Policy: "first",
		Ping:   "/ping",
		RequestHeader: []string{
			"X-Request-ID:1",
		},
		Header: []string{
			"X-Server:a",
		},
		Hosts: []string{"aslant.site"},
		Backends: []string{
			"http://127.0.0.1:5001",
		},
		Rewrites: []string{
			"/api/*:/$1",
		},
	}
	cfg.SetBackend(backend)
	backends := cfg.GetBackends()
	assert.Equal(len(backends), 1)
	firstBackend := backends[0]
	assert.Equal(firstBackend.Name, backend.Name)
	assert.Equal(firstBackend.Policy, backend.Policy)
	assert.Equal(firstBackend.RequestHeader, backend.RequestHeader)
	assert.Equal(firstBackend.Header, backend.Header)
	assert.Equal(firstBackend.Prefixs, backend.Prefixs)
	assert.Equal(firstBackend.Hosts, backend.Hosts)
	assert.Equal(firstBackend.Backends, backend.Backends)
	assert.Equal(firstBackend.Rewrites, backend.Rewrites)
}
