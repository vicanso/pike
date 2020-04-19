package main

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/vicanso/elton"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	e := elton.New()
	e.GET("/ping", func(c *elton.Context) error {
		c.BodyBuffer = bytes.NewBufferString("pong")
		return nil
	})
	e.GET("/no-cache", func(c *elton.Context) error {
		// 仅为了测试proxy时的损耗，处理尽可能简单
		c.NoCache()
		c.BodyBuffer = bytes.NewBufferString("Hello, World!")
		return nil
	})

	e.GET("/", func(c *elton.Context) (err error) {
		buf, err := ioutil.ReadFile("./repos.json")
		if err != nil {
			return
		}
		c.CacheMaxAge("10m")
		c.SetContentTypeByExt(".json")
		c.BodyBuffer = bytes.NewBuffer(buf)
		return
	})

	// http1与http2均支持
	e.Server = &http.Server{
		Handler: h2c.NewHandler(e, &http2.Server{}),
	}
	err := e.ListenAndServe(":3000")
	if err != nil {
		panic(err)
	}
}
