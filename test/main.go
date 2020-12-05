package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/vicanso/elton"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func httpGet(url string) (data []byte, err error) {
	res, err := http.Get(url)
	if err != nil {
		return
	}
	defer res.Body.Close()
	data, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	return
}

func main() {
	e := elton.New()

	e.GET("/repos", func(c *elton.Context) (err error) {
		buf, err := httpGet("https://api.github.com/users/vicanso/repos")
		if err != nil {
			return
		}
		c.SetContentTypeByExt(".json")
		c.CacheMaxAge(5 * time.Minute)
		c.BodyBuffer = bytes.NewBuffer(buf)
		return
	})

	e.GET("/ping", func(c *elton.Context) (err error) {
		c.BodyBuffer = bytes.NewBufferString("pong")
		return
	})

	// http1与http2均支持
	e.Server = &http.Server{
		Handler: h2c.NewHandler(e, &http2.Server{}),
	}

	err := e.ListenAndServe(":3001")
	if err != nil {
		panic(err)
	}
}
