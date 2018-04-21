package main

import (
	"github.com/vicanso/pike/cache"
)

func main() {
	c := &cache.Client{
		Path: "~/tmp/pike",
	}
	c.Init()

}
