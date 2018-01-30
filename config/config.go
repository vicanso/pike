package config

import (
	"io/ioutil"
	"log"
	"strings"
	"time"

	"gopkg.in/yaml.v2"

	debug "github.com/visionmedia/go-debug"
)

// Debug debug日志输出
var Debug = debug.Debug("pike.config")

const (
	defaultListen = ":3015"
	defaultDB     = "/tmp/pike"
)

// Current 当前配置
var Current = &Config{
	Listen: defaultListen,
	DB:     defaultDB,
}

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

// Header HTTP response header
type Header struct {
	Key   []byte
	Value []byte
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
	ExtraHeaders         []*Header
}

// InitFromFile 从文件中读取配置初始化
func InitFromFile(file string) {

	buf, err := ioutil.ReadFile(file)
	if err != nil {
		log.Println("get the config file fail,", err)
	}

	err = yaml.Unmarshal(buf, Current)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	if len(Current.Listen) == 0 {
		Current.Listen = defaultListen
	}
	if len(Current.DB) == 0 {
		Current.DB = defaultDB
	}
	if len(Current.ResponseHeader) != 0 {
		for _, str := range Current.ResponseHeader {
			arr := strings.Split(str, ":")
			if len(arr) != 2 {
				continue
			}
			h := &Header{
				Key:   []byte(arr[0]),
				Value: []byte(arr[1]),
			}
			Current.ExtraHeaders = append(Current.ExtraHeaders, h)
		}
	}
	Debug("conf: %v", Current)
}
