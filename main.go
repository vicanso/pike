package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
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
	config.InitFromFile(configFile)
	conf := config.Current
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
	err = server.Start()
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
