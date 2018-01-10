package main

import (
	"flag"
	"io/ioutil"
	"log"
	"runtime"
	"time"

	"./cache"
	"./director"
	"./server"
	"./util"
	"gopkg.in/yaml.v2"
)

var (
	config string
)

func clear(interval time.Duration) {
	time.Sleep(interval)
	cache.ClearExpired()
	go clear(interval)
}

func main() {
	flag.StringVar(&config, "c", "/etc/pike/config.yml", "the config file")
	flag.Parse()
	buf, err := ioutil.ReadFile(config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	conf := &server.PikeConfig{}
	util.Debug("conf:%v", conf)
	err = yaml.Unmarshal(buf, conf)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	if conf.Cpus > 0 {
		runtime.GOMAXPROCS(conf.Cpus)
	}
	go clear(conf.ExpiredClearInterval)

	db, err := cache.InitDB(conf.DB)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	defer db.Close()
	directorList := director.GetDirectors(conf.Directors)
	err = server.Start(conf, directorList)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
