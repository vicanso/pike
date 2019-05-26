package config

import (
	"os"
	"sort"
	"strings"
	"time"

	"github.com/go-yaml/yaml"
	"github.com/spf13/viper"
	"github.com/vicanso/pike/df"
)

const (
	// FileType file type
	FileType = iota
)

const (
	defaultConfigName = "config"
	defaultConfigType = "yml"
)

const (
	identityKey              = "identity"
	headerKey                = "header"
	requestHeaderKey         = "request_header"
	concurrencyKey           = "concurrency"
	enableServerTimingKey    = "enable_server_timing"
	cacheZoneKey             = "cache.zone"
	cacheSizeKey             = "cache.size"
	hitForPassKey            = "cache.hit_for_pass"
	compressLevelKey         = "compress.level"
	compressMinLengthKey     = "compress.min_length"
	compressFilterKey        = "compress.filter"
	timeoutIdleConnKey       = "timeout.idle_conn"
	timeoutExpectContinueKey = "timeout.expect_continue"
	timeoutResponseHeaderKey = "timeout.response_header"
	timeoutConnectKey        = "timeout.connect"
	timeoutTLSHandshakeKey   = "timeout.tls_handshake"
	adminPrefixKey           = "admin.prefix"
	adminUserKey             = "admin.user"
	adminPasswordKey         = "admin.password"
)

const (
	backendPolicyKey        = "policy"
	backendPingKey          = "ping"
	backendRequestHeaderKey = "request_header"
	backendHeaderKey        = "header"
	backendPrefixsKey       = "prefixs"
	backendHostsKey         = "hosts"
	backendBackendsKey      = "backends"
	backendRewritesKey      = "rewrites"
)

type (
	// Config config struct
	Config struct {
		modified bool
		Viper    *viper.Viper
		// Name config's name
		Name string
		// Type config's type
		Type int
	}
	// Backend backend config
	Backend struct {
		Name          string
		Policy        string
		Ping          string
		RequestHeader []string `yaml:"requestHeader"`
		Header        []string
		Prefixs       []string
		Hosts         []string
		Backends      []string
		Rewrites      []string
	}
)

// IsTest is test mode
func IsTest() bool {
	return os.Getenv("GO_MODE") == "test"
}

// New create a config instance
func New() *Config {
	return NewFileConfig(defaultConfigName)
}

// NewFileConfig create a file config instance
func NewFileConfig(name string) *Config {
	v := viper.New()
	v.SetConfigType(defaultConfigType)
	for _, value := range df.ConfigPathList {
		v.AddConfigPath(value)
	}
	c := &Config{
		Viper: v,
		Name:  name,
	}
	return c
}

// Fetch fetch config
func (c *Config) Fetch() error {
	if c.Type == FileType {
		return c.readInConfig()
	}
	return nil
}

// WriteConfig write config
func (c *Config) WriteConfig() (err error) {
	if c.Type == FileType {
		err = c.Viper.WriteConfig()
		_, ok := err.(viper.ConfigFileNotFoundError)
		// 如果是找不到配置文件，则先创建
		if ok {
			file := df.ConfigPathList[0] + "/" + c.Name + "." + defaultConfigType
			_, err = os.Stat(file)
			if os.IsNotExist(err) {
				_, err = os.Create(file)
			}
			if err != nil {
				return
			}
			err = c.Viper.WriteConfig()
		}
		return
	}
	return
}

// readInConfig read config from file
func (c *Config) readInConfig() (err error) {
	v := c.Viper
	v.SetConfigName(c.Name)
	err = v.ReadInConfig()
	_, ok := err.(viper.ConfigFileNotFoundError)
	// 如果是找不到配置文件，则不需抛出异常
	if ok {
		err = nil
	}
	return
}

func (c *Config) set(key string, value interface{}) {
	c.modified = true
	c.Viper.Set(key, value)
}

// GetIdentity get identity
func (c *Config) GetIdentity() string {
	return c.Viper.GetString(identityKey)
}

// SetIdentity set identity
func (c *Config) SetIdentity(value string) {
	c.set(identityKey, value)
}

// GetHeader get response's header
func (c *Config) GetHeader() []string {
	return c.Viper.GetStringSlice(headerKey)
}

