package proxy

import (
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"
)

// UpstreamHost proxy upstream
type UpstreamHost struct {
	// Conns 连接数
	Conns     int64
	MaxConns  int64
	Host      string
	Fails     int32
	Successes int32
	Healthy   int32
}

// UpstreamHostPool 保存Upstream列表
type UpstreamHostPool []*UpstreamHost

// Full 检查当前upstream是否已满载
func (uh *UpstreamHost) Full() bool {
	return uh.MaxConns > 0 && atomic.LoadInt64(&uh.Conns) >= uh.MaxConns
}

// Available 判断当前upstream是否可用
func (uh *UpstreamHost) Available() bool {
	return atomic.LoadInt32(&uh.Healthy) != 0 && !uh.Full()
}

func (uh *UpstreamHost) healthCheck(ping string) {
	url := "http://" + uh.Host + ping
	interval := time.Second
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	resp := fasthttp.AcquireResponse()
	client := &fasthttp.Client{}
	start := time.Now()
	err := client.DoTimeout(req, resp, 3*time.Second)
	// 每个upstream每次只有一个health check在运行
	if err != nil {
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
	// 根据调用时间决定延时
	time.Sleep(interval - use)
	go uh.healthCheck(ping)
}

// Upstream 保存backend列表
type Upstream struct {
	Name   string
	Hosts  UpstreamHostPool
	Policy Policy
}

// StartHealthcheck 启动 health check
func (us *Upstream) StartHealthcheck(ping string) {
	// url := uh.Host + uh.Ping
	// log.Println(url)
	for _, uh := range us.Hosts {
		go uh.healthCheck(ping)
	}
}

// AddBackend 增加upstream的backend
func (us *Upstream) AddBackend(host string) *UpstreamHost {
	if us.Hosts == nil {
		us.Hosts = make([]*UpstreamHost, 0)
	}
	uh := &UpstreamHost{
		Conns:     0,
		MaxConns:  0,
		Host:      host,
		Fails:     0,
		Successes: 0,
		Healthy:   0,
	}
	us.Hosts = append(us.Hosts, uh)
	return uh
}
