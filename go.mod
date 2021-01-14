module github.com/vicanso/pike

go 1.15

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

replace github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.5

require (
	github.com/andybalholm/brotli v1.0.1
	github.com/coreos/etcd v3.3.25+incompatible
	github.com/dustin/go-humanize v1.0.0
	github.com/frankban/quicktest v1.11.3 // indirect
	github.com/fsnotify/fsnotify v1.4.9
	github.com/go-playground/validator/v10 v10.4.1
	github.com/gobuffalo/packr/v2 v2.8.1
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e
	github.com/golang/snappy v0.0.2
	github.com/google/uuid v1.1.4 // indirect
	github.com/klauspost/compress v1.11.7
	github.com/pierrec/lz4 v2.6.0+incompatible
	github.com/robfig/cron/v3 v3.0.1
	github.com/shirou/gopsutil/v3 v3.20.12
	github.com/spf13/cobra v1.1.1
	github.com/stretchr/testify v1.7.0
	github.com/vicanso/elton v1.2.4
	github.com/vicanso/elton-jwt v1.1.1
	github.com/vicanso/hes v0.3.0
	github.com/vicanso/upstream v0.1.0
	go.uber.org/atomic v1.7.0
	go.uber.org/automaxprocs v1.3.0
	go.uber.org/zap v1.16.0
	golang.org/x/net v0.0.0-20201224014010-6772e930b67b
	gopkg.in/yaml.v2 v2.4.0
	sigs.k8s.io/yaml v1.2.0 // indirect
)
