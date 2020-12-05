// MIT License

// Copyright (c) 2020 Tree Xie

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Location相关处理函数，根据host,url判断当前请求所属location

package location

import (
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/log"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

// Location location config
type (
	Location struct {
		Name     string
		Upstream string
		Prefixes []string
		Rewrites []string
		Hosts    []string
		// Querystrings   []string
		ProxyTimeout   time.Duration
		ResponseHeader http.Header
		RequestHeader  http.Header
		Query          url.Values
		URLRewriter    Rewriter
		priority       atomic.Int32
	}
	rewriteRegexp struct {
		Regexp *regexp.Regexp
		Value  string
	}
)
type Rewriter func(req *http.Request)

// Locations location list
type Locations struct {
	mutex     *sync.RWMutex
	locations []*Location
}

var defaultLocations = NewLocations()

func captureTokens(pattern *regexp.Regexp, input string) *strings.Replacer {
	groups := pattern.FindAllStringSubmatch(input, -1)
	if groups == nil {
		return nil
	}
	values := groups[0][1:]
	replace := make([]string, 2*len(values))
	for i, v := range values {
		j := 2 * i
		replace[j] = "$" + strconv.Itoa(i+1)
		replace[j+1] = v
	}
	return strings.NewReplacer(replace...)
}

// generateURLRewriter generate url rewriter
func generateURLRewriter(arr []string) Rewriter {
	size := len(arr)
	if size == 0 {
		return nil
	}
	rewrites := make([]*rewriteRegexp, 0, size)

	for _, value := range arr {
		arr := strings.Split(value, ":")
		if len(arr) != 2 {
			continue
		}
		k := arr[0]
		v := arr[1]
		k = strings.Replace(k, "*", "(\\S*)", -1)
		reg, err := regexp.Compile(k)
		if err != nil {
			log.Default().Error("rewrite compile error",
				zap.String("value", k),
				zap.Error(err),
			)
			continue
		}
		rewrites = append(rewrites, &rewriteRegexp{
			Regexp: reg,
			Value:  v,
		})
	}
	if len(rewrites) == 0 {
		return nil
	}
	return func(req *http.Request) {
		urlPath := req.URL.Path
		for _, rewrite := range rewrites {
			replacer := captureTokens(rewrite.Regexp, urlPath)
			if replacer != nil {
				urlPath = replacer.Replace(rewrite.Value)
			}
		}
		req.URL.Path = urlPath
	}
}

// Match check location's hosts and prefixes match host/url
func (l *Location) Match(host, url string) bool {
	if len(l.Hosts) != 0 {
		found := false
		for _, item := range l.Hosts {
			if item == host {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if len(l.Prefixes) != 0 {
		found := false
		for _, item := range l.Prefixes {
			if strings.HasPrefix(url, item) {
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

func (l *Location) mergeHeader(dst, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

// AddRequestHeader add request header
func (l *Location) AddRequestHeader(header http.Header) {
	l.mergeHeader(header, l.RequestHeader)
}

// AddResponseHeader add response header
func (l *Location) AddResponseHeader(header http.Header) {
	l.mergeHeader(header, l.ResponseHeader)
}

// ShouldModifyQuery should modify query
func (l *Location) ShouldModifyQuery() bool {
	return len(l.Query) != 0
}

// AddQuery add query to request
func (l *Location) AddQuery(req *http.Request) {
	query := req.URL.Query()
	for key, values := range l.Query {
		for _, value := range values {
			query.Add(key, value)
		}
	}
	req.URL.RawQuery = query.Encode()
}

func (l *Location) getPriority() int {
	priority := l.priority.Load()
	if priority != 0 {
		return int(priority)
	}
	// 默认设置为8
	priority = 8
	if len(l.Prefixes) != 0 {
		priority -= 4
	}
	if len(l.Hosts) != 0 {
		priority -= 2
	}
	l.priority.Store(priority)
	return int(priority)
}

// NewLocations new a location list
func NewLocations(opts ...Location) *Locations {
	ls := &Locations{
		mutex: &sync.RWMutex{},
	}
	ls.Set(opts)
	return ls
}

// Set set location list
func (ls *Locations) Set(locations []Location) {
	data := make([]*Location, len(locations))
	for index, l := range locations {
		p := &l
		p.URLRewriter = generateURLRewriter(p.Rewrites)
		data[index] = p
	}

	// Sort sort locations
	sort.Slice(data, func(i, j int) bool {
		return data[i].getPriority() < data[j].getPriority()
	})
	ls.mutex.Lock()
	defer ls.mutex.Unlock()
	ls.locations = data
}

// GetLocations get locations
func (ls *Locations) GetLocations() []*Location {
	ls.mutex.RLock()
	defer ls.mutex.RUnlock()
	locations := ls.locations
	return locations
}

// Get get match location
func (ls *Locations) Get(host, url string, names ...string) *Location {
	locations := ls.GetLocations()
	for _, item := range locations {
		for _, name := range names {
			if item.Name == name && item.Match(host, url) {
				return item
			}
		}
	}
	return nil
}

// enhanceGetValue 如果以$开头，则优先从env中获取，如果获取失败，则直接返回原值
func enhanceGetValue(key string) string {
	if strings.HasPrefix(key, "$") {
		return os.Getenv(key[1:])
	}
	return key
}

func convertConfigs(configs []config.LocationConfig) []Location {
	locations := make([]Location, 0)
	fn := func(arr []string) http.Header {
		h := make(http.Header)
		for _, value := range arr {
			arr := strings.Split(value, ":")
			if len(arr) != 2 {
				continue
			}
			h.Add(enhanceGetValue(arr[0]), enhanceGetValue(arr[1]))
		}
		return h
	}
	// 将配置转换为header与url.values
	for _, item := range configs {
		d, _ := time.ParseDuration(item.ProxyTimeout)
		l := Location{
			Name:         item.Name,
			Upstream:     item.Upstream,
			Prefixes:     item.Prefixes,
			Rewrites:     item.Rewrites,
			Hosts:        item.Hosts,
			ProxyTimeout: d,
		}
		l.ResponseHeader = fn(item.RespHeaders)
		l.RequestHeader = fn(item.ReqHeaders)
		if len(item.QueryStrings) != 0 {
			query := make(url.Values)
			for _, str := range item.QueryStrings {
				arr := strings.Split(str, ":")
				if len(arr) != 2 {
					continue
				}
				query.Add(enhanceGetValue(arr[0]), enhanceGetValue(arr[1]))
			}
			l.Query = query
		}

		locations = append(locations, l)
	}
	return locations
}

// Reset reset location list to default
func Reset(configs []config.LocationConfig) {
	defaultLocations.Set(convertConfigs(configs))
}

// Get get location form default locations
func Get(host, url string, names ...string) *Location {
	return defaultLocations.Get(host, url, names...)
}
