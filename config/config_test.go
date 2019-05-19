package config

import (
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vicanso/pike/df"
)

func converHeader(arr []string) string {
	sort.Sort(sort.StringSlice(arr))
	return strings.Join(arr, "\n")
}

func TestConfig(t *testing.T) {
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
