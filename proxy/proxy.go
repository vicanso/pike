package proxy

import (
	"bytes"
	"errors"
	"net"
	"sync/atomic"

	"github.com/valyala/fasthttp"
	"github.com/vicanso/pike/vars"
)

var (
	supportedPolicies = make(map[string]func(string) Policy)
)

// RegisterPolicy adds a custom policy to the proxy.
func RegisterPolicy(name string, policy func(string) Policy) {
	supportedPolicies[name] = policy
}

// Do 将当前请求转发至upstream
func Do(ctx *fasthttp.RequestCtx, us *Upstream, isPass bool) (*fasthttp.Response, error) {
	policy := us.Policy
	uh := policy.Select(us.Hosts, ctx)
	if uh == nil {
		return nil, errors.New("No backend is avaliable")
	}
	atomic.AddInt64(&uh.Conns, 1)
	defer atomic.AddInt64(&uh.Conns, -1)
	uri := string(ctx.RequestURI())
	url := "http://" + uh.Host + uri
	req := fasthttp.AcquireRequest()
	// 复制HTTP请求头
	ctx.Request.Header.CopyTo(&req.Header)

	// 设置x-forwarded-for
	xFor := vars.XForwardedFor
	orginalXFor := ctx.Request.Header.PeekBytes(xFor)
	localIP := ctx.LocalAddr().String()
	if len(orginalXFor) == 0 {
		remoteAddr := ctx.RemoteAddr().String()
		clientIP, _, err := net.SplitHostPort(remoteAddr)
		if err != nil {
			clientIP = remoteAddr
		}
		req.Header.SetCanonical(xFor, []byte(clientIP))
	} else {
		// 如果原有HTTP头有x-forwarded-for
		req.Header.SetCanonical(xFor, bytes.Join(
			[][]byte{orginalXFor, []byte(localIP)},
			[]byte(",")))
	}
	postBody := ctx.PostBody()
	if len(postBody) != 0 {
		req.SetBody(postBody)
	}
	req.SetRequestURI(url)
	// 设置支持gzip
	req.Header.SetCanonical(vars.AcceptEncoding, vars.Gzip)
	resp := fasthttp.AcquireResponse()
	client := &fasthttp.Client{}
	err := client.Do(req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// CreateUpstream 创建一个新的Upstream
func CreateUpstream(name string, policy string) *Upstream {
	policyFunc := supportedPolicies[policy]
	if policyFunc == nil {
		policyFunc = supportedPolicies[vars.RoundRobin]
	}
	return &Upstream{
		Name:   name,
		Hosts:  make([]*UpstreamHost, 0),
		Policy: policyFunc(name),
	}
}
