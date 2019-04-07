package df

const (
	// Version application's version
	Version = "1.0.0"
	// APP application's name
	APP = "pike"

	// GZIP gzip compress
	GZIP = "gzip"
	// BR brotli compress
	BR = "br"
)

// key set to context
const (
	Status = "status"
	Cache  = "cache"

	ProxyDoneCallback = "proxyDoneCallback"
)

var (
	// ConfigPathList config path list
	ConfigPathList = []string{
		".",
		"$HOME/.pike",
		"/etc/pike",
	}
)

// HTTP header
const (
	// HeaderAge http age header
	HeaderAge = "Age"
	// HeaderContentLength http content length header
	HeaderContentLength = "Content-Length"
)

var (
	// BuildedAt application builded at ???
	BuildedAt = ""
	// CommitID git commit id
	CommitID = ""
)
