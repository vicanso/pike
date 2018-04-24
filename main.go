package main

import (
	"os"
	"strconv"
	"time"

	"github.com/vicanso/pike/vars"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/middleware"
	"github.com/vicanso/pike/proxy"
)

func main() {
	// Echo instance
	e := echo.New()

	level, _ := strconv.Atoi(os.Getenv("LVL"))
	if level != 0 {
		e.Logger.SetLevel(log.Lvl(level))
	}
	client := &cache.Client{
		Path: "/tmp/test.cache",
	}

	err := client.Init()
	if err != nil {
		panic(err)
	}
	defer client.Close()
	e.HTTPErrorHandler = customMiddleware.CreateErrorHandler(e, client)

	directors := make(proxy.Directors, 0)
	d := &proxy.Director{
		Name: "aslant",
		Ping: "/ping",
		Backends: []string{
			"http://127.0.0.1:5018",
		},
	}
	go d.StartHealthCheck(5 * time.Second)
	directors = append(directors, d)

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// 对于websocke的直接不支持
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if c.IsWebSocket() {
				return vars.ErrNotSupportWebSocket
			}
			return next(c)
		}
	})

	e.Use(customMiddleware.Identifier(client))

	e.Use(customMiddleware.DirectorPicker(directors))

	e.Use(customMiddleware.CacheFetcher(client))

	// e.Use(middleware.Gzip())

	e.Use(customMiddleware.ProxyWithConfig(customMiddleware.ProxyConfig{}))

	e.Use(customMiddleware.Dispatcher(client))

	// Start server
	e.Logger.Fatal(e.Start(":3015"))

}
