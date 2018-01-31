package config

import (
	"io/ioutil"
	"log"
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
	TextTypeBytes        [][]byte
	Directors            []*Director
	Favicon              string   `yaml:"favicon"`
	ResponseHeader       []string `yaml:"responseHeader"`
}

// InitFromFile 从文件中读取配置初始化
func InitFromFile(file string) *Config {

	buf, err := ioutil.ReadFile(file)
	if err != nil {
		log.Println("get the config file fail,", err)
	}
	conf := &Config{}
	err = yaml.Unmarshal(buf, conf)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	if len(conf.DB) == 0 {
		conf.DB = defaultDB
	}
	Debug("conf: %v", conf)
	return conf
}
