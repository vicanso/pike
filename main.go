package main

import (
	"os"

	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/log"
	"github.com/vicanso/pike/server"
	"go.uber.org/zap"
)

func main() {
	ins := server.Instance{
		EnabledAdminServer: os.Getenv("ENABLED_ADMIN_SERVER") != "",
	}

	err := ins.Start()
	if err != nil {
		panic(err)
	}
	logger := log.Default()
	config.Watch(func(changeType config.ChangeType, value string) {
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
	c := make(chan os.Signal)
	<-c
}
