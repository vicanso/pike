package df

import "time"

const (
	// Version application's version
	Version = "2.0.1"
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
)

// HTTP header
const (
	// HeaderAge http age header
	HeaderAge = "Age"
	// HeaderStatus http response status
	HeaderStatus = "X-Status"
)

var (
	// BuildedAt application builded at ???
	BuildedAt = ""
	// CommitID git commit id
	CommitID = ""
	// StartedAt application started at ???
	StartedAt = time.Now().Format(time.RFC3339)
)
