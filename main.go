package main

import (
	"encoding/json"
	"flag"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	// _ "net/http/pprof"

	log "github.com/sirupsen/logrus"
	funk "github.com/thoas/go-funk"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/controller"
	"github.com/vicanso/pike/httplog"
	"github.com/vicanso/pike/middleware"
	"github.com/vicanso/pike/pike"
	"github.com/vicanso/pike/vars"
)

// buildAt 构建时间
var buildAt = "20180101.000000"
var disabledPing int32 = 0

const (
	defaultExpiredClearInterval = 300 * time.Second
	maxIdleConns                = 5 * 1024
)

// startExpiredClearTask 定时清理过期数据
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

// check 检查程序是否正常运行
func check(conf *config.Config) {
	url := "http://127.0.0.1" + conf.Listen + "/ping"
	resp, err := http.Get(url)
	if err != nil {
		log.Error("health check fail, ", err)
		os.Exit(1)
		return
	}
	statusCode := resp.StatusCode
	if statusCode < 200 || statusCode >= 400 {
		log.Errorf("helth check fail, status:%d", statusCode)
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
	// 初始化日志输出级别
	logLevel := os.Getenv("LVL")
	if logLevel != "" {
		lv, _ := strconv.Atoi(logLevel)
		log.SetLevel(log.Level(lv))
	}

	// go func() {
	// 	fmt.Println(http.ListenAndServe("0.0.0.0:6060", nil))
	// }()
	args := os.Args[1:]
	if funk.ContainsString(args, "version") {
		log.Infof("Pike version %s build at %s", vars.Version, getBuildAtDesc())
		return
	}
	var configFile string
	flag.StringVar(&configFile, "c", "./config.yml", "the config file")
	flag.Parse()
	dc, err := config.InitFromFile(configFile)
	if err != nil {
		panic(err)
	}
	if funk.ContainsString(args, "test") {
		configJSON, err := json.MarshalIndent(dc, "", "  ")
		if err != nil {
			panic(err)
		}
		log.Infof("the config file test done, config: %s", string(configJSON))
		return
	}
	if funk.ContainsString(args, "check") {
		check(dc)
		return
	}
	log.Infof("start pike use the config: %s", configFile)

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

	// 生成director列表
	directors := make(pike.Directors, 0)
	for _, item := range dc.Directors {
		policy := item.Policy
		err := pike.AddPolicySelectFunc(policy)
		if err != nil {
			log.Panic("create policy fail, ", err)
		}
		d := &pike.Director{
			Name:          item.Name,
			Policy:        policy,
			Ping:          item.Ping,
			Backends:      item.Backends,
			Hosts:         item.Hosts,
			Prefixs:       item.Prefixs,
			Rewrites:      item.Rewrites,
			RequestHeader: item.RequestHeader,
			Header:        item.Header,
			TargetURLMap:  make(map[string]*url.URL),
		}
		d.Prepare()
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

	p := pike.New()

	p.ErrorHandler = middleware.CreateErrorHandler(client)

	// ping health check
	p.Use(middleware.Ping(middleware.PingConfig{
		DisabledPing: &disabledPing,
		URL:          "/ping",
	}))

	// admin管理后台
	adminConfig := controller.AdminConfig{
		Prefix:       dc.AdminPath,
		Token:        dc.AdminToken,
		Client:       client,
		Directors:    directors,
		DisabledPing: &disabledPing,
	}
	p.Use(controller.AdminHandler(adminConfig))

	// 配置logger中间件
	if len(dc.AccessLog) != 0 {
		logWriter := getLogger(dc)
		defer logWriter.Close()
		p.Use(middleware.Logger(middleware.LoggerConfig{
			LogFormat: dc.LogFormat,
			Writer:    logWriter,
		}))
	}

	p.Use(middleware.Recover(middleware.DefaultRecoverConfig))

	// 初始化中间件的参数
	initConfig := middleware.InitializationConfig{
		Header:        dc.Header,
		RequestHeader: dc.RequestHeader,
		Concurrency:   dc.Concurrency,
	}
	p.Use(middleware.Initialization(initConfig))

	// 生成请求唯一标识与状态中间件
	p.Use(middleware.Identifier(middleware.IdentifierConfig{
		Format: dc.Identity,
	}, client))

	// 获取director的中间件
	p.Use(middleware.DirectorPicker(middleware.DirectorPickerConfig{}, directors))

	// 缓存读取中间件
	p.Use(middleware.CacheFetcher(middleware.CacheFetcherConfig{}, client))

	// 代理转发中间件
	proxyConfig := middleware.ProxyConfig{
		ETag:     dc.ETag,
		Rewrites: dc.Rewrites,
		Timeout:  dc.ConnectTimeout,
	}
	p.Use(middleware.Proxy(proxyConfig))

	// http响应头设置中间件
	headerSetterConfig := middleware.HeaderSetterConfig{}
	p.Use(middleware.HeaderSetter(headerSetterConfig))

	// 判断客户端缓存请求是否fresh的中间件
	freshCheckerConfig := middleware.FreshCheckerConfig{}
	p.Use(middleware.FreshChecker(freshCheckerConfig))

	// 响应数据处理中间件
	dispatcherConfig := middleware.DispatcherConfig{
		CompressTypes:     dc.TextTypes,
		CompressMinLength: dc.CompressMinLength,
		CompressLevel:     dc.CompressLevel,
	}
	p.Use(middleware.Dispatcher(dispatcherConfig, client))

	p.ListenAndServe(dc.Listen)
}
