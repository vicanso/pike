package vars

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
	// CacheControl http response cache control header
	CacheControl = "Cache-Control"
	// GzipEncoding gzip encoding
	GzipEncoding = "gzip"
	// BrEncoding br encoding
	BrEncoding = "br"
	// AdminToken admin token
	AdminToken = "X-Admin-Token"
	// PingURL health check ping url
	PingURL = "/ping"
)

var (
	// SpaceByte 空格
	SpaceByte = byte(' ')
)
