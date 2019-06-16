package config

import (
	"testing"

	"github.com/go-yaml/yaml"
	"github.com/stretchr/testify/assert"
)

var basicConfigExample = `admin:
  prefix: /pike
  user: pike
  password: abcd
concurrency: 256000
enable_server_timing: true
identity: host method
header:
- X-Server:$SERVER
- X-Location:GZ
request_header:
- X-Server:$SERVER
- X-Location:GZ
compress:
  level: 1
  min_length: 1024
  filter: text|javascript|json
cache:
  zone: 1024
  size: 10240
  hit_for_pass: 600
timeout:
  idle_conn: 40s
  expect_continue: 1s
  response_header: 10s
  connect: 15s
  tls_handshake: 5s
`

var directorConfigExample = `tiny:
  policy: cookie:jt
  ping: /ping
  prefixs:
  - /api
  rewrites:
  - /api/*:/$1
  hosts:
  - tiny.aslant.site
  header:
  - X-Powered-By:koa
  request_header:
  - X-Version:${VERSION}
  backends:
  - http://127.0.0.1:5018
  - http://192.168.31.3:3001|backup
  - http://192.168.31.3:3002
`

func TestBasicConfig(t *testing.T) {
	t.Run("marshal", func(t *testing.T) {
		assert := assert.New(t)
		c := &BasicConfig{}
		err := yaml.Unmarshal([]byte(basicConfigExample), c)
		assert.Nil(err)

		out, _ := yaml.Marshal(c)
		assert.Equal(basicConfigExample, string(out))
	})

	t.Run("fill default", func(t *testing.T) {
		assert := assert.New(t)
		bc := Config{}
		bc.fillDefault()
		out, err := yaml.Marshal(bc.Data)
		assert.Nil(err)
		assert.Equal(string(out), `concurrency: 256000
compress:
  min_length: 1024
  filter: text|javascript|json
cache:
  zone: 1024
  size: 1024
  hit_for_pass: 300
timeout:
  idle_conn: 1m30s
  expect_continue: 3s
  response_header: 10s
  connect: 1m30s
  tls_handshake: 5s
`)
	})

	t.Run("read config from file", func(t *testing.T) {
		assert := assert.New(t)
		bc := NewFileConfig()
		err := bc.ReadConfig()
		assert.Nil(err)
	})
}

func TestDirectorConfig(t *testing.T) {
	t.Run("marshal", func(t *testing.T) {
		assert := assert.New(t)
		dc := new(BackendConfigs)
		err := yaml.Unmarshal([]byte(directorConfigExample), dc)
		assert.Nil(err)

		out, _ := yaml.Marshal(dc)
		assert.Equal(directorConfigExample, string(out))
	})

	t.Run("read config from file", func(t *testing.T) {
		assert := assert.New(t)
		dc := NewFileDirectorConfig()
		err := dc.ReadConfig()
		assert.Nil(err)
	})
}
