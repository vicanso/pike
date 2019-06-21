package main

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/vicanso/pike/df"
	"github.com/vicanso/pike/log"
	"github.com/vicanso/pike/server"
	"go.uber.org/zap"
)

var (
	// BuildedAt application builded at ???
	BuildedAt = ""
	// CommitID git commit id
	CommitID = ""
)

func init() {
	if BuildedAt != "" {
		t, _ := time.Parse("20060102.150405", BuildedAt)
		df.BuildedAt = t.Format(time.RFC3339)
	}
	df.CommitID = strings.ToUpper(CommitID)
}

func getListen() string {
	v := os.Getenv("LISTEN")
	if v == "" {
		v = ":3015"
	}
	return v
}

func main() {
	logger := log.Default()

	listen := getListen()

	logger.Info("pike is starting",
		zap.String("listen", listen),
	)

	cluster, err := server.NewCluster()
	if err != nil {
		panic(err)
	}

	s := http.Server{
		Addr:    listen,
		Handler: cluster,
	}
	err = s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
