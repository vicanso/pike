package proxy

import (
	"strings"
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"
)

// UpstreamHost proxy upstream
type UpstreamHost struct {
	// Conns 连接数
	Conns int64 `json:"connections"`
	// 最大连接数
	MaxConns int64 `json:"maxConnections"`
	// 对应的host配置
	Host string `json:"host"`
	// 失败次数
	Fails int32 `json:"fails"`
	// 成功次数
	Successes int32 `json:"success"`
	// 该节点是否健康
	Healthy int32 `json:"healthy"`
	// 表示该upstream为禁止状态
	Disabled bool `json:"disabled"`
	// 该节点对应的client
	Client *fasthttp.HostClient `json:"-"`
}

// UpstreamHostPool 保存Upstream列表
type UpstreamHostPool []*UpstreamHost

// Full 检查当前upstream是否已满载
func (uh *UpstreamHost) Full() bool {
	return uh.MaxConns > 0 && atomic.LoadInt64(&uh.Conns) >= uh.MaxConns
}

// Available 判断当前upstream是否可用
func (uh *UpstreamHost) Available() bool {
	healthy := atomic.LoadInt32(&uh.Healthy)
	// disabled 通过配置修改，每次检测，可以忽略原子性的问题
	return !uh.Disabled && healthy != 0 && !uh.Full()
}

// Disable 禁用该upstream
func (uh *UpstreamHost) Disable() {
	uh.Disabled = true
}

// Enable 启用该upstream
func (uh *UpstreamHost) Enable() {
	uh.Disabled = false
}

func (uh *UpstreamHost) healthCheck(ping string, interval time.Duration) {
	if len(ping) == 0 {
		uh.Healthy = 1
		return
	}
	url := uh.Host + ping
	if !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}
	if interval <= 0 {
		interval = time.Second
	}

	// 如果该upstream为禁止状态，则直接延时做health check
	if uh.Disabled {
		// 等待时间调整为2倍
		time.Sleep(2 * interval)
		go uh.healthCheck(ping, interval)
		return
	}
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	resp := fasthttp.AcquireResponse()
	client := &fasthttp.Client{}
	start := time.Now()
	err := client.DoTimeout(req, resp, 3*time.Second)
	statusCode := resp.StatusCode()
	// 每个upstream每次只有一个health check在运行
	if err != nil || (statusCode >= 400) {
		uh.Fails++
	} else {
		uh.Successes++
	}
	// 如果检测有3次成功，则backend可用
	if uh.Successes > 3 {
		uh.Fails = 0
		uh.Successes = 0
		if uh.Healthy == 0 {
			uh.Healthy = 1
		}
	} else if uh.Fails > 3 {
		// 如果检测有3次失败，则backend不可用
		uh.Fails = 0
		uh.Successes = 0
		if uh.Healthy == 1 {
			uh.Healthy = 0
		}
	}
	use := time.Since(start)
	delay := interval - use
	// 如果该upstream为不可用，检测时间缩短，加快检测
	if uh.Healthy == 0 {
		delay = delay / 2
	}
	if delay < 0 {
		delay = 0
	}
	// 根据调用时间决定延时
	time.Sleep(delay)
	go uh.healthCheck(ping, interval)
}

// Upstream 保存backend列表
type Upstream struct {
	Name   string           `json:"name"`
	Hosts  UpstreamHostPool `json:"hosts"`
	Policy Policy           `json:"policy"`
}

// UpstreamConfig upstream的配置
type UpstreamConfig struct {
	Name     string
	Backends []string
	Policy   string
	Arg      string
	Ping     string
}

// StartHealthcheck 启动 health check
func (us *Upstream) StartHealthcheck(ping string, interval time.Duration) {
	for _, uh := range us.Hosts {
		go uh.healthCheck(ping, interval)
	}
}

// AddBackend 增加upstream的backend
func (us *Upstream) AddBackend(host string) *UpstreamHost {
	if us.Hosts == nil {
		us.Hosts = make([]*UpstreamHost, 0)
	}
	addr := host
	index := strings.Index(addr, "//")
	isTLS := false
	if index != -1 {
		if addr[0:index] == "https:" {
			isTLS = true
		}
		addr = addr[index+2:]
	}
	uh := &UpstreamHost{
		Conns:     0,
		MaxConns:  0,
		Host:      host,
		Fails:     0,
		Successes: 0,
		Healthy:   0,
		Client: &fasthttp.HostClient{
			IsTLS: isTLS,
			Addr:  addr,
		},
	}
	us.Hosts = append(us.Hosts, uh)
	return uh
}