// SetHeader set response's header
func (c *Config) SetHeader(value []string) {
	c.set(headerKey, value)
}

// GetRequestHeader get request's header
func (c *Config) GetRequestHeader() []string {
	return c.Viper.GetStringSlice(requestHeaderKey)
}

// SetRequestHeader set request's header
func (c *Config) SetRequestHeader(value []string) {
	c.set(requestHeaderKey, value)
}

// GetConcurrency get concurrency limit
func (c *Config) GetConcurrency() uint32 {
	v := c.Viper.GetInt32(concurrencyKey)
	if v <= 0 {
		return 256 * 1000
	}
	return uint32(v)
}

// SetConcurrency set concurrency limit
func (c *Config) SetConcurrency(value uint32) {
	c.set(concurrencyKey, value)
}

// GetEnableServerTiming get enable server timing flag
func (c *Config) GetEnableServerTiming() bool {
	return c.Viper.GetBool(enableServerTimingKey)
}

// SetEnableServerTiming set enable server timing flag
func (c *Config) SetEnableServerTiming(value bool) {
	c.set(enableServerTimingKey, value)
}

// getIntDefault get int default
func (c *Config) getIntDefault(key string, defaultInt int) int {
	v := c.Viper.GetInt(key)
	if v <= 0 {
		return defaultInt
	}
	return v
}

// GetCacheZoneSize get cache zone size
func (c *Config) GetCacheZoneSize() int {
	return c.getIntDefault(cacheZoneKey, 1024)
}

// SetCacheZoneSize set cache zone size
func (c *Config) SetCacheZoneSize(value int) {
	c.set(cacheZoneKey, value)
}

// GetCacheSize get cache size
func (c *Config) GetCacheSize() int {
	return c.getIntDefault(cacheSizeKey, 1024)
}

// SetCacheSzie set cache size
func (c *Config) SetCacheSzie(value int) {
	c.set(cacheSizeKey, value)
}

// GetHitForPassTTL get hit for pass ttl
func (c *Config) GetHitForPassTTL() int {
	return c.getIntDefault(hitForPassKey, 300)
}

// SetHitForPassTTL set hit for pass ttl
func (c *Config) SetHitForPassTTL(value int) {
	c.set(hitForPassKey, value)
}

// GetCompressLevel get compress level
func (c *Config) GetCompressLevel() int {
	v := c.Viper.GetInt(compressLevelKey)
	if v < 0 {
		return 0
	}
	return v
}

// SetCompressLevel set compress level
func (c *Config) SetCompressLevel(value int) {
	c.set(compressLevelKey, value)
}

// GetCompressMinLength get compress min length
func (c *Config) GetCompressMinLength() int {
	return c.getIntDefault(compressMinLengthKey, 1024)
}

// SetCompressMinLength set compress min length
func (c *Config) SetCompressMinLength(value int) {
	c.set(compressMinLengthKey, value)
}

// GetTextFilter get text filter
func (c *Config) GetTextFilter() string {
	v := c.Viper.GetString(compressFilterKey)
	if v == "" {
		v = "text|javascript|json"
	}
	return v
}

// SetTextFilter set text filter
func (c *Config) SetTextFilter(value string) {
	c.set(compressFilterKey, value)
}

func (c *Config) getDurationDefault(key string, defaultDuration time.Duration) time.Duration {
	v := c.Viper.GetDuration(key)
	if v == 0 {
		return defaultDuration
	}
	return v
}

// GetIdleConnTimeout get idle conn timeout
func (c *Config) GetIdleConnTimeout() time.Duration {
	return c.getDurationDefault(timeoutIdleConnKey, 90*time.Second)
}

// SetIdleConnTimeout set idle conn timeout
func (c *Config) SetIdleConnTimeout(value time.Duration) {
	c.set(timeoutIdleConnKey, value)
}

// GetExpectContinueTimeout get expect continue timeout
func (c *Config) GetExpectContinueTimeout() time.Duration {
	return c.getDurationDefault(timeoutExpectContinueKey, 1*time.Second)
}

// SetExpectContinueTimeout set expect continue timeout
func (c *Config) SetExpectContinueTimeout(value time.Duration) {
	c.set(timeoutExpectContinueKey, value)
}

