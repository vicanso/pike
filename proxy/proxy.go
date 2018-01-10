package proxy

import (
	"bytes"
	"sync/atomic"
	"time"

	"../util"
	"../vars"
	"github.com/valyala/fasthttp"
)

var (
	supportedPolicies = make(map[string]func(string) Policy)
)

// RegisterPolicy adds a custom policy to the proxy.
func RegisterPolicy(name string, policy func(string) Policy) {
	supportedPolicies[name] = policy
}

// Do 将当前请求转发至upstream
func Do(ctx *fasthttp.RequestCtx, us *Upstream, timeout time.Duration) (*fasthttp.Response, error) {
	policy := us.Policy
	uh := policy.Select(us.Hosts, ctx)
	if uh == nil {
		return nil, vars.ErrServiceUnavailable
	}
	atomic.AddInt64(&uh.Conns, 1)
	defer atomic.AddInt64(&uh.Conns, -1)
	uri := string(ctx.RequestURI())
	url := "http://" + uh.Host + uri
	req := fasthttp.AcquireRequest()
	reqHeader := &req.Header
	// 复制HTTP请求头
	ctx.Request.Header.CopyTo(reqHeader)

	// 设置x-forwarded-for
	xFor := vars.XForwardedFor
	orginalXFor := ctx.Request.Header.PeekBytes(xFor)
	clientIP := util.GetClientIP(ctx)
	if len(orginalXFor) == 0 {
		reqHeader.SetCanonical(xFor, []byte(clientIP))
	} else {
		// 如果原有HTTP头有x-forwarded-for
		reqHeader.SetCanonical(xFor, bytes.Join(
			[][]byte{orginalXFor, []byte(clientIP)},
			[]byte(",")))
	}
	postBody := ctx.PostBody()
	if len(postBody) != 0 {
		req.SetBody(postBody)
	}
	req.SetRequestURI(url)
	// 删除有可能304的处理
	reqHeader.DelBytes(vars.IfModifiedSince)
	reqHeader.DelBytes(vars.IfNoneMatch)
	// 设置支持gzip
	reqHeader.SetCanonical(vars.AcceptEncoding, vars.Gzip)
	resp := fasthttp.AcquireResponse()
	client := &fasthttp.Client{}
	var err error
	if timeout > 0 {
		err = client.DoTimeout(req, resp, timeout)
	} else {
		err = client.Do(req, resp)
	}
	if err != nil {
		if err == fasthttp.ErrTimeout {
			return nil, vars.ErrGatewayTimeout
		}
		return nil, err
	}
	return resp, nil
}

// CreateUpstream 创建一个新的Upstream
func CreateUpstream(name, policy, arg string) *Upstream {
	policyFunc := supportedPolicies[policy]
	if policyFunc == nil {
		policyFunc = supportedPolicies[vars.RoundRobin]
	}
	return &Upstream{
		Name:   name,
		Hosts:  make([]*UpstreamHost, 0),
		Policy: policyFunc(arg),
	}
}
