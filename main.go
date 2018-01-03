package main

import (
	"io/ioutil"
	"log"
	"sort"
	"time"

	"./cache"
	"./director"
	"./dispatch"
	"./proxy"
	"./util"
	"./vars"
	"github.com/valyala/fasthttp"
	"gopkg.in/yaml.v2"
)

const hitForPassTTL = 10

// PikeConfig 程序配置
type PikeConfig struct {
	Name string
	// pass 直接pass的url规则
	Directors []*director.Config
}

// DirectorSlice 用于director排序
type DirectorSlice []*director.Director

// Len 获取director slice的长度
func (s DirectorSlice) Len() int {
	return len(s)
}

// Swap 元素互换
func (s DirectorSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s DirectorSlice) Less(i, j int) bool {
	return s[i].Priority < s[j].Priority
}

var upstreamMap = make(map[string]*proxy.Upstream)
var directorList = make(DirectorSlice, 0, 10)

// getDirector 获取director
func getDirector(host, uri []byte) *director.Director {
	var found *director.Director
	// 查找可用的director
	for _, d := range directorList {
		if found == nil && d.Match(host, uri) {
			found = d
		}
	}
	return found
}

func doProxy(ctx *fasthttp.RequestCtx, us *proxy.Upstream) (*fasthttp.Response, []byte, []byte, error) {
	resp, err := proxy.Do(ctx, us)
	if err != nil {
		return nil, nil, nil, err
	}
	body, err := dispatch.GetResponseBody(resp)
	if err != nil {
		return nil, nil, nil, err
	}
	header := dispatch.GetResponseHeader(resp)
	return resp, header, body, nil
}

func handler(ctx *fasthttp.RequestCtx) {
	host := ctx.Request.Host()
	uri := ctx.RequestURI()
	found := getDirector(host, uri)
	if found == nil {
		// 没有可用的配置（）
		dispatch.ErrorHandler(ctx, vars.ErrDirectorUnavailable)
		return
	}
	us := upstreamMap[found.Name]
	// 判断该请求是否直接pass到backend
	isPass := util.Pass(ctx, found.Passes)
	status := vars.Pass
	var key []byte
	if !isPass {
		key = util.GenRequestKey(ctx)
		s, c := cache.GetRequestStatus(key)
		status = s
		if c != nil {
			status = <-c
		}
	}
	switch status {
	case vars.Pass:
		resp, header, body, err := doProxy(ctx, us)
		if err != nil {
			dispatch.ErrorHandler(ctx, err)
			return
		}
		ctx.Response.SetStatusCode(resp.StatusCode())
		dispatch.ResponseBytes(ctx, header, body)
	case vars.Fetching, vars.HitForPass:
		resp, header, body, err := doProxy(ctx, us)
		if err != nil {
			cache.HitForPass(key, hitForPassTTL)
			dispatch.ErrorHandler(ctx, err)
			return
		}
		ctx.Response.SetStatusCode(resp.StatusCode())
		dispatch.ResponseBytes(ctx, header, body)
		cacheAge := util.GetCacheAge(&resp.Header)
		if cacheAge == 0 {
			cache.HitForPass(key, hitForPassTTL)
		} else {
			bucket := []byte(found.Name)
			statusCode := uint16(resp.StatusCode())
			err = cache.SaveResponseData(bucket, key, body, header, statusCode, cacheAge)
			if err != nil {
				cache.HitForPass(key, hitForPassTTL)
			} else {
				cache.Cacheable(key, cacheAge)
			}
		}
	case vars.Cacheable:
		bucket := []byte(found.Name)
		respData, err := cache.GetResponse(bucket, key)
		if err != nil {
			dispatch.ErrorHandler(ctx, err)
			return
		}
		dispatch.ResponseBytes(ctx, respData.Header, respData.Body)
	}
}

func main() {
	buf, err := ioutil.ReadFile("./config.yml")
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	_, err = cache.InitDB("/tmp/pike.db")
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	conf := PikeConfig{}
	err = yaml.Unmarshal(buf, &conf)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	for _, directorConf := range conf.Directors {
		d := director.CreateDirector(directorConf)
		name := d.Name
		up := proxy.CreateUpstream(name, d.Policy, "")
		for _, backend := range d.Backends {
			up.AddBackend(backend)
		}
		up.StartHealthcheck(d.Ping, time.Second)
		upstreamMap[name] = up
		cache.InitBucket([]byte(name))
		directorList = append(directorList, d)
	}
	sort.Sort(directorList)

	server := &fasthttp.Server{
		Handler:     handler,
		Concurrency: 1024 * 1024,
	}
	server.ListenAndServe(":3015")
}
