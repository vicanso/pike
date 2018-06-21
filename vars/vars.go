package vars

import (
	"net/http"

	"github.com/labstack/echo"
)

const (
	// Version 版本号
	Version = "1.0.0"
	// CacheClient cache client
	CacheClient = "cacheClient"
	// Directors the director list
	Directors = "directors"
	// StaticFile static file name
	StaticFile = "static-file"

	// HitForPassTTL hit for pass的有效期
	HitForPassTTL = 600
	// CompressMinLength the min length to gzip
	CompressMinLength = 1024
	// Age http response age
	Age = "Age"
	// XStatus cache状态
	XStatus = "X-Status"
	// Pass pass status
	Pass = "pass"
	// Fetching fetching status
	Fetching = "fetching"
	// HitForPass hit for pass status
	HitForPass = "hitForPass"
	// Cacheable cacheable status
	Cacheable = "cacheable"
	// CacheControl http response cache control header
	CacheControl = "Cache-Control"
	// SetCookie http set cookie header
	SetCookie = "Set-Cookie"
	// GzipEncoding gzip encoding
	GzipEncoding = "gzip"
	// BrEncoding br encoding
	BrEncoding = "br"
	// IfNoneMatch http request IfNoneMatch header
	IfNoneMatch = "If-None-Match"
	// ETag http response etag
	ETag = "ETag"
	// ServerTiming http response server timing
	ServerTiming = "Server-Timing"
	// AdminToken admin token
	AdminToken = "X-Admin-Token"
	// PingURL health check ping url
	PingURL = "/ping"
)

var (
	// ErrNotSupportWebSocket 不支持websocket
	ErrNotSupportWebSocket = echo.NewHTTPError(http.StatusNotImplemented, "not support web socket")
	// ErrDirectorNotFound 未找到可用的director
	ErrDirectorNotFound = echo.NewHTTPError(http.StatusNotImplemented, "director not found")
	// ErrRequestStatusNotSet 未设置请求的status
	ErrRequestStatusNotSet = echo.NewHTTPError(http.StatusNotImplemented, "request status not set")
	// ErrIdentityNotSet 未设置Identity
	ErrIdentityNotSet = echo.NewHTTPError(http.StatusNotImplemented, "identity not set")
	// ErrResponseNotSet 未设置response
	ErrResponseNotSet = echo.NewHTTPError(http.StatusNotImplemented, "response not set")
	// ErrContentEncodingNotSupport 未支持此content encoding
	ErrContentEncodingNotSupport = echo.NewHTTPError(http.StatusNotImplemented, "content enconding not support")
	// ErrBodyCotentNotFound 未找到content
	ErrBodyCotentNotFound = echo.NewHTTPError(http.StatusInternalServerError, "body content not found")
	// ErrNoBackendAvaliable 没有可用的backend
	ErrNoBackendAvaliable = echo.NewHTTPError(http.StatusServiceUnavailable, "no backend avaliable")
	// ErrGatewayTimeout 网关超时
	ErrGatewayTimeout = echo.NewHTTPError(http.StatusGatewayTimeout, "gateway timeout")
	// ErrTooManyRequest 太多的请求正在处理中
	ErrTooManyRequest = echo.NewHTTPError(http.StatusTooManyRequests, "too many request is handling")
	// ErrTokenInvalid token校验失败
	ErrTokenInvalid = echo.NewHTTPError(http.StatusUnauthorized, "token is invalid")
)
