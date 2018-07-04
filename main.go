package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	// _ "net/http/pprof"

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

// buildAt 构建时间
var buildAt = "20180101.000000"

const (
	defaultExpiredClearInterval = 300 * time.Second
	maxIdleConns                = 5 * 1024
)

func startExpiredClearTask(client *cache.Client, interval time.Duration) {
	defer func() {
		if err := recover(); err != nil {
			startExpiredClearTask(client, interval)
			return
		}
	}()

	if interval == 0 {
		interval = defaultExpiredClearInterval
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
	if dc.AccessLog == "console" {
		logWriter = &httplog.Console{}
	} else if strings.HasPrefix(dc.AccessLog, udpPrefix) {
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

func getBuildAtDesc() string {
	reg := regexp.MustCompile(`(\d{4})(\d{2})(\d{2}).(\d{2})(\d{2})(\d{2})`)
	str := reg.ReplaceAllString(buildAt, "$1-$2-$3 $4:$5:$6.000Z")
	return strings.Replace(str, " ", "T", 1)
}

func main() {
	// go func() {
	// 	fmt.Println(http.ListenAndServe("0.0.0.0:6060", nil))
	// }()
	args := os.Args[1:]
	if funk.ContainsString(args, "version") {
		fmt.Println("Pike version " + vars.Version + ", build at " + getBuildAtDesc())
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
			Name:         item.Name,
			Policy:       item.Policy,
			Ping:         item.Ping,
			Backends:     item.Backend,
			Hosts:        item.Host,
			Prefixs:      item.Prefix,
			Rewrites:     item.Rewrites,
			TargetURLMap: make(map[string]*url.URL),
		}
		d.RefreshPriority()
		d.GenRewriteRegexp()
		d.SetTransport(
			&http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   dc.ConnectTimeout,
					KeepAlive: 30 * time.Second,
					DualStack: true,
				}).DialContext,
				MaxIdleConns:          maxIdleConns,
				MaxIdleConnsPerHost:   maxIdleConns,
				IdleConnTimeout:       10 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			})
		// 定时检测director是否可用
		go d.StartHealthCheck(5 * time.Second)
		directors = append(directors, d)
	}
	sort.Sort(directors)

	// Middleware
	e.Pre(middleware.Recover())

	// 对于websocke的直接不支持
	e.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.IsWebSocket() {
				return vars.ErrNotSupportWebSocket
			}
			return next(c)
		}
	})

	// 创建自定义的pike context
	e.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			pc := custommiddleware.NewContext(c)
			defer custommiddleware.ReleaseContext(pc)
			if !dc.EnableServerTiming {
				pc.DisableServerTiming()
			}
			requestURI := c.Request().RequestURI
			// 对于ping检测 skip
			if requestURI == vars.PingURL {
				pc.Skip = true
			}
			// 对于管理后台请求 skip
			if len(dc.AdminPath) != 0 && strings.HasPrefix(requestURI, dc.AdminPath) {
				pc.Skip = true
			}

			return next(pc)
		}
	})

	// 配置logger中间件
	if len(dc.AccessLog) != 0 {
		logWriter := getLogger(dc)
		defer logWriter.Close()
		e.Pre(custommiddleware.Logger(custommiddleware.LoggerConfig{
			LogFormat: dc.LogFormat,
			Writer:    logWriter,
		}))
	}

	defaultSkipper := func(c echo.Context) bool {
		pc, ok := c.(*custommiddleware.Context)
		if !ok {
			return true
		}
		return pc.Skip
	}

	// 初始化中间件的参数
	initConfig := custommiddleware.InitializationConfig{
		Header:      dc.Header,
		Skipper:     defaultSkipper,
		Concurrency: dc.Concurrency,
	}
	e.Pre(custommiddleware.Initialization(initConfig))

	// 生成请求唯一标识与状态中间件
	idConfig := custommiddleware.IdentifierConfig{
		Skipper: defaultSkipper,
	}
	e.Pre(custommiddleware.Identifier(idConfig, client))

	// 获取director的中间件
	directorConfig := custommiddleware.DirectorPickerConfig{
		Skipper: defaultSkipper,
	}
	e.Pre(custommiddleware.DirectorPicker(directorConfig, directors))

	// 缓存读取中间件
	cacheFetcherConfig := custommiddleware.CacheFetcherConfig{
		Skipper: defaultSkipper,
	}
	e.Pre(custommiddleware.CacheFetcher(cacheFetcherConfig, client))

	// 代理转发中间件
	proxyConfig := custommiddleware.ProxyConfig{
		ETag:     dc.ETag,
		Skipper:  defaultSkipper,
		Rewrites: dc.Rewrites,
		Timeout:  dc.ConnectTimeout,
	}
	e.Pre(custommiddleware.Proxy(proxyConfig))

	// http响应头设置中间件
	headerSetterConfig := custommiddleware.HeaderSetterConfig{
		Skipper: defaultSkipper,
	}
	e.Pre(custommiddleware.HeaderSetter(headerSetterConfig))

	// 判断客户端缓存请求是否fresh的中间件
	freshCheckerConfig := custommiddleware.FreshCheckerConfig{
		Skipper: defaultSkipper,
	}
	e.Pre(custommiddleware.FreshChecker(freshCheckerConfig))

	// 响应数据处理中间件
	dispatcherConfig := custommiddleware.DispatcherConfig{
		CompressTypes:     dc.TextTypes,
		CompressMinLength: dc.CompressMinLength,
		CompressLevel:     dc.CompressLevel,
		Skipper:           defaultSkipper,
	}
	e.Pre(custommiddleware.Dispatcher(dispatcherConfig, client))

	// 后续route需要使用原有的context
	e.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			pc, ok := c.(*custommiddleware.Context)
			if ok {
				return next(pc.Context)
			}
			return next(c)
		}
	})

	// ping检测
	e.GET(vars.PingURL, func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	// admin group
	adminGroup := e.Group(dc.AdminPath, func(next echo.HandlerFunc) echo.HandlerFunc {
		// 添加相应字段
		return func(c echo.Context) error {
			c.Set(vars.Directors, directors)
			c.Set(vars.CacheClient, client)
			return next(c)
		}
	}, func(next echo.HandlerFunc) echo.HandlerFunc {
		// 校验 token
		return func(c echo.Context) error {
			uri := c.Request().RequestURI
			ext := path.Ext(uri)
			// 静态文件不校验token
			if len(ext) != 0 {
				file := uri[len(dc.AdminPath)+1:]
				c.Set(vars.StaticFile, file)
				return next(c)
			}

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
	adminGroup.GET("/fetchings", controller.GetFetchingList)
	adminGroup.DELETE("/cacheds/:key", controller.RemoveCached)

	adminGroup.GET("/*", controller.Serve)

	// Start server
	e.Logger.Fatal(e.Start(dc.Listen))
}
