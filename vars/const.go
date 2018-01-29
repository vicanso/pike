package vars

import "errors"

var (
	// Version the application version
	Version = "0.1.0"
	// Name the name of application
	Name = []byte("Pike")
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
	// ETag http response etag header
	ETag = []byte("ETag")
	// LastModified httpresponse LastModified header
	LastModified = []byte("LastModified")
	// ServerTiming http response ServerTiming header
	ServerTiming = []byte("Server-Timing")
	// XCache whether the data of response is from cache
	XCache = []byte("X-Cache")
	// XCacheHit the response hit the cache
	XCacheHit = []byte("hit")
	// XCacheMiss the response miss the cache
	XCacheMiss = []byte("miss")

	// Age http response age header
	Age = []byte("Age")
	// CacheControl http response cache control header
	CacheControl = []byte("Cache-Control")
	// ContentType http response content type header
	ContentType = []byte("Content-Type")
	// MultipartFormData http multi part form post
	MultipartFormData = []byte("multipart/form-data")
	// JSON http response content type json
	JSON = []byte("application/json; charset=utf-8")
	// NoCache http response cache-control: no-cache
	NoCache = []byte("no-cache")

	// Gzip gzip compress
	Gzip = []byte("gzip")

	// Br brotli compress
	Br = []byte("br")

	// Get http get method
	Get = []byte("GET")

	// Head http head method
	Head = []byte("HEAD")
	// LineBreak 换行符
	LineBreak = []byte("\r\n")
	// Colon 冒号
	Colon = byte(':')
	// Space 空格
	Space = byte(' ')
)

var (
	// PingURL the ping url
	PingURL = []byte("/ping")
	// FaviconURL the favicon url
	FaviconURL = []byte("/favicon.ico")
)

// errors
var (
	// ErrDirectorUnavailable 没有配置可用的director
	ErrDirectorUnavailable = errors.New("director unavailable")
	// ErrServiceUnavailable 服务器不可用
	ErrServiceUnavailable = errors.New("service unavailable")
	// ErrGatewayTimeout 网关超时（未能从upstream中获取数据)）
	ErrGatewayTimeout = errors.New("gateway timeout")
	// ErrDbNotInit 没有初始化db
	ErrDbNotInit = errors.New("db isn't init")
	// ErrAccessIsNotAlloed 不允许访问
	ErrAccessIsNotAlloed = errors.New("access is not allowed")
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
