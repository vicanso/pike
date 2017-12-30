package vars

import "errors"

var (
	// AcceptEncoding http request accept enconding header
	AcceptEncoding = []byte("Accept-Encoding")

	// ContentEncoding http response content encoding header
	ContentEncoding = []byte("Content-Encoding")

	// XForwardedFor http request x-forwarded-fox header
	XForwardedFor = []byte("X-Forwarded-For")

	// Gzip gzip compress
	Gzip = []byte("gzip")

	// Br brotli compress
	Br = []byte("br")

	// Get http get method
	Get = []byte("GET")

	// Head http head method
	Head = []byte("HEAD")

	// CacheControl http response cache control header
	CacheControl = []byte("Cache-Control")

	// ErrDirectorUnavailable 没有配置可用的director
	ErrDirectorUnavailable = errors.New("director unavailable")

	// ErrServiceUnavailable 服务器不可用
	ErrServiceUnavailable = errors.New("service unavailable")
	// ErrDbNotInit 没有初始化db
	ErrDbNotInit = errors.New("db isn't init")
)

const (
	// CompressMinLength the min length to gzip
	CompressMinLength = 1024
	// Random random policy
	Random = "random"
	// RoundRobin round robin policy
	RoundRobin = "roundRobin"
	// LeastConn least conn policy
	LeastConn = "leastConn"
	// IPHash ip hash policy
	IPHash = "ipHash"
	// URIHash uri hash policy
	URIHash = "uriHash"
	// First first policy
	First = "first"
	// Header policy
	Header = "header"
	// None request stauts: none
	None = "none"
	// Fetching request status: fetching
	Fetching = "fetching"
	// HitForPass request status: hitForPass
	HitForPass = "hitForPass"
)
