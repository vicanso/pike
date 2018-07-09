package middleware

import (
	"net/http"

	"github.com/vicanso/pike/pike"
)

const (
	// HitForPassTTL hit for pass的有效期
	HitForPassTTL = 600
)

var (
	// ErrDirectorNotFound 未找到可用的director
	ErrDirectorNotFound = pike.NewHTTPError(http.StatusNotImplemented, "director not found")
	// ErrRequestStatusNotSet 未设置请求的status
	ErrRequestStatusNotSet = pike.NewHTTPError(http.StatusNotImplemented, "request status not set")

	// ErrIdentityNotSet 未设置Identity
	ErrIdentityNotSet = pike.NewHTTPError(http.StatusNotImplemented, "identity not set")

	// ErrResponseNotSet 未设置response
	ErrResponseNotSet = pike.NewHTTPError(http.StatusNotImplemented, "response not set")

	// ErrContentEncodingNotSupport 未支持此content encoding
	ErrContentEncodingNotSupport = pike.NewHTTPError(http.StatusNotImplemented, "content enconding not support")
	// ErrNoBackendAvaliable 没有可用的backend
	ErrNoBackendAvaliable = pike.NewHTTPError(http.StatusServiceUnavailable, "no backend avaliable")
	// ErrGatewayTimeout 网关超时
	ErrGatewayTimeout = pike.NewHTTPError(http.StatusGatewayTimeout, "gateway timeout")
	// ErrTooManyRequest 太多的请求正在处理中
	ErrTooManyRequest = pike.NewHTTPError(http.StatusTooManyRequests, "too many request is handling")

// // ErrTokenInvalid token校验失败
// ErrTokenInvalid = pike.NewHTTPError(http.StatusUnauthorized, "token is invalid")
)
