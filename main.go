package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/director"
	"github.com/vicanso/pike/proxy"
	"github.com/vicanso/pike/server"
	"github.com/vicanso/pike/vars"
)

func clear(interval time.Duration) {
	time.Sleep(interval)
	cache.ClearExpired()
	go clear(interval)
}

func contains(list []string, s string) bool {
	for _, item := range list {
		if s == item {
			return true
		}
	}
	return false
}

func main() {
	if contains(os.Args[1:], "version") {
		fmt.Println("Pike version " + vars.Version)
		return
	}
	// 优先从ENV中获取配置文件路径
	configFile := os.Getenv("PIKE_CONFIG")
	// 如果ENV中没有配置，则从启动命令获取
	if len(configFile) == 0 {
		flag.StringVar(&configFile, "c", "/etc/pike/config.yml", "the config file")
		flag.Parse()
	}
	conf, err := config.InitFromFile(configFile)
	if err != nil {
		log.Fatalf("get config fail, %v", err)
	}
	if contains(os.Args[1:], "test") {
		fmt.Print(conf)
		fmt.Println("the config file test done")
		return
	}
	clearInterval := conf.ExpiredClearInterval
	if clearInterval <= 0 {
		clearInterval = 300 * time.Second
	}
	go clear(clearInterval)
	db, err := cache.InitDB(conf.DB)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	defer db.Close()

	for _, d := range conf.Directors {
		director.Append(&director.Config{
			Name:     d.Name,
			Policy:   d.Type,
			Ping:     d.Ping,
			Pass:     d.Pass,
			Prefix:   d.Prefix,
			Host:     d.Host,
			Backends: d.Backends,
		})
		proxy.AppendUpstream(&proxy.UpstreamConfig{
			Name:     d.Name,
			Policy:   d.Type,
			Ping:     d.Ping,
			Backends: d.Backends,
		})
	}
	if len(conf.Name) == 0 {
		conf.Name = string(vars.Name)
	}
	err = server.Start(&server.Config{
		Name:                 conf.Name,
		Concurrency:          conf.Concurrency,
		DisableKeepalive:     conf.DisableKeepalive,
		ReadBufferSize:       conf.ReadBufferSize,
		WriteBufferSize:      conf.WriteBufferSize,
		ETag:                 conf.ETag,
		ConnectTimeout:       conf.ConnectTimeout,
		ReadTimeout:          conf.ReadTimeout,
		WriteTimeout:         conf.WriteTimeout,
		MaxConnsPerIP:        conf.MaxConnsPerIP,
		MaxKeepaliveDuration: conf.MaxKeepaliveDuration,
		MaxRequestBodySize:   conf.MaxRequestBodySize,
		Listen:               conf.Listen,
		HitForPass:           conf.HitForPass,
		AdminPath:            conf.AdminPath,
		AdminToken:           conf.AdminToken,
		TextTypes:            conf.TextTypes,
		CertFile:             conf.CertFile,
		KeyFile:              conf.KeyFile,
		ResponseHeader:       conf.ResponseHeader,
		EnableServerTiming:   conf.EnableServerTiming,
		Favicon:              conf.Favicon,
		// HTTP日志相关
		LogFormat: conf.LogFormat,
		LogType:   conf.LogType,
		UDPLog:    conf.UDPLog,
		AccessLog: conf.AccessLog,
	})
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
