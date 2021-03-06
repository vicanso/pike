package upstream

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vicanso/elton"
	"github.com/vicanso/pike/config"
	us "github.com/vicanso/upstream"
	"golang.org/x/net/http2"
)

func TestNewTransport(t *testing.T) {
	assert := assert.New(t)

	transport := newTransport(true)

	h2Transport, ok := transport.(*http2.Transport)
	assert.True(ok)
	assert.True(h2Transport.AllowHTTP)

	transport = newTransport(false)

	hTransport, ok := transport.(*http.Transport)
	assert.True(ok)
	assert.True(hTransport.ForceAttemptHTTP2)
}

func TestNewTargetPicker(t *testing.T) {
	assert := assert.New(t)

	uh := &us.HTTP{
		Policy: us.PolicyLeastconn,
	}
	err := uh.Add("http://127.0.0.1:3000")
	assert.Nil(err)
	for _, up := range uh.GetUpstreamList() {
		up.Healthy()
	}
	fn := newTargetPicker(uh)
	c := elton.NewContext(nil, nil)
	url, done, err := fn(c)
	assert.Nil(err)
	assert.NotNil(done)
	assert.Equal("http://127.0.0.1:3000", url.String())
	done(c)
}

func TestUpstreamServer(t *testing.T) {
	assert := assert.New(t)
	addr := "https://www.bing.com/"
	server := NewUpstreamServer(UpstreamServerOption{
		Policy:      us.PolicyLeastconn,
		Name:        "bing",
		HealthCheck: "/",
		Servers: []UpstreamServerConfig{
			{
				Addr: addr,
			},
			{
				Addr:   "https://bing.com/",
				Backup: true,
			},
		},
		OnStatus: func(info StatusInfo) {
			assert.Equal("healthy", info.Status)
		},
	})
	up, done := server.HTTPUpstream.Next()
	assert.NotNil(up)
	assert.False(up.Backup)
	assert.Equal(addr, up.URL.String())
	assert.NotNil(done)

	statusList := server.GetServerStatusList()
	assert.Equal(len(server.servers), len(statusList))

	server.Destroy()
}

func TestUpstreamServers(t *testing.T) {
	assert := assert.New(t)
	baidu := "baidu"
	servers := NewUpstreamServers([]UpstreamServerOption{
		{
			Name:        baidu,
			HealthCheck: "/",
			Servers: []UpstreamServerConfig{
				{
					Addr: "https://www.baidu.com",
				},
			},
		},
	})
	baiduServer := servers.Get(baidu)
	assert.NotNil(baiduServer)
	assert.Equal(baidu, baiduServer.Option.Name)

	bing := "bing"
	servers.Reset([]UpstreamServerOption{
		{
			Name:        bing,
			HealthCheck: "/",
			Servers: []UpstreamServerConfig{
				{
					Addr: "https://www.bing.com/",
				},
			},
		},
	})
	assert.Nil(servers.Get(baidu))
	bingServer := servers.Get(bing)
	assert.NotNil(bingServer)
	assert.Equal(bing, bingServer.Option.Name)
}

func TestConvertConfig(t *testing.T) {
	assert := assert.New(t)

	name := "upstream-test"
	healthCheck := "/ping"
	policy := "first"
	enableH2C := true
	acceptEncoding := "gzip, br"
	addr := "http://127.0.0.1:3015"
	backup := true

	configs := []config.UpstreamConfig{
		{
			Name:           name,
			HealthCheck:    healthCheck,
			Policy:         policy,
			EnableH2C:      enableH2C,
			AcceptEncoding: acceptEncoding,
			Servers: []config.UpstreamServerConfig{
				{
					Addr:   addr,
					Backup: backup,
				},
			},
		},
	}
	opts := convertConfigs(configs, nil)
	assert.Equal(1, len(opts))
	assert.Equal(name, opts[0].Name)
	assert.Equal(healthCheck, opts[0].HealthCheck)
	assert.Equal(policy, opts[0].Policy)
	assert.Equal(enableH2C, opts[0].EnableH2C)
	assert.Equal(acceptEncoding, opts[0].AcceptEncoding)
	assert.Equal(1, len(opts[0].Servers))
	assert.Equal(addr, opts[0].Servers[0].Addr)
	assert.True(opts[0].Servers[0].Backup)
}

func TestDefaultUpstreamServers(t *testing.T) {
	bing := "bing"
	assert := assert.New(t)
	Reset([]config.UpstreamConfig{
		{
			Name:        bing,
			HealthCheck: "/",
			Servers: []config.UpstreamServerConfig{
				{
					Addr: "https://www.bing.com/",
				},
			},
		},
	})
	bingServer := Get(bing)
	assert.NotNil(bingServer)
	assert.Equal(bing, bingServer.Option.Name)
}
