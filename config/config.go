package config

import (
	"flag"
	"io/ioutil"
	"time"

	"github.com/go-yaml/yaml"
)

// Director 服务器配置列表
type Director struct {
	Name    string
	Policy  string
	Ping    string
	Prefix  []string
	Host    []string
	Backend []string
}

// Config 应用配置
type Config struct {
	Name                 string        `yaml:"name"`
	Listen               string        `yaml:"listen"`
	DB                   string        `yaml:"db"`
	ETag                 bool          `yaml:"etag"`
	CompressMinLength    int           `yaml:"compressMinLength"`
	CompressLevel        int           `yaml:"compressLevel"`
	Directors            []*Director   `yaml:"directors"`
	TextTypes            []string      `yaml:"textTypes"`
	ExpiredClearInterval time.Duration `yaml:"expiredClearInterval"`
	ConnectTimeout       time.Duration `yaml:"connectTimeout"`
}

var defaultConfig = &Config{}

func init() {
	var file string
	flag.StringVar(&file, "c", "./config.yml", "the config file")

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
