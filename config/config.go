package config

import (
	"io/ioutil"
	"time"

	"github.com/go-yaml/yaml"
)

// Director 服务器配置列表
type Director struct {
	Name          string
	Policy        string
	Ping          string
	RequestHeader []string `yaml:"requestHeader"`
	Header        []string
	Prefixs       []string
	Hosts         []string
	Backends      []string
	Rewrites      []string
}

// Config 应用配置
type Config struct {
	Name                 string        `yaml:"name"`
	Listen               string        `yaml:"listen"`
	DB                   string        `yaml:"db"`
	Identity             string        `yaml:"identity"`
	ETag                 bool          `yaml:"etag"`
	Header               []string      `yaml:"header"`
	RequestHeader        []string      `yaml:"requestHeader"`
	EnableServerTiming   bool          `yaml:"enableServerTiming"`
	CompressMinLength    int           `yaml:"compressMinLength"`
	CompressLevel        int           `yaml:"compressLevel"`
	Concurrency          int           `yaml:"concurrency"`
	Directors            []*Director   `yaml:"directors"`
	TextTypes            []string      `yaml:"textTypes"`
	Rewrites             []string      `yaml:"rewrites"`
	ExpiredClearInterval time.Duration `yaml:"expiredClearInterval"`
	ConnectTimeout       time.Duration `yaml:"connectTimeout"`
	LogFormat            string        `yaml:"logFormat"`
	AccessLog            string        `yaml:"accessLog"`
	LogType              string        `yaml:"logType"`
	AdminPath            string        `yaml:"adminPath"`
	AdminToken           string        `yaml:"adminToken"`
}

// InitFromFile 获取默认的配置
func InitFromFile(file string) (c *Config, err error) {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	c = &Config{}
	err = yaml.Unmarshal(buf, c)
	return
}
