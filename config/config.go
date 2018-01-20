package config

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v2"

	debug "github.com/tj/go-debug"
)

// Debug debug日志输出
var Debug = debug.Debug("pike.config")

// Current 当前配置
var Current = &Config{}

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
	Cpus                 int
	Listen               string
	DB                   string
	AdminPath            string `yaml:"adminPath"`
	AdminToken           string `yaml:"adminToken"`
	DisableKeepalive     bool   `yaml:"disableKeepalive"`
	Concurrency          int
	HitForPass           int           `yaml:"hitForPass"`
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
	Directors            []*Director
}

func init() {
	// 优先从ENV中获取配置文件路径
	config := os.Getenv("PIKE_CONFIG")
	// 如果ENV中没有配置，则从启动命令获取
	if len(config) == 0 {
		flag.StringVar(&config, "c", "/etc/pike/config.yml", "the config file")
		flag.Parse()
	}
	buf, err := ioutil.ReadFile(config)
	if err != nil {
		log.Println("get the config file fail,", err)
	}

	err = yaml.Unmarshal(buf, Current)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	if len(Current.Listen) == 0 {
		Current.Listen = ":3015"
	}
	if len(Current.DB) == 0 {
		Current.DB = "/tmp/pike"
	}
	Debug("conf: %v", Current)
}
