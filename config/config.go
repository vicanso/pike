package config

import (
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"

	debug "github.com/visionmedia/go-debug"
)

// Debug debug日志输出
var Debug = debug.Debug("pike.config")

const (
	defaultDB = "/tmp/pike"
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

// Config 程序配置
type Config struct {
	Name                 string
	Listen               string
	DB                   string
	AdminPath            string `yaml:"adminPath"`
	AdminToken           string `yaml:"adminToken"`
	DisableKeepalive     bool   `yaml:"disableKeepalive"`
	Concurrency          int
	HitForPass           int           `yaml:"hitForPass"`
	ETag                 bool          `yaml:"etag"`
	CompressMinLength    int           `yaml:"compressMinLength"`
	ReadBufferSize       int           `yaml:"readBufferSize"`
	WriteBufferSize      int           `yaml:"writeBufferSize"`
	ConnectTimeout       time.Duration `yaml:"connectTimeout"`
	ReadTimeout          time.Duration `yaml:"readTimeout"`
	WriteTimeout         time.Duration `yaml:"writeTimeout"`
	MaxConnsPerIP        int           `yaml:"maxConnsPerIP"`
	MaxKeepaliveDuration time.Duration `yaml:"maxKeepaliveDuration"`
	MaxRequestBodySize   int           `yaml:"maxRequestBodySize"`
	ExpiredClearInterval time.Duration `yaml:"expiredClearInterval"`
	LogFormat            string        `yaml:"logFormat"`
	UDPLog               string        `yaml:"udpLog"`
	AccessLog            string        `yaml:"accessLog"`
	LogType              string        `yaml:"logType"`
	EnableServerTiming   bool          `yaml:"enableServerTiming"`
	TextTypes            []string      `yaml:"textTypes"`
	CertFile             string        `yaml:"certFile"`
	KeyFile              string        `yaml:"keyFile"`
	Directors            []*Director
	Favicon              string   `yaml:"favicon"`
	ResponseHeader       []string `yaml:"responseHeader"`
}

// InitFromFile 从文件中读取配置初始化
func InitFromFile(file string) (*Config, error) {

	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	conf := &Config{}
	err = yaml.Unmarshal(buf, conf)
	if err != nil {
		return nil, err
	}
	if len(conf.DB) == 0 {
		conf.DB = defaultDB
	}
	Debug("conf: %v", conf)
	return conf, nil
}
