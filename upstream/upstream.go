package upstream

import (
	"hash/fnv"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/vicanso/cod"
	"github.com/vicanso/hes"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/util"
	UP "github.com/vicanso/upstream"

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
	isTestMode = os.Getenv("GO_MODE") == "test"
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
	// ServerInfo upstream server's information
	ServerInfo struct {
		URL    string `json:"url,omitempty"`
		Backup bool   `json:"backup,omitempty"`
		Status string `json:"status,omitempty"`
	}
	// Info upstream's information
	Info struct {
		Policy        string       `json:"policy,omitempty"`
		Priority      int          `json:"priority,omitempty"`
		Name          string       `json:"name,omitempty"`
		Ping          string       `json:"ping,omitempty"`
		Header        http.Header  `json:"header,omitempty"`
		RequestHeader http.Header  `json:"requestHeader,omitempty"`
		Hosts         []string     `json:"hosts,omitempty"`
		Prefixs       []string     `json:"prefixs,omitempty"`
		Rewrites      []string     `json:"rewrites,omitempty"`
		Servers       []ServerInfo `json:"servers,omitempty"`
	}
	// OnStatusChange on status change
	OnStatusChange func(up *UP.HTTPUpstream, status string)
	// Director director
	Director struct {
		Transport      *http.Transport
		Upstreams      Upstreams
		OnStatusChange OnStatusChange
	}
	// Upstreams upstream list
	Upstreams []*Upstream
)

// SetBackends set backends
func (d *Director) SetBackends(backends []config.BackendConfig) {
	upstreams := make(Upstreams, len(backends))
	for index, item := range backends {
		upstreams[index] = New(item, d.Transport)
	}
	sort.Sort(upstreams)
	d.Upstreams = upstreams
}

// Proxy proxy
func (d *Director) Proxy(c *cod.Context) (err error) {
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
	// 则 health check 也需要调整
	for _, up := range d.Upstreams {
		up.Server.OnStatus(func(status int32, upstream *UP.HTTPUpstream) {
			if d.OnStatusChange != nil {
				d.OnStatusChange(upstream, UP.ConvertStatusToString(status))
			}
		})
		up.Server.DoHealthCheck()
		go up.Server.StartHealthCheck()
	}
}

// GetUpstreamInfos get upstream information of director
func (d *Director) GetUpstreamInfos() []Info {
	statsInfo := make([]Info, len(d.Upstreams))
	for index, up := range d.Upstreams {
		servers := make([]ServerInfo, 0)
		for _, item := range up.Server.GetUpstreamList() {
			servers = append(servers, ServerInfo{
				URL:    item.URL.String(),
				Backup: item.Backup,
				Status: item.StatusDesc(),
			})
		}
		info := Info{
			Policy:        up.Policy,
			Priority:      up.Priority,
			Name:          up.Name,
			Ping:          up.Server.Ping,
			Header:        up.Header,
			RequestHeader: up.RequestHeader,
			Hosts:         up.Hosts,
			Prefixs:       up.Prefixs,
			Rewrites:      up.Rewrites,
			Servers:       servers,
		}
		statsInfo[index] = info
	}
	return statsInfo
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
	fn := func(c *cod.Context) (*url.URL, proxy.Done, error) {
		var result *UP.HTTPUpstream
		var done proxy.Done
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
				done = func(_ *cod.Context) {
					result.Dec()
				}
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
			return nil, done, errNoAvailableUpstream
		}
		return result.URL, done, nil
	}
	return fn
}

// createProxyHandler create proxy handler
func createProxyHandler(us *Upstream, transport *http.Transport) cod.Handler {
	fn := createTargetPicker(us)

	cfg := proxy.Config{
		// Transport http transport
		Transport:    transport,
		TargetPicker: fn,
	}
	// 如果测试，则清除transport，方便测试
	if isTestMode {
		cfg.Transport = nil
	}
	if len(us.Rewrites) != 0 {
		cfg.Rewrites = us.Rewrites
	}
	return proxy.New(cfg)
}

func createUpstreamFromBackend(backend config.BackendConfig) *Upstream {
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

	h := util.ConvertToHTTPHeader(backend.ResponseHeader)
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
func New(backend config.BackendConfig, transport *http.Transport) *Upstream {
	us := createUpstreamFromBackend(backend)
	us.Handler = createProxyHandler(us, transport)
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
