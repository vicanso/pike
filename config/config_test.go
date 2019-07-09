package config

import (
	"testing"
	"time"

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
response_header:
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
  hit_for_pass: 30s
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
  response_header:
  - X-Powered-By:koa
  request_header:
  - X-Version:${VERSION}
  backends:
  - http://127.0.0.1:5018
  - http://192.168.31.3:3001|backup
  - http://192.168.31.3:3002
`

type (
	MockReadWriter struct {
		data []byte
	}
)

func (mrw *MockReadWriter) ReadConfig() ([]byte, error) {
	return mrw.data, nil
}

func (mrw *MockReadWriter) WriteConfig(data []byte) error {
	mrw.data = data
	return nil
}

func (mrw *MockReadWriter) Watch(fn func()) {
	time.Sleep(100 * time.Millisecond)
	fn()
}

func (mrw *MockReadWriter) Close() error {
	return nil
}

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
		assert.Equal(`concurrency: 256000
compress:
  min_length: 1024
  filter: text|javascript|json
cache:
  zone: 1024
  size: 1024
  hit_for_pass: 5m0s
timeout:
  idle_conn: 1m30s
  expect_continue: 3s
  response_header: 10s
  connect: 15s
  tls_handshake: 5s
`, string(out))
	})

	t.Run("read config from file", func(t *testing.T) {
		assert := assert.New(t)
		bc, err := NewBasicConfig("")
		assert.Nil(err)
		err = bc.ReadConfig()
		assert.Nil(err)
	})

	t.Run("write config", func(t *testing.T) {
		assert := assert.New(t)
		mrw := new(MockReadWriter)
		conf := Config{
			rw: mrw,
		}
		conf.Data = BasicConfig{
			Concurrency: 100,
		}
		result := "concurrency: 100\n"
		err := conf.WriteConfig()
		assert.Nil(err)
		assert.Equal(result, string(mrw.data))
		data, err := conf.YAML()
		assert.Nil(err)
		assert.Equal(result, string(data))
	})

	t.Run("watch", func(t *testing.T) {
		assert := assert.New(t)
		mrw := new(MockReadWriter)
		conf := Config{
			rw: mrw,
		}
		ch := make(chan bool)
		done := false
		go conf.OnConfigChange(func() {
			done = true
			ch <- true
		})
		<-ch
		assert.True(done)
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
		dc, err := NewDirectorConfig("")
		assert.Nil(err)
		err = dc.ReadConfig()
		assert.Nil(err)
	})

	mrw := new(MockReadWriter)
	name := "aslant"
	t.Run("add backend", func(t *testing.T) {
		assert := assert.New(t)
		conf := DirectorConfig{
			rw: mrw,
		}
		err := conf.AddBackend(BackendConfig{
			Name:     name,
			Backends: []string{"http://aslant.site/"},
		})
		assert.Nil(err)
		err = conf.WriteConfig()
		assert.Nil(err)
		assert.Equal(`aslant:
  backends:
  - http://aslant.site/
`, string(mrw.data))
		data, err := conf.YAML()
		assert.Nil(err)
		assert.Equal(data, mrw.data)
		err = conf.AddBackend(BackendConfig{
			Name: name,
		})
		assert.Equal(err, errBackendExists)
	})

	t.Run("update backend", func(t *testing.T) {
		assert := assert.New(t)
		conf := DirectorConfig{
			rw: mrw,
		}

		err := conf.ReadConfig()
		assert.Nil(err)
		conf.UpdateBackend(BackendConfig{
			Name:     name,
			Backends: []string{"http://127.0.0.1:3015/"},
		})
		err = conf.WriteConfig()
		assert.Nil(err)
		assert.Equal(`aslant:
  backends:
  - http://127.0.0.1:3015/
`, string(mrw.data))

		err = conf.UpdateBackend(BackendConfig{
			Name: "abcd",
		})
		assert.Equal(err, errBackendNotExists)
	})

	t.Run("get backends", func(t *testing.T) {
		assert := assert.New(t)
		conf := DirectorConfig{
			rw: mrw,
		}
		err := conf.ReadConfig()
		assert.Nil(err)
		backends := conf.GetBackends()
		assert.Equal(1, len(backends))
	})

	t.Run("remove backend", func(t *testing.T) {
		assert := assert.New(t)
		conf := DirectorConfig{
			rw: mrw,
		}
		err := conf.ReadConfig()
		assert.Nil(err)
		conf.RemoveBackend(name)
		err = conf.WriteConfig()
		assert.Nil(err)
		assert.Equal("{}\n", string(mrw.data))
	})
}
