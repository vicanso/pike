package server

import (
	"net"
	"net/http"
	"os"
	"regexp"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/vicanso/cod"

	recover "github.com/vicanso/cod-recover"
	"github.com/vicanso/hes"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/log"
	"github.com/vicanso/pike/stats"
	"github.com/vicanso/pike/upstream"
	"github.com/vicanso/pike/util"
	UP "github.com/vicanso/upstream"
	"go.uber.org/zap"
)

type (
	// Cluster server cluster
	Cluster struct {
		mutex          sync.Mutex
		ins            *Instance
		dsp            *cache.Dispatcher
		basicConfig    *config.Config
		directorConfig *config.DirectorConfig
	}
	// Instance server instance
	Instance struct {
		Director *upstream.Director
		Server   *cod.Cod
	}
)

// Destroy destroy instance
func (ins *Instance) Destroy() {
	ins.Director.ClearUpstreams()
}

// NewInstance create a new instance
func NewInstance(basicConfig *config.Config, directorConfig *config.DirectorConfig, dsp *cache.Dispatcher) (ins *Instance) {
	logger := log.Default()
	insStats := stats.New()

	timeoutConfig := basicConfig.Data.Timeout
	director := &upstream.Director{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   timeoutConfig.Connect,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       timeoutConfig.IdleConn,
			TLSHandshakeTimeout:   timeoutConfig.TLSHandshake,
			ExpectContinueTimeout: timeoutConfig.ExpectContinue,
			ResponseHeaderTimeout: timeoutConfig.ResponseHeader,
		},
		OnStatusChange: func(upstream *UP.HTTPUpstream, status string) {
			logger.Info("upstream status change",
				zap.String("uri", upstream.URL.String()),
				zap.String("status", status),
			)
		},
	}

	director.SetBackends(directorConfig.GetBackends())

	director.StartHealthCheck()

	d := NewServer(Options{
		BasicConfig:    basicConfig,
		DirectorConfig: directorConfig,
		Director:       director,
		Dispatcher:     dsp,
		Stats:          insStats,
	})

	// 出错时输出相关出错日志
	d.OnError(func(c *cod.Context, err error) {
		he := hes.Wrap(err)
		// 如果是recover，记录统计
		if he.Category == recover.ErrCategory {
			insStats.IncreaseRecoverCount()
		}
		logger.Error(he.Message,
			zap.String("category", he.Category),
			zap.String("method", c.Request.Method),
			zap.String("host", c.Request.Host),
			zap.String("url", c.Request.RequestURI),
			zap.Strings("stack", util.GetStack(4, 9)),
		)
	})
	ins = &Instance{
		Director: director,
		Server:   d,
	}
	return
}

func (cls *Cluster) newInstance() *Instance {
	return NewInstance(cls.basicConfig, cls.directorConfig, cls.dsp)
}

func (cls *Cluster) newDispatcher() *cache.Dispatcher {
	basicConfig := cls.basicConfig
	cacheConfig := basicConfig.Data.Cache
	compresConfig := basicConfig.Data.Compress
	textFilter := regexp.MustCompile(compresConfig.Filter)
	dsp := cache.NewDispatcher(cache.Options{
		Size:              cacheConfig.Size,
		ZoneSize:          cacheConfig.Zone,
		CompressLevel:     compresConfig.Level,
		CompressMinLength: compresConfig.MinLength,
		HitForPassTTL:     int(cacheConfig.HitForPass.Seconds()),
		TextFilter:        textFilter,
	})
	return dsp
}

// SetInstance set instance to cluster
func (cls *Cluster) SetInstance(ins *Instance) (oldIns *Instance) {
	oldPoint := atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&cls.ins)), unsafe.Pointer(ins))
	if oldPoint == nil {
		return
	}
	oldIns = (*Instance)(oldPoint)
	return
}

// GetInstance get instance from cluster
func (cls *Cluster) GetInstance() *Instance {
	return (*Instance)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&cls.ins))))
}

// ToggleInstance toggle instance
func (cls *Cluster) ToggleInstance(updateConfig string) (err error) {
	log.Default().Info("toggle instance",
		zap.String("name", updateConfig),
	)
	cls.mutex.Lock()
	defer cls.mutex.Unlock()
	done := false
	defer func() {
		if done {
			return
		}
		go func() {
			// 延时一分钟后重新执行
			time.Sleep(time.Minute)
			cls.ToggleInstance(updateConfig)
		}()
	}()
	switch updateConfig {
	case config.BasicConfigName:
		err = cls.basicConfig.ReadConfig()
	default:
		err = cls.directorConfig.ReadConfig()
	}
	if err != nil {
		log.Default().Error("fetch config fail",
			zap.Error(err),
		)
		return
	}
	// 如果是更新了基本配置，则需要更新缓存
	if updateConfig == config.BasicConfigName {
		cls.dsp = cls.newDispatcher()
	}
	ins := cls.newInstance()
	oldIns := cls.SetInstance(ins)
	// 要清除原有实例
	if oldIns != nil {
		oldIns.Destroy()
	}
	done = true
	return
}

// Watch watch config file change
func (cls *Cluster) Watch() {
	go cls.basicConfig.OnConfigChange(func() {
		cls.ToggleInstance(config.BasicConfigName)
	})
	go cls.directorConfig.OnConfigChange(func() {
		cls.ToggleInstance(config.DirectorConfigName)
	})
}

// ServeHTTP http serve
func (cls *Cluster) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	ins := cls.GetInstance()
	if ins == nil {
		resp.WriteHeader(500)
		resp.Write([]byte("No instance is avaliable"))
		return
	}
	ins.Server.ServeHTTP(resp, req)
}

// NewCluster create a new cluster
func NewCluster() (cluster *Cluster, err error) {

	configURI := os.Getenv("CONFIG")

	basicConfig, err := config.NewBasicConfig(configURI)
	if err != nil {
		return
	}
	directorConfig, err := config.NewDirectorConfig(configURI)
	if err != nil {
		return
	}

	err = basicConfig.ReadConfig()
	if err != nil {
		return
	}
	err = directorConfig.ReadConfig()
	if err != nil {
		return
	}

	cluster = &Cluster{
		basicConfig:    basicConfig,
		directorConfig: directorConfig,
	}
	cluster.dsp = cluster.newDispatcher()
	cluster.SetInstance(cluster.newInstance())
	cluster.Watch()

	return
}
