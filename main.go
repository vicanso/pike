package main

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/httplog"

	"github.com/vicanso/pike/vars"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/middleware"
	"github.com/vicanso/pike/proxy"
)

func startExpiredClearTask(client *cache.Client, interval time.Duration) {
	defer func() {
		if err := recover(); err != nil {
			startExpiredClearTask(client, interval)
			return
		}
	}()

	if interval == 0 {
		interval = 300 * time.Second
	}
	ticker := time.NewTicker(interval)
	for _ = range ticker.C {
		client.ClearExpired(60)
	}
}

func main() {
	// Echo instance
	e := echo.New()
	dc := config.GetDefault()

	level, _ := strconv.Atoi(os.Getenv("LVL"))
	if level != 0 {
		e.Logger.SetLevel(log.Lvl(level))
	}
	client := &cache.Client{
		Path: dc.DB,
	}
	err := client.Init()
	if err != nil {
		panic(err)
	}
	defer client.Close()

	go startExpiredClearTask(client, dc.ExpiredClearInterval)
	e.HTTPErrorHandler = custommiddleware.CreateErrorHandler(e, client)

	directors := make(proxy.Directors, 0)
	for _, item := range dc.Directors {
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

	// 配置logger中间件
	if len(dc.AccessLog) != 0 {

		writeCategory := httplog.Normal
		if dc.LogType == "date" {
			writeCategory = httplog.Date
		}
		var logWriter httplog.Writer
		udpPrefix := "udp://"
		if strings.HasPrefix(dc.AccessLog, udpPrefix) {
			logWriter = &httplog.UDPWriter{
				URI: dc.AccessLog[len(udpPrefix):],
			}
		} else {
			logWriter = &httplog.FileWriter{
				Path:     dc.AccessLog,
				Category: writeCategory,
			}
		}
		if logWriter != nil {
			defer logWriter.Close()
		}

		e.Use(custommiddleware.Logger(custommiddleware.LoggerConfig{
			LogFormat: dc.LogFormat,
			Writer:    logWriter,
		}))
	}

	// 对于websocke的直接不支持
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if c.IsWebSocket() {
				return vars.ErrNotSupportWebSocket
			}
			return next(c)
		}
	})

	e.Use(custommiddleware.Initialization(custommiddleware.InitializationConfig{
		Header: dc.Header,
	}))

	e.Use(custommiddleware.Identifier(client))

	e.Use(custommiddleware.DirectorPicker(directors))

	e.Use(custommiddleware.CacheFetcher(client))

	e.Use(custommiddleware.Proxy(custommiddleware.ProxyConfig{
		Timeout: dc.ConnectTimeout,
		ETag:    dc.ETag,
	}))

	e.Use(custommiddleware.HeaderSetter())

	e.Use(custommiddleware.FreshChecker())

	e.Use(custommiddleware.Dispatcher(custommiddleware.DispatcherConfig{
		CompressTypes:     dc.TextTypes,
		CompressMinLength: dc.CompressMinLength,
		CompressLevel:     dc.CompressLevel,
	}, client))

	// Start server
	e.Logger.Fatal(e.Start(dc.Listen))

}
