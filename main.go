package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/vicanso/pike/config"

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
	defaultConfig := config.GetDefault()

	level, _ := strconv.Atoi(os.Getenv("LVL"))
	if level != 0 {
		e.Logger.SetLevel(log.Lvl(level))
	}
	client := &cache.Client{
		Path: defaultConfig.DB,
	}
	fmt.Println(defaultConfig.Directors[0])
	err := client.Init()
	if err != nil {
		panic(err)
	}
	defer client.Close()
	e.HTTPErrorHandler = custommiddleware.CreateErrorHandler(e, client)

	directors := make(proxy.Directors, 0)
	for _, item := range defaultConfig.Directors {
		d := &proxy.Director{
			Name:     item.Name,
			Policy:   item.Policy,
			Ping:     item.Ping,
			Backends: item.Backend,
			Hosts:    item.Host,
			Prefixs:  item.Prefix,
		}
		go d.StartHealthCheck(5 * time.Second)
		directors = append(directors, d)
	}

	// Middleware
	// e.Use(middleware.Logger())
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

	e.Use(custommiddleware.Identifier(client))

	e.Use(custommiddleware.DirectorPicker(directors))

	e.Use(custommiddleware.CacheFetcher(client))

	e.Use(custommiddleware.ProxyWithConfig(custommiddleware.ProxyConfig{}))

	e.Use(custommiddleware.Dispatcher(client))

	// Start server
	e.Logger.Fatal(e.Start(":3015"))

}