// GetResponseHeaderTimeout get response header timeout
func (c *Config) GetResponseHeaderTimeout() time.Duration {
	return c.getDurationDefault(timeoutResponseHeaderKey, 10*time.Second)
}

// SetResponseHeaderTimeout set response header timeout
func (c *Config) SetResponseHeaderTimeout(value time.Duration) {
	c.set(timeoutResponseHeaderKey, value)
}

// GetConnectTimeout get connect timeout
func (c *Config) GetConnectTimeout() time.Duration {
	return c.getDurationDefault(timeoutConnectKey, 15*time.Second)
}

// SetConnectTimeout set connect timeout
func (c *Config) SetConnectTimeout(value time.Duration) {
	c.set(timeoutConnectKey, value)
}

// GetTLSHandshakeTimeout get tls hand shake timeout
func (c *Config) GetTLSHandshakeTimeout() time.Duration {
	return c.getDurationDefault(timeoutTLSHandshakeKey, 5*time.Second)
}

// SetTLSHandshakeTimeout set tls handshake timeout
func (c *Config) SetTLSHandshakeTimeout(value time.Duration) {
	c.set(timeoutTLSHandshakeKey, value)
}

// GetAdminPath get admin path
func (c *Config) GetAdminPath() string {
	return c.Viper.GetString(adminPrefixKey)
}

// SetAdminPath set admin path
func (c *Config) SetAdminPath(value string) {
	c.set(adminPrefixKey, value)
}

// GetAdminUser get admin user
func (c *Config) GetAdminUser() string {
	return c.Viper.GetString(adminUserKey)
}

// SetAdminUser set admin user
func (c *Config) SetAdminUser(value string) {
	c.set(adminUserKey, value)
}

// GetAdminPassword get admin password
func (c *Config) GetAdminPassword() string {
	return c.Viper.GetString(adminPasswordKey)
}

// SetAdminPassword set admin password
func (c *Config) SetAdminPassword(value string) {
	c.set(adminPasswordKey, value)
}

// GetBackends get backends
func (c *Config) GetBackends() []Backend {
	vp := c.Viper
	backends := make([]Backend, 0)
	keys := vp.AllKeys()
	nameList := []string{}
	for _, key := range keys {
		name := strings.Split(key, ".")[0]
		found := false
		for _, item := range nameList {
			if item == name {
				found = true
			}
		}
		if !found {
			nameList = append(nameList, name)
		}
	}
	sort.Sort(sort.StringSlice(nameList))
	fn := func(name string) Backend {
		return Backend{
			Name:          name,
			Policy:        vp.GetString(name + "." + backendPolicyKey),
			Ping:          vp.GetString(name + "." + backendPingKey),
			RequestHeader: vp.GetStringSlice(name + "." + backendRequestHeaderKey),
			Header:        vp.GetStringSlice(name + "." + backendHeaderKey),
			Prefixs:       vp.GetStringSlice(name + "." + backendPrefixsKey),
			Hosts:         vp.GetStringSlice(name + "." + backendHostsKey),
			Backends:      vp.GetStringSlice(name + "." + backendBackendsKey),
			Rewrites:      vp.GetStringSlice(name + "." + backendRewritesKey),
		}
	}
	for _, name := range nameList {
		backends = append(backends, fn(name))
	}
	return backends
}

// SetBackend set backend
func (c *Config) SetBackend(backed Backend) {
	name := backed.Name
	if name == "" {
		return
	}
	c.set(name+"."+backendPolicyKey, backed.Policy)
	c.set(name+"."+backendPingKey, backed.Ping)
	c.set(name+"."+backendRequestHeaderKey, backed.RequestHeader)
	c.set(name+"."+backendHeaderKey, backed.Header)
	c.set(name+"."+backendPrefixsKey, backed.Prefixs)
	c.set(name+"."+backendHostsKey, backed.Hosts)
	c.set(name+"."+backendBackendsKey, backed.Backends)
	c.set(name+"."+backendRewritesKey, backed.Rewrites)
}

// ToYAML config to yaml
func (c *Config) ToYAML() ([]byte, error) {
	settings := c.Viper.AllSettings()
	return yaml.Marshal(settings)
}
