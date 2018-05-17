package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	funk "github.com/thoas/go-funk"

	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/controller"
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

func getLogger(dc *config.Config) httplog.Writer {
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
	return logWriter
}

func check(conf *config.Config) {
	url := "http://127.0.0.1" + conf.Listen + "/ping"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
		return
	}
	statusCode := resp.StatusCode
	if statusCode < 200 || statusCode >= 400 {
		fmt.Println(err)
		os.Exit(1)
		return
	}
	os.Exit(0)
}

func main() {
	args := os.Args[1:]
	if funk.ContainsString(args, "version") {
		fmt.Println("Pike version " + vars.Version)
		return
	}
	var configFile string
	flag.StringVar(&configFile, "c", "./config.yml", "the config file")
	flag.Parse()
	// Echo instance
	e := echo.New()
	dc, err := config.InitFromFile(configFile)
	if err != nil {
		panic(err)
	}
	if funk.ContainsString(args, "test") {
		configJSON, err := json.MarshalIndent(dc, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(configJSON))
		fmt.Println("the config file test done")
		return
	}
	fmt.Println("pike config file: " + configFile)
	if funk.ContainsString(args, "check") {
		check(dc)
		return
	}

	level, _ := strconv.Atoi(os.Getenv("LVL"))
	if level != 0 {
		e.Logger.SetLevel(log.Lvl(level))
	}

	// 初始化缓存
	client := &cache.Client{
		Path: dc.DB,
	}
	err = client.Init()
	if err != nil {
		panic(err)
	}
	defer client.Close()
	// 定时任务清除过期缓存
	go startExpiredClearTask(client, dc.ExpiredClearInterval)

	// 自定义出错处理
	e.HTTPErrorHandler = custommiddleware.CreateErrorHandler(e, client)

	// 生成director列表
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
		d.RefreshPriority()
		// 定时检测director是否可用
		go d.StartHealthCheck(5 * time.Second)
		directors = append(directors, d)
	}
	sort.Sort(directors)

	// Middleware
	// e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// 配置logger中间件
	if len(dc.AccessLog) != 0 {
		logWriter := getLogger(dc)
		defer logWriter.Close()
		e.Use(custommiddleware.Logger(custommiddleware.LoggerConfig{
			LogFormat: dc.LogFormat,
			Writer:    logWriter,
		}))
	}

	defaultSkipper := middleware.DefaultSkipper
	if len(dc.AdminPath) != 0 {
		defaultSkipper = func(c echo.Context) bool {
			requestURI := c.Request().RequestURI
			// 对于ping检测 skip
			if requestURI == vars.PingURL {
				return true
			}
			// 对于管理后台请求 skip
			return strings.HasPrefix(requestURI, dc.AdminPath)
		}
	}

	// 对于websocke的直接不支持
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.IsWebSocket() {
				return vars.ErrNotSupportWebSocket
			}
			return next(c)
		}
	})

	// 初始化中间件的参数
	initConfig := custommiddleware.InitializationConfig{
		Header:  dc.Header,
		Skipper: defaultSkipper,
	}
	e.Use(custommiddleware.Initialization(initConfig))

	// 生成请求唯一标识与状态中间件
	idConfig := custommiddleware.IdentifierConfig{
		Skipper: defaultSkipper,
	}
	e.Use(custommiddleware.Identifier(idConfig, client))

	// 获取director的中间件
	directorConfig := custommiddleware.DirectorPickerConfig{
		Skipper: defaultSkipper,
	}
	e.Use(custommiddleware.DirectorPicker(directorConfig, directors))

	// 缓存读取中间件
	cacheFetcherConfig := custommiddleware.CacheFetcherConfig{
		Skipper: defaultSkipper,
	}
	e.Use(custommiddleware.CacheFetcher(cacheFetcherConfig, client))

	// 代理转发中间件
	proxyConfig := custommiddleware.ProxyConfig{
		Timeout:  dc.ConnectTimeout,
		ETag:     dc.ETag,
		Skipper:  defaultSkipper,
		Rewrites: dc.Rewrites,
	}
	e.Use(custommiddleware.Proxy(proxyConfig))

	// http响应头设置中间件
	headerSetterConfig := custommiddleware.HeaderSetterConfig{
		Skipper: defaultSkipper,
	}
	e.Use(custommiddleware.HeaderSetter(headerSetterConfig))

	// 判断客户端缓存请求是否fresh的中间件
	freshCheckerConfig := custommiddleware.FreshCheckerConfig{
		Skipper: defaultSkipper,
	}
	e.Use(custommiddleware.FreshChecker(freshCheckerConfig))

	// 响应数据处理中间件
	dispatcherConfig := custommiddleware.DispatcherConfig{
		CompressTypes:     dc.TextTypes,
		CompressMinLength: dc.CompressMinLength,
		CompressLevel:     dc.CompressLevel,
		Skipper:           defaultSkipper,
	}
	e.Use(custommiddleware.Dispatcher(dispatcherConfig, client))

	// ping检测
	e.GET(vars.PingURL, func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	// stats
	adminGroup := e.Group(dc.AdminPath, func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(vars.Directors, directors)
			c.Set(vars.CacheClient, client)
			return next(c)
		}
	}, func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := c.Request().Header.Get(vars.AdminToken)
			if token != dc.AdminToken {
				return vars.ErrTokenInvalid
			}
			return next(c)
		}
	})

	adminGroup.GET("/stats", controller.GetStats)

	adminGroup.GET("/directors", controller.GetDirectors)
	adminGroup.GET("/cacheds", controller.GetCachedList)
	adminGroup.DELETE("/cacheds/:key", controller.RemoveCached)

	// Start server
	e.Logger.Fatal(e.Start(dc.Listen))
}
