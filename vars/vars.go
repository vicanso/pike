package vars

import (
	"net/http"

	"github.com/labstack/echo"
)

const (
	// Status 请求状态（waiting fecthing等）
	Status = "status"
	// Identity 根据url生成的标识
	Identity = "identity"
	// Director 保存匹配的director
	Director = "director"
	// Response 响应数据
	Response = "response"
	// HitForPassTTL hit for pass的有效期
	HitForPassTTL = 600
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
	// CacheControl http response header
	CacheControl = "Cache-Control"
)

var (
	// ErrNotSupportWebSocket 不支持websocket
	ErrNotSupportWebSocket = echo.NewHTTPError(http.StatusNotImplemented, "not support web socket")
	// ErrDirectorNotFound 未找到可用的director
	ErrDirectorNotFound = echo.NewHTTPError(http.StatusNotImplemented, "director not found")
	// ErrRequestStatusNotSet 未设置请求的status
	ErrRequestStatusNotSet = echo.NewHTTPError(http.StatusNotImplemented, "request status not set")
	// ErrNoBackendAvaliable 没有可用的backend
	ErrNoBackendAvaliable = echo.NewHTTPError(http.StatusServiceUnavailable, "no backend avaliable")

	// CacheControl http response cache control header
	CacheControl = "Cache-Control"
	// ContentEncoding http response content encoding
	ContentEncoding = "Content-Encoding"
)
