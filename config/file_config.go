package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type (
	// FileConfig file config
	FileConfig struct {
		Name     string
		Type     string
		OnChange func()
	}
)

const (
	homeENV = "$HOME"
)

var (
	// configPathList config path list
	configPathList = []string{
		".",
		homeENV + "/.pike",
		"/etc/pike",
	}
)

func (fc *FileConfig) getFile() (file string, err error) {
	files := []string{}
	for _, item := range configPathList {
		name := fc.Name
		if fc.Type != "" {
			name += ("." + fc.Type)
		}
		s := filepath.Join(item, name)
		if strings.HasPrefix(s, homeENV) {
			s = strings.Replace(s, homeENV, os.Getenv(homeENV[1:]), 1)
		}
		files = append(files, s)
	}
	file = files[0]
	for _, item := range files {
		_, err := os.Stat(item)
		if err == nil {
			file = item
			break
		}
	}

	return
}

// ReadConfig read config from file
func (fc *FileConfig) ReadConfig() (buf []byte, err error) {
	file, err := fc.getFile()
	if err != nil {
		return
	}
	return ioutil.ReadFile(file)
}

// WriteConfig write config to file
func (fc *FileConfig) WriteConfig(data []byte) (err error) {
	file, err := fc.getFile()
	if err != nil {
		return
	}
	err = ioutil.WriteFile(file, data, 0600)
	// 只简单在修改时触发，并不监听文件变化
	if err == nil && fc.OnChange != nil {
		fc.OnChange()
	}
	return
}

// Watch watch the config file change
func (fc *FileConfig) Watch(fn func()) {
	fc.OnChange = fn
}
