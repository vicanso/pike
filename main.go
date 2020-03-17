package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/vicanso/pike/application"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/log"
	"github.com/vicanso/pike/server"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
)

var (
	// BuildedAt application builded at ???
	BuildedAt = "" // nolint
	// CommitID git commit id
	CommitID = "" // nolint
)

// cmds
var (
	// 配置保存的路径，支持etcd或者文件形式
	configPath string
	// initMode模式在首次未配时服务时启用
	initMode bool
)
var rootCmd = &cobra.Command{
	Use:   "Pike",
	Short: "Pike is a very fast http cache server",
	Long: fmt.Sprintf(`Pike support gzip and brotli compress.
Versions: build at %s, commit id is %s`, BuildedAt, CommitID),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.NewConfig(configPath)
		if err != nil {
			log.Default().Error(err.Error())
			os.Exit(1)
		}
		startServer(cfg)
	},
}

func init() {
	app := application.Default()
	app.SetBuildedAt(BuildedAt)
	app.SetCommitID(CommitID)
	// 测试模式自动添加启动参数
	if os.Getenv("GO_MODE") == "test" {
		rootCmd.SetArgs([]string{
			"--config",
			"etcd://127.0.0.1:2379/pike",
			"--init",
		})
	}

	rootCmd.Flags().StringVarP(&configPath, "config", "c", "", "the config's address, E.g: etcd://127.0.0.1:6379/pike or /tmp/pike (required)")
	_ = rootCmd.MarkFlagRequired("config")

	rootCmd.Flags().BoolVar(&initMode, "init", false, "init mode will enabled server listen on :3015")

	_, _ = maxprocs.Set(maxprocs.Logger(func(format string, args ...interface{}) {
		value := fmt.Sprintf(format, args...)
		log.Default().Info(value)
	}))
}

func startServer(cfg *config.Config) {
	ins := server.Instance{
		Config:             cfg,
		EnabledAdminServer: initMode,
	}

	err := ins.Start()
	if err != nil {
		panic(err)
	}
	logger := log.Default()
	cfg.Watch(func(changeType config.ChangeType, value string) {
		err := ins.Fetch()
		if err != nil {
			logger.Error("fetch config fail",
				zap.Error(err),
			)
			return
		}
		ins.Restart()
	})

	// TODO 增加监听信息关闭服务
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	for s := range c {
		switch s {
		case syscall.SIGINT:
			fallthrough
		case syscall.SIGTERM:
			fallthrough
		case syscall.SIGQUIT:
			cfg.Close()
			// TODO 将server设置为stop，延时退出
			os.Exit(0)
		default:
			logger.Info("exit should use sigquit")
		}
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Default().Error(err.Error())
		os.Exit(1)
	}
}
