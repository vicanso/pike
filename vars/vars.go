package vars

import "errors"

const (
	// Status 请求状态（waiting fecthing等）
	Status = "status"
	// Identity 根据url生成的标识
	Identity = "identity"
	// Director 保存匹配的director
	Director = "director"
	// Body http响应数据
	Body = "body"
	// Code http响应码
	Code = "code"
	// Header http响应头
	Header = "header"
)

var (
	// ErrNotSupportWebSocket 不支持websocket
	ErrNotSupportWebSocket = errors.New("not support web socket")
	// ErrDirectorNotFound 未找到可用的director
	ErrDirectorNotFound = errors.New("director not found")
	// ErrNoBackendAvaliable 没有可用的backend
	ErrNoBackendAvaliable = errors.New("no backend avaliable")

	// CacheControl http response cache control header
	CacheControl = "Cache-Control"

	// ContentEncoding http response content encoding
	ContentEncoding = "Content-Encoding"
)
