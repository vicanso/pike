package main

import (
	"log"
	"runtime"
	"time"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/server"
)

func clear(interval time.Duration) {
	time.Sleep(interval)
	cache.ClearExpired()
	go clear(interval)
}

func main() {
	conf := config.Current
	if conf.Cpus > 0 {
		runtime.GOMAXPROCS(conf.Cpus)
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
	err = server.Start()
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
