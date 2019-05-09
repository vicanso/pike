package config

import (
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/spf13/viper"
	"github.com/vicanso/pike/df"
	"github.com/vicanso/pike/util"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	for _, value := range df.ConfigPathList {
		viper.AddConfigPath(value)
	}
	err := viper.ReadInConfig()

	_, ok := err.(viper.ConfigFileNotFoundError)
	// 如果是找不到配置文件，则不需抛出异常
	if ok {
		err = nil
	}

	if err != nil {
		panic(err)
	}
}

// IsTest is test mode
func IsTest() bool {
	return os.Getenv("GO_MODE") == "test"
}

// GetListenAddress get listen address
func GetListenAddress() string {
	addr := viper.GetString("listen")
	if addr == "" {
		return ":3015"
	}
	return addr
}

// GetIdentity get identity
func GetIdentity() string {
	return viper.GetString("identity")
}

// GetHeader get response's header
func GetHeader() http.Header {
	return util.ConvertToHTTPHeader(viper.GetStringSlice("header"))
}

// GetRequestHeader get request's header
func GetRequestHeader() http.Header {
	return util.ConvertToHTTPHeader(viper.GetStringSlice("requestHeader"))
}

// GetConcurrency get concurrency limit
func GetConcurrency() uint32 {
	v := viper.GetInt32("concurrency")
	if v <= 0 {
		return 256 * 1000
	}
	return uint32(v)
}

// IsEnableServerTiming is enable server timing
func IsEnableServerTiming() bool {
	return viper.GetBool("enableServerTiming")
}

// getIntDefault get int default
func getIntDefault(key string, defaultInt int) int {
	v := viper.GetInt(key)
	if v <= 0 {
		return defaultInt
	}
	return v
}

// GetCacheZoneSize get cache zone size
func GetCacheZoneSize() int {
	return getIntDefault("cache.zone", 1024)
}

// GetCacheSize get cache size
func GetCacheSize() int {
	return getIntDefault("cache.size", 1024)
}

// GetHitForPassTTL get hit for pass ttl
func GetHitForPassTTL() int {
	return getIntDefault("cache.hitForPass", 300)
}

// GetCompressLevel get compress level
func GetCompressLevel() int {
	v := viper.GetInt("compress.level")
	if v < 0 {
		return 0
	}
	return v
}

// GetCompressMinLength get compress min length
func GetCompressMinLength() int {
	return getIntDefault("compress.minLength", 1024)
}

// GetTextFilter get text filter
func GetTextFilter() *regexp.Regexp {
	v := viper.GetString("compress.filter")
	if v == "" {
		v = "text|javascript|json"
	}
	return regexp.MustCompile(v)
}

func getDurationDefault(key string, defaultDuration time.Duration) time.Duration {
	v := viper.GetDuration(key)
	if v == 0 {
		return defaultDuration
	}
	return v
}

// GetIdleConnTimeout get idle conn timeout
func GetIdleConnTimeout() time.Duration {
	return getDurationDefault("timeout.idleConn", 90*time.Second)
}

// GetExpectContinueTimeout get expect continue timeout
func GetExpectContinueTimeout() time.Duration {
	return getDurationDefault("timeout.expectContinue", 1*time.Second)
}

// GetResponseHeaderTimeout get response header timeout
func GetResponseHeaderTimeout() time.Duration {
	return getDurationDefault("timeout.responseHeader", 10*time.Second)
}

// GetConnectTimeout get connect timeout
func GetConnectTimeout() time.Duration {
	return getDurationDefault("timeout.connect", 15*time.Second)
}

// GetTLSHandshakeTimeout get tls hand shake timeout
func GetTLSHandshakeTimeout() time.Duration {
	return getDurationDefault("timeout.tlsHandshake", 5*time.Second)
}

// GetAdminPath get admin path
func GetAdminPath() string {
	return viper.GetString("admin.prefix")
}
