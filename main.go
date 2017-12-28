package main

import (
	"errors"
	"io/ioutil"
	"log"
	"sort"

	"github.com/valyala/fasthttp"
	"github.com/vicanso/pike/director"
	"github.com/vicanso/pike/dispatch"
	"github.com/vicanso/pike/proxy"
	"github.com/vicanso/pike/util"
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

func handler(ctx *fasthttp.RequestCtx) {
	// log.Println(string(ctx.Request.Host()))
	// log.Println(string(ctx.Request.RequestURI()))
	host := ctx.Request.Host()
	uri := ctx.RequestURI()

	var found *director.Director
	for _, d := range directorList {
		if found == nil && d.Match(host, uri) {
			found = d
		}
	}
	if found == nil {
		// 没有可用的backend
		err := errors.New("No avaliable backend")
		dispatch.ErrorHandler(ctx, err)
		return
	}
	us := upstreamMap[found.Name]
	// 判断该请求是否直接pass到backend
	isPass := util.Pass(ctx, found.Passes)
	resp, err := proxy.Do(ctx, us, isPass)
	if err != nil {
		dispatch.ErrorHandler(ctx, err)
		return
	}
	dispatch.Response(ctx, resp)
	// resp.Header.CopyTo(&ctx.Response.Header)
	// 对压缩数据的处理，是否需要生成新的ETag
	// Content-Encoding 的处理
	// ctx.SetBody(resp.Body())
	// Content-Length 的处理
	// log.Println(ctx.ID())
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
		up := proxy.CreateUpstream(name, d.Policy)
		for _, backend := range d.Backends {
			up.AddBackend(backend)
		}
		up.StartHealthcheck(d.Ping)
		upstreamMap[name] = up
		directorList = append(directorList, d)
	}
	sort.Sort(directorList)

	fasthttp.ListenAndServe(":3015", handler)
}
