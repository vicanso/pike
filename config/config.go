package config

import (
	"flag"
	"io/ioutil"

	"github.com/go-yaml/yaml"
)

// Director 服务器配置列表
type Director struct {
	Name     string
	Type     string
	Ping     string
	Prefix   []string
	Host     []string
	Pass     []string
	Backends []string
}

// Config 应用配置
type Config struct {
	Name      string
	Directors []*Director
}

var defaultConfig = &Config{}

func init() {
	var file string
	flag.StringVar(&file, "c", "/etc/pike/config.yml", "the config file")

	buf, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(buf, defaultConfig)
	if err != nil {
		panic(err)
	}
}

// GetDefault 获取默认的配置
func GetDefault() *Config {
	return defaultConfig
}
