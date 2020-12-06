package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/compress"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/location"
	"github.com/vicanso/pike/log"
	"github.com/vicanso/pike/server"
	"github.com/vicanso/pike/upstream"
	"go.uber.org/zap"
)

var (
	// BuildedAt application builded at ???
	BuildedAt = "" // nolint
	// CommitID git commit id
	CommitID = "" // nolint
)

// alarmURL 告警发送的地址
var alarmURL string

// doAlarm 发送告警
func doAlarm(category, message string) {
	if alarmURL == "" {
		return
	}
	data := fmt.Sprintf(`{"application": "pike", "category": "%s", "message": "%s"}`, category, message)
	resp, err := http.Post(alarmURL, "application/json", bytes.NewBufferString(data))
	if err != nil {
		log.Default().Error("do alarm fail",
			zap.Error(err),
		)
		return
	}
	defer resp.Body.Close()
	result, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		log.Default().Error("do alarm fail",
			zap.Int("status", resp.StatusCode),
			zap.String("result", string(result)),
		)
	}
}

func update() (err error) {
	pikeConfig, err := config.Read()
	if err != nil {
		return
	}
	// 重置压缩列表
	compress.Reset(pikeConfig.Compresses)
	// 重置默认dispatcher列表
	cache.ResetDispatchers(pikeConfig.Caches)
	// 重置默认的upstream列表
	upstream.ResetWithOnStats(pikeConfig.Upstreams, func(si upstream.StatusInfo) {
		log.Default().Info("upstream status change",
			zap.String("name", si.Name),
			zap.String("status", si.Status),
			zap.String("addr", si.URL),
		)

		if si.Status == "sick" {
			message := fmt.Sprintf("%s is %s, addr: %s", si.Name, si.Status, si.URL)
			go doAlarm("upstream", message)
		}
	})
	// 重置location列表
	location.Reset(pikeConfig.Locations)

	server.Reset(pikeConfig.Servers)
	return server.Start()
}

func startAdminServer(addr string) error {
	pikeConfig, err := config.Read()
	if err != nil {
		return err
	}
	return server.StartAdminServer(server.AdminServerConfig{
		Addr:     addr,
		User:     pikeConfig.Admin.User,
		Password: pikeConfig.Admin.Password,
	})
}

func run() {
	logger := log.Default()

	go config.Watch(func() {
		err := update()
		if err != nil {
			logger.Error("update config fail",
				zap.Error(err),
			)
			go doAlarm("config", err.Error())
		} else {
			logger.Info("update config success")
		}
	})

	err := update()
	if err != nil {
		panic(err)
	}
}

func isHelpCmd() bool {
	for _, arg := range os.Args {
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}
func main() {
	logger := log.Default()
	server.SetBuildInfo(BuildedAt, CommitID)
	defer config.Close()
	configURL := ""
	adminAddr := ""

	var rootCmd = &cobra.Command{
		Use:   "pike",
		Short: "Pike is a http cache server",
		PreRun: func(cmd *cobra.Command, args []string) {
			// 初始化配置
			err := config.InitDefaultClient(configURL)
			if err != nil {
				panic(err)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			if adminAddr != "" {
				go func() {
					err := startAdminServer(adminAddr)
					if err != nil {
						logger.Error("start admin server fail",
							zap.String("addr", adminAddr),
							zap.Error(err),
						)
						go doAlarm("admin", adminAddr+", "+err.Error())
					}
				}()
			}
			run()
		},
	}
	rootCmd.Flags().StringVar(&configURL, "config", "pike.yml", "The config of pike, support etcd or file, etcd://127.0.0.1:2379/pike or /opt/pike")
	rootCmd.Flags().StringVar(&adminAddr, "admin", "", "The address of admin web page, e.g.: :9013")
	rootCmd.Flags().StringVar(&alarmURL, "alarm", "", "The alarm request url, alarm will post to the url, e.g.: http://192.168.1.2:3000/alarms")

	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
	if isHelpCmd() {
		return
	}

	logger.Info("pike is running")
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	for range c {
		// TODO 添加graceful close
		os.Exit(0)
	}
}
