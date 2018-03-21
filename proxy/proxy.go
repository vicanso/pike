package proxy

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/vicanso/pike/util"
	"github.com/vicanso/pike/vars"
)

// Config proxy的配置
type Config struct {
	Timeout time.Duration
	ETag    bool
}

var (
	supportedPolicies = make(map[string]func(string) Policy)
	upstreams         = make(map[string]*Upstream)
	httpBytes         = []byte("http")
	httpProtoBytes    = []byte("http://")
)

// genETag 获取数据对应的ETag
func genETag(buf []byte) string {
	size := len(buf)
	if size == 0 {
		return "\"0-2jmj7l5rSw0yVb_vlWAYkK_YBwk=\""
	}
	h := sha1.New()
	h.Write(buf)
	hash := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("\"%x-%s\"", size, hash)
}

// RegisterPolicy adds a custom policy to the proxy.
func RegisterPolicy(name string, policy func(string) Policy) {
	supportedPolicies[name] = policy
}

// Do 将当前请求转发至upstream
func Do(ctx *fasthttp.RequestCtx, us *Upstream, config *Config) (*fasthttp.Response, error) {
	reqHeader := &ctx.Request.Header
	startedAt := time.Now()
	defer func() {
		ms := util.GetTimeConsuming(startedAt)
		reqHeader.SetCanonical(vars.TimingFetch, []byte(strconv.Itoa(ms)))
	}()
	policy := us.Policy
	uh := policy.Select(us.Hosts, ctx)
	if uh == nil {
		return nil, vars.ErrServiceUnavailable
	}
	atomic.AddInt64(&uh.Conns, 1)
	defer atomic.AddInt64(&uh.Conns, -1)

	// 设置x-forwarded-for
	xFor := vars.XForwardedFor
	orginalXFor := reqHeader.PeekBytes(xFor)

	ip := []byte(ctx.RemoteIP().String())
	if len(orginalXFor) == 0 {
		reqHeader.SetCanonical(xFor, ip)
	} else {
		// 如果原有HTTP头有x-forwarded-for
		reqHeader.SetCanonical(xFor, bytes.Join(
			[][]byte{orginalXFor, ip},
			[]byte(",")))
	}
	// 由于pike的未支持304的方式从缓存中读取
	// 因此调用接口时，需要将有可能导致304的HTTP暂时删除
	ifModifiedSince := reqHeader.PeekBytes(vars.IfModifiedSince)
	ifNoneMatch := reqHeader.PeekBytes(vars.IfNoneMatch)
	if len(ifModifiedSince) != 0 {
		reqHeader.DelBytes(vars.IfModifiedSince)
	}
	if len(ifNoneMatch) != 0 {
		reqHeader.DelBytes(vars.IfNoneMatch)
	}
	client := uh.Client
	req := &ctx.Request
	resp := &fasthttp.Response{}
	var err error
	timeout := config.Timeout
	// TODO 等待fasthttp修复
	// 由于fasthttp中使用 do timeout时，会复制req，但是未将multipart数据复制
	// 由此使用此方式临时修复（将超时设置为0）
	if timeout > 0 && bytes.HasPrefix(reqHeader.PeekBytes(vars.ContentType), vars.MultipartFormData) {
		timeout = 0
	}
	if timeout > 0 {
		err = client.DoTimeout(req, resp, timeout)
	} else {
		err = client.Do(req, resp)
	}
	// 请求结束后，将相应的HTTP头重置
	if len(ifModifiedSince) != 0 {
		reqHeader.SetCanonical(vars.IfModifiedSince, ifModifiedSince)
	}
	if len(ifNoneMatch) != 0 {
		reqHeader.SetCanonical(vars.IfNoneMatch, ifNoneMatch)
	}
	if err != nil {
		if err == fasthttp.ErrTimeout {
			return nil, vars.ErrGatewayTimeout
		}
		return nil, err
	}
	// 如果程序没有生成ETag，自动填充
	if config.ETag && len(resp.Header.PeekBytes(vars.ETag)) == 0 {
		etag := genETag(resp.Body())
		resp.Header.SetCanonical(vars.ETag, []byte(etag))
	}
	return resp, nil
}

// AppendUpstream 创建一个新的Upstream
func AppendUpstream(upstreamConfig *UpstreamConfig) {
	policyFunc := supportedPolicies[upstreamConfig.Policy]
	if policyFunc == nil {
		policyFunc = supportedPolicies[vars.RoundRobin]
	}
	name := upstreamConfig.Name
	up := &Upstream{
		Name:   name,
		Hosts:  make([]*UpstreamHost, 0),
		Policy: policyFunc(upstreamConfig.Arg),
	}
	for _, backend := range upstreamConfig.Backends {
		up.AddBackend(backend)
	}
	up.StartHealthcheck(upstreamConfig.Ping, time.Second)
	upstreams[name] = up
}

// GetUpStream 获取upstream
func GetUpStream(name string) *Upstream {
	return upstreams[name]
}

// List 获取所有的upstream
func List() []*Upstream {
	streams := make([]*Upstream, 0)
	for _, v := range upstreams {
		streams = append(streams, v)
	}
	return streams
}
