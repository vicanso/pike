package upstream

import (
	"hash/fnv"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/vicanso/cod"
	"github.com/vicanso/hes"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/log"
	"github.com/vicanso/pike/util"
	UP "github.com/vicanso/upstream"
	"go.uber.org/zap"

	"github.com/go-yaml/yaml"
	"github.com/vicanso/pike/df"

	proxy "github.com/vicanso/cod-proxy"
)

var (
	errNoMatchUpstream = &hes.Error{
		StatusCode: http.StatusInternalServerError,
		Category:   df.APP,
		Message:    "no match upstream",
		Exception:  true,
	}
	errNoAvailableUpstream = &hes.Error{
		StatusCode: http.StatusInternalServerError,
		Category:   df.APP,
		Message:    "no available upstream",
		Exception:  true,
	}
	defaultTransport *http.Transport
)

const (
	// backupTag backup server tag
	backupTag = "|backup"

	policyFirst      = "first"
	policyRandom     = "random"
	policyRoundRobin = "roundRobin"
	policyLeastconn  = "leastconn"
	policyIPHash     = "ipHash"
	headerHashPrefix = "header:"
	cookieHashPrefix = "cookie:"
)

type (
	// Backend backend config
	Backend struct {
		Name          string
		Policy        string
		Ping          string
		RequestHeader []string `yaml:"requestHeader"`
		Header        []string
		Prefixs       []string
		Hosts         []string
		Backends      []string
		Rewrites      []string
	}
	// Upstream Upstream
	Upstream struct {
		Policy        string
		Priority      int
		Name          string
		Header        http.Header
		RequestHeader http.Header
		Server        UP.HTTP
		Hosts         []string
		Prefixs       []string
		Rewrites      []string
		Handler       cod.Handler
	}
	// Director director
	Director struct {
		sync.RWMutex
		URI       string
		Upstreams Upstreams
	}
	// Upstreams upstream list
	Upstreams []*Upstream
)

func init() {
	defaultTransport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   config.GetConnectTimeout(),
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       config.GetIdleConnTimeout(),
		TLSHandshakeTimeout:   config.GetTLSHandshakeTimeout(),
		ExpectContinueTimeout: config.GetExpectContinueTimeout(),
		ResponseHeaderTimeout: config.GetResponseHeaderTimeout(),
	}
}

// Fetch fetch upstreams
func (d *Director) Fetch() {
	d.Lock()
	defer d.Unlock()
	d.Upstreams = NewUpstreamsFromConfig()
}

// Proxy proxy
func (d *Director) Proxy(c *cod.Context) (err error) {
	d.RLock()
	defer d.RUnlock()
	var found *Upstream
	for _, item := range d.Upstreams {
		if item.Match(c) {
			found = item
			break
		}
	}
	if found == nil {
		return errNoMatchUpstream
	}
	// 设置请求头
	for key, values := range found.RequestHeader {
		for _, value := range values {
			c.AddRequestHeader(key, value)
		}
	}
	// 设置响应头
	for key, values := range found.Header {
		for _, value := range values {
			c.AddHeader(key, value)
		}
	}

	return found.Handler(c)
}

// ClearUpstreams clear upstreams
func (d *Director) ClearUpstreams() {
	for _, up := range d.Upstreams {
		up.Server.StopHealthCheck()
	}
	d.Upstreams = nil
}

// StartHealthCheck start health check
func (d *Director) StartHealthCheck() {
	logger := log.Default()
	for _, up := range d.Upstreams {
		up.Server.DoHealthCheck()
		up.Server.OnStatus(func(status int32, upstream *UP.HTTPUpstream) {
			logger.Info("upstream status change",
				zap.String("uri", upstream.URL.String()),
				zap.Int32("status", status),
			)
		})
		go up.Server.StartHealthCheck()
	}
}

func (s Upstreams) Len() int {
	return len(s)
}

func (s Upstreams) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Upstreams) Less(i, j int) bool {
	return s[i].Priority < s[j].Priority
}

// NewUpstreamsFromConfig new upstreams from config
func NewUpstreamsFromConfig() Upstreams {
	backendConfig := &struct {
		Director []Backend
	}{
		make([]Backend, 0),
	}
	for _, path := range df.ConfigPathList {
		file := filepath.Join(path, "backends.yml")
		buf, _ := ioutil.ReadFile(file)
		if len(buf) != 0 {
			err := yaml.Unmarshal(buf, backendConfig)
			if err != nil {
				break
			}
		}
	}

	upstreams := make(Upstreams, len(backendConfig.Director))
	for index, item := range backendConfig.Director {
		upstreams[index] = New(item)
	}
	sort.Sort(upstreams)
	return upstreams
}

