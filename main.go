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
	"./vars"
	"github.com/valyala/fasthttp"
	"gopkg.in/yaml.v2"
)

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
	// isPass := util.Pass(ctx, found.Passes)
	key := string(host) + string(uri)
	status := cache.GetStatus(key)
	log.Println(status)
	// log.Println(key)
	// // hitForPass := cache.HitForPass(key)
	// // log.Println(hitForPass)
	// // 判断相同的key有没有fetching，没有则此请求需要从upstream获取数据
	// isFetching := cache.IsFetching(key)
	// // TODO
	// if isFetching {
	// 	// 进入等待队列
	// 	return
	// }
	// log.Println(isFetching)

	resp, err := proxy.Do(ctx, us)
	if status == "none" {
		// cache.ResetFetching(key)
	}
	if err != nil {
		dispatch.ErrorHandler(ctx, err)
		return
	}
	dispatch.Response(ctx, resp)
}

func main() {
	buf, err := ioutil.ReadFile("./config.yml")
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
		directorList = append(directorList, d)
	}
	sort.Sort(directorList)

	fasthttp.ListenAndServe(":3015", handler)
}
