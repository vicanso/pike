package vars

import "errors"

var (
	// AcceptEncoding http request accept enconding header
	AcceptEncoding = []byte("Accept-Encoding")

	// ContentEncoding http response content encoding header
	ContentEncoding = []byte("Content-Encoding")
	// ContentLength http response content length
	ContentLength = []byte("Content-Length")
	// XForwardedFor http request x-forwarded-fox header
	XForwardedFor = []byte("X-Forwarded-For")
	// IfModifiedSince http request IfModifiedSince header
	IfModifiedSince = []byte("If-Modified-Since")
	// IfNoneMatch http request IfNoneMatch header
	IfNoneMatch = []byte("If-None-Match")
	// Age http response age header
	Age = []byte("Age")
	// CacheControl http response cache control header
	CacheControl = []byte("Cache-Control")

	// Gzip gzip compress
	Gzip = []byte("gzip")

	// Br brotli compress
	Br = []byte("br")

	// Get http get method
	Get = []byte("GET")

	// Head http head method
	Head = []byte("HEAD")
	// LineBreak 换行符
	LineBreak = []byte("\n")
	// Colon 冒号
	Colon = []byte(":")[0]
	// Space 空格
	Space = []byte(" ")[0]

	// ErrDirectorUnavailable 没有配置可用的director
	ErrDirectorUnavailable = errors.New("director unavailable")

	// ErrServiceUnavailable 服务器不可用
	ErrServiceUnavailable = errors.New("service unavailable")
	// ErrDbNotInit 没有初始化db
	ErrDbNotInit = errors.New("db isn't init")
)

const (
	// Pass request status: pass
	Pass = iota
	// Fetching request status: fetching
	Fetching
	// Waiting request status: wating
	Waiting
	// HitForPass request status: hitForPass
	HitForPass
	// Cacheable request status: cacheable
	Cacheable
)

const (
	// RawData 原始数据
	RawData = iota
	// GzipData gzip压缩数据
	GzipData
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
)
