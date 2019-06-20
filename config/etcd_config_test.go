package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEtcdConfig(t *testing.T) {
	now := []byte(time.Now().String())
	etcdURI := "etcd://127.0.0.1:2379/test"
	configName := "abcd"
	t.Run("parse etcd config", func(t *testing.T) {
		assert := assert.New(t)

		conf, path, err := parseEtcdConfig("etcd://127.0.0.1:2379/pike")
		assert.Nil(err)
		assert.Equal([]string{"127.0.0.1:2379"}, conf.Endpoints)
		assert.Equal("/pike", path)

		conf, _, err = parseEtcdConfig("etcd://user:pass@127.0.0.1:2379,127.0.0.1:12379/pike")
		assert.Nil(err)
		assert.Equal([]string{"127.0.0.1:2379", "127.0.0.1:12379"}, conf.Endpoints)
		assert.Equal("user", conf.Username)
		assert.Equal("pass", conf.Password)
	})

	t.Run("new config", func(t *testing.T) {
		assert := assert.New(t)
		etcdConfig, err := NewEtcdConfig("etcd://127.0.0.1:2379/pike")
		assert.Nil(err)
		defer etcdConfig.Close()
		assert.Equal("/pike", etcdConfig.path)
	})

	t.Run("write config", func(t *testing.T) {
		assert := assert.New(t)
		etcdConfig, err := NewEtcdConfig(etcdURI)
		assert.Nil(err)
		defer etcdConfig.Close()
		etcdConfig.Name = configName
		err = etcdConfig.WriteConfig(now)
		assert.Nil(err)
	})

	t.Run("get config", func(t *testing.T) {
		assert := assert.New(t)
		etcdConfig, err := NewEtcdConfig(etcdURI)
		assert.Nil(err)
		defer etcdConfig.Close()
		etcdConfig.Name = configName
		data, err := etcdConfig.ReadConfig()
		assert.Nil(err)
		assert.Equal(now, data)
	})
}
