package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileConfig(t *testing.T) {
	fc := FileConfig{
		Name: "basic",
		Type: "yml",
	}
	t.Run("get file", func(t *testing.T) {
		assert := assert.New(t)
		file, err := fc.getFile()
		assert.Nil(err)
		assert.Equal("basic.yml", file)
	})

	t.Run("read/write config file", func(t *testing.T) {
		data := []byte("abcd")
		assert := assert.New(t)
		_, err := fc.ReadConfig()
		done := false
		fc.Watch(func() {
			done = true
		})
		assert.True(os.IsNotExist(err))

		file, err := fc.getFile()
		assert.Nil(err)
		defer os.Remove(file)

		err = fc.WriteConfig(data)
		assert.Nil(err)
		assert.True(done)

		buf, err := fc.ReadConfig()
		assert.Nil(err)
		assert.Equal(buf, data)

		assert.Nil(fc.Close())
	})
}