// hash calculates a hash based on string s
func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// createTargetPicker create target picker function
func createTargetPicker(us *Upstream) proxy.TargetPicker {
	isHeaderPolicy := false
	isCookiePolicy := false
	key := ""
	policy := us.Policy
	if strings.HasPrefix(policy, headerHashPrefix) {
		key = policy[len(headerHashPrefix):]
		isHeaderPolicy = true
	} else if strings.HasPrefix(policy, cookieHashPrefix) {
		key = policy[len(cookieHashPrefix):]
		isCookiePolicy = true
	}
	server := &us.Server
	fn := func(c *cod.Context) (*url.URL, error) {
		var result *UP.HTTPUpstream
		switch policy {
		case policyFirst:
			result = server.PolicyFirst()
		case policyRandom:
			result = server.PolicyRandom()
		case policyRoundRobin:
			result = server.PolicyRoundRobin()
		case policyLeastconn:
			result = server.PolicyLeastconn()
			if result != nil {
				// 连接数+1
				result.Inc()
				// 设置 callback
				c.Set(df.ProxyDoneCallback, result.Dec)
			}
		case policyIPHash:
			result = server.GetAvailableUpstream(hash(c.RealIP()))
		default:
			var index uint32
			if isHeaderPolicy {
				index = hash(c.GetRequestHeader(key))
			} else if isCookiePolicy {
				cookie, _ := c.Cookie(key)
				if cookie != nil {
					index = hash(cookie.Value)
				}
			}
			result = server.GetAvailableUpstream(index)
		}
		if result == nil {
			return nil, errNoAvailableUpstream
		}
		return result.URL, nil
	}
	return fn
}

// createProxyHandler create proxy handler
func createProxyHandler(us *Upstream) cod.Handler {
	fn := createTargetPicker(us)

	cfg := proxy.Config{
		// Transport http transport
		Transport:    defaultTransport,
		TargetPicker: fn,
	}
	// 如果测试，则清除transport，方便测试
	if config.IsTest() {
		cfg.Transport = nil
	}
	if len(us.Rewrites) != 0 {
		cfg.Rewrites = us.Rewrites
	}
	return proxy.New(cfg)
}

func createUpstreamFromBackend(backend Backend) *Upstream {
	priority := 8
	if len(backend.Hosts) != 0 {
		priority -= 4
	}
	if len(backend.Prefixs) != 0 {
		priority -= 2
	}
	uh := UP.HTTP{
		// use http request check
		Ping: backend.Ping,
	}
	for _, item := range backend.Backends {
		backup := false
		if strings.Contains(item, backupTag) {
			item = strings.Replace(item, backupTag, "", 1)
			backup = true
		}
		item = strings.TrimSpace(item)
		if backup {
			uh.AddBackup(item)
		} else {
			uh.Add(item)
		}
	}

	us := Upstream{
		Policy:   backend.Policy,
		Name:     backend.Name,
		Server:   uh,
		Prefixs:  backend.Prefixs,
		Hosts:    backend.Hosts,
		Rewrites: backend.Rewrites,
		Priority: priority,
	}
	// 默认使用 round robin算法
	if us.Policy == "" {
		us.Policy = policyRoundRobin
	}

	h := util.ConvertToHTTPHeader(backend.Header)
	if h != nil {
		us.Header = h
	}
	rh := util.ConvertToHTTPHeader(backend.RequestHeader)
	if rh != nil {
		us.RequestHeader = rh
	}
	return &us
}

// New new upstream
func New(backend Backend) *Upstream {
	us := createUpstreamFromBackend(backend)
	us.Handler = createProxyHandler(us)
	return us
}

// Match match
func (us *Upstream) Match(c *cod.Context) bool {
	hosts := us.Hosts
	if len(hosts) != 0 {
		found := false
		currentHost := c.Request.Host
		for _, host := range hosts {
			if currentHost == host {
				found = true
				break
			}
		}
		// 不匹配host
		if !found {
			return false
		}
	}

	prefixs := us.Prefixs
	if len(prefixs) != 0 {
		found := false
		path := c.Request.URL.Path
		for _, prefix := range prefixs {
			if strings.HasPrefix(path, prefix) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
