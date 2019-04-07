package config

import (
	"net/http"
	"regexp"

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

// GetCacheZoneSize get cache zone size
func GetCacheZoneSize() int {
	v := viper.GetInt("cache.zone")
	if v <= 0 {
		return 1024
	}
	return v
}

// GetCacheSize get cache size
func GetCacheSize() int {
	v := viper.GetInt("cache.size")
	if v <= 0 {
		return 1024
	}
	return v
}

// GetHitForPassTTL get hit for pass ttl
func GetHitForPassTTL() int {
	v := viper.GetInt("cache.hitForPass")
	if v <= 0 {
		return 600
	}
	return v
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
	v := viper.GetInt("compress.minLength")
	if v <= 0 {
		return 1024
	}
	return v
}

// GetTextFilter get text filter
func GetTextFilter() *regexp.Regexp {
	v := viper.GetString("compress.filter")
	if v == "" {
		v = "text|javascript|json"
	}
	return regexp.MustCompile(v)
}
