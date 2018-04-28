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
	// Timing timing
	Timing = "timing"
	// Fresh request fresh flag
	Fresh = "fresh"

	// PikeMetric pike总体的处理时间
	PikeMetric = "0PIKE"
	// WaitForRequestStatusMetric 等待请求状态的时间
	WaitForRequestStatusMetric = "0WRFS"
	// GetResponseFromCacheMetric 从缓存数据库中取数据的时间
	GetResponseFromCacheMetric = "0GRFC"
	// GetResponseFromProxyMetric 从proxy中获取数据的时间
	GetResponseFromProxyMetric = "0GRFP"
	// DispatchResponseMetric 响应数据的处理时间
	DispatchResponseMetric = "0DR"

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
)

var (
	// ErrNotSupportWebSocket 不支持websocket
	ErrNotSupportWebSocket = echo.NewHTTPError(http.StatusNotImplemented, "not support web socket")
	// ErrDirectorNotFound 未找到可用的director
	ErrDirectorNotFound = echo.NewHTTPError(http.StatusNotImplemented, "director not found")
	// ErrRequestStatusNotSet 未设置请求的status
	ErrRequestStatusNotSet = echo.NewHTTPError(http.StatusNotImplemented, "request status not set")
	// ErrIdentityStatusNotSet 未设置Identity
	ErrIdentityStatusNotSet = echo.NewHTTPError(http.StatusNotImplemented, "identity not set")
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
)
