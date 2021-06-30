module github.com/vicanso/pike

go 1.16

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

replace github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.5

require (
	github.com/andybalholm/brotli v1.0.3
	github.com/coreos/etcd v3.3.25+incompatible
	github.com/dgraph-io/badger/v3 v3.2103.0
	github.com/dustin/go-humanize v1.0.0
	github.com/frankban/quicktest v1.13.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9
	github.com/go-playground/validator/v10 v10.6.1
	github.com/go-redis/redis/v8 v8.11.0
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da
	github.com/golang/snappy v0.0.3
	github.com/google/uuid v1.2.0 // indirect
	github.com/klauspost/compress v1.13.1
	github.com/pierrec/lz4 v2.6.1+incompatible
	github.com/robfig/cron/v3 v3.0.1
	github.com/shirou/gopsutil/v3 v3.21.5
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	github.com/vicanso/elton v1.4.2
	github.com/vicanso/elton-jwt v1.2.1
	github.com/vicanso/hes v0.3.9
	github.com/vicanso/upstream v0.2.0
	go.mongodb.org/mongo-driver v1.5.3
	go.uber.org/atomic v1.8.0
	go.uber.org/automaxprocs v1.4.0
	go.uber.org/zap v1.18.1
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.4.0
	sigs.k8s.io/yaml v1.2.0 // indirect
)
