package config

import (
	"os"
	"time"

	"github.com/vicanso/hes"

	"github.com/go-yaml/yaml"
)

const (
	// FileType file type
	FileType = iota
)

const (
	basicConfigFile    = "config"
	directorConfigFile = "director"
	configType         = "yml"

	defaultConcurrency           = 256 * 1000
	defaultZoneSize              = 1024
	defaultCacheSize             = 1024
	defaultHitForPass            = 5 * time.Minute
	defaultCompressMinLength     = 1024
	defaultCompressFilter        = "text|javascript|json"
	defaultIdelConnTimeout       = 90 * time.Second
	defaultExpectContinueTimeout = 3 * time.Second
	defaultResponseHeaderTimeout = 10 * time.Second
	defaultConnectTimeout        = 15 * time.Second
	defaultTLSHandshakeTimeout   = 5 * time.Second
)

type (
	// BasicConfig basic config data
	BasicConfig struct {
		Admin struct {
			Prefix   string `yaml:"prefix,omitempty" json:"prefix,omitempty"`
			User     string `yaml:"user,omitempty" json:"user,omitempty"`
			Password string `yaml:"password,omitempty" json:"password,omitempty"`
		} `yaml:"admin,omitempty" json:"admin,omitempty"`
		Concurrency        int      `yaml:"concurrency,omitempty" json:"concurrency,omitempty"`
		EnableServerTiming bool     `yaml:"enable_server_timing,omitempty" json:"enableServerTiming,omitempty"`
		Identity           string   `yaml:"identity,omitempty" json:"identity,omitempty"`
		ResponseHeader     []string `yaml:"response_header,omitempty" json:"responseHeader,omitempty"`
		RequestHeader      []string `yaml:"request_header,omitempty" json:"requestHeader,omitempty"`
		Compress           struct {
			Level     int    `yaml:"level,omitempty" json:"level,omitempty"`
			MinLength int    `yaml:"min_length,omitempty" json:"minLength,omitempty"`
			Filter    string `yaml:"filter,omitempty" json:"filter,omitempty"`
		} `yaml:"compress,omitempty" json:"compress,omitempty"`
		Cache struct {
			Zone       int           `yaml:"zone,omitempty" json:"zone,omitempty"`
			Size       int           `yaml:"size,omitempty" json:"size,omitempty"`
			HitForPass time.Duration `yaml:"hit_for_pass,omitempty" json:"hitForPass,omitempty"`
		} `yaml:"cache,omitempty" json:"cache,omitempty"`
		Timeout struct {
			IdleConn       time.Duration `yaml:"idle_conn,omitempty" json:"idleConn,omitempty"`
			ExpectContinue time.Duration `yaml:"expect_continue,omitempty" json:"expectContinue,omitempty"`
			ResponseHeader time.Duration `yaml:"response_header,omitempty" json:"responseHeader,omitempty"`
			Connect        time.Duration `yaml:"connect,omitempty" json:"connect,omitempty"`
			TLSHandshake   time.Duration `yaml:"tls_handshake,omitempty" json:"tlsHandshake,omitempty"`
		} `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	}
	// BackendConfig backend config
	BackendConfig struct {
		Name          string   `yaml:"-" json:"name,omitempty"`
		Policy        string   `yaml:"policy,omitempty" json:"policy,omitempty"`
		Ping          string   `yaml:"ping,omitempty" json:"ping,omitempty"`
		Prefixs       []string `yaml:"prefixs,omitempty" json:"prefixs,omitempty"`
		Rewrites      []string `yaml:"rewrites,omitempty" json:"rewrites,omitempty"`
		Hosts         []string `yaml:"hosts,omitempty" json:"hosts,omitempty"`
		Header        []string `yaml:"header,omitempty" json:"header,omitempty"`
		RequestHeader []string `yaml:"request_header,omitempty" json:"requestHeader,omitempty"`
		Backends      []string `yaml:"backends,omitempty" json:"backends,omitempty"`
	}
	// BackendConfigs upstream backend config data
	BackendConfigs map[string]BackendConfig
	// Config basic config of pike
	Config struct {
		Data BasicConfig
		rw   ReadWriter
	}
	// DirectorConfig director config
	DirectorConfig struct {
		Data BackendConfigs
		rw   ReadWriter
	}
	// ReadWriter config reader writer
	ReadWriter interface {
		ReadConfig() ([]byte, error)
		WriteConfig([]byte) error
		Watch(func())
	}
)

func readConfig(rw ReadWriter, v interface{}) (err error) {
	buf, err := rw.ReadConfig()
	// 配置文件
	if err != nil && os.IsNotExist(err) {
		err = nil
	}
	if err != nil {
		return
	}
	err = yaml.Unmarshal(buf, v)
	return
}

func writeConfig(rw ReadWriter, v interface{}) (err error) {
	buf, err := yaml.Marshal(v)
	if err != nil {
		return
	}
	return rw.WriteConfig(buf)
}

// fillDefault fill the config with default
func (bc *Config) fillDefault() {
	data := &bc.Data
	if data.Concurrency <= 0 {
		data.Concurrency = defaultConcurrency
	}

	cache := &data.Cache
	if cache.Zone <= 0 {
		cache.Zone = defaultZoneSize
	}
	if cache.Size <= 0 {
		cache.Size = defaultCacheSize
	}
	if cache.HitForPass < time.Second {
		cache.HitForPass = defaultHitForPass
	}

	compress := &data.Compress
	if compress.Level < 0 {
		compress.Level = 0
	}
	if compress.MinLength == 0 {
		compress.MinLength = defaultCompressMinLength
	}
	if compress.Filter == "" {
		compress.Filter = defaultCompressFilter
	}

	timeout := &data.Timeout
	if timeout.Connect == 0 {
		timeout.Connect = defaultIdelConnTimeout
	}
	if timeout.ExpectContinue == 0 {
		timeout.ExpectContinue = defaultExpectContinueTimeout
	}
	if timeout.IdleConn == 0 {
		timeout.IdleConn = defaultIdelConnTimeout
	}
	if timeout.ResponseHeader == 0 {
		timeout.ResponseHeader = defaultResponseHeaderTimeout
	}
	if timeout.TLSHandshake == 0 {
		timeout.TLSHandshake = defaultTLSHandshakeTimeout
	}

	admin := &data.Admin
	if admin.Prefix == "" {
		admin.Prefix = os.Getenv("ADMIN_PATH")
	}
}

// ReadConfig read config
func (bc *Config) ReadConfig() (err error) {
	err = readConfig(bc.rw, &bc.Data)
	if err != nil {
		return
	}
	bc.fillDefault()
	return
}

// WriteConfig write config
func (bc *Config) WriteConfig() (err error) {
	return writeConfig(bc.rw, &bc.Data)
}

// OnConfigChange watch config change
func (bc *Config) OnConfigChange(fn func()) {
	bc.rw.Watch(fn)
}

// YAML to yaml
func (bc *Config) YAML() ([]byte, error) {
	return yaml.Marshal(&bc.Data)
}

// ReadConfig read config
func (dc *DirectorConfig) ReadConfig() (err error) {
	err = readConfig(dc.rw, &dc.Data)
	if err != nil {
		return
	}
	return
}

// GetBackends get backends
func (dc *DirectorConfig) GetBackends() []BackendConfig {
	result := make([]BackendConfig, 0)
	for key := range dc.Data {
		value := dc.Data[key]
		value.Name = key
		result = append(result, value)
	}
	return result
}

// YAML to yaml
func (dc *DirectorConfig) YAML() ([]byte, error) {
	return yaml.Marshal(&dc.Data)
}

// WriteConfig write config
func (dc *DirectorConfig) WriteConfig() (err error) {
	return writeConfig(dc.rw, &dc.Data)
}

// OnConfigChange watch config change
func (dc *DirectorConfig) OnConfigChange(fn func()) {
	dc.rw.Watch(fn)
}

// AddBackend add backend
func (dc *DirectorConfig) AddBackend(backend BackendConfig) (err error) {
	_, exists := dc.Data[backend.Name]
	if exists {
		err = hes.New("backend is already exists")
	}
	dc.Data[backend.Name] = backend
	return
}

// UpdateBackend update backend
func (dc *DirectorConfig) UpdateBackend(backend BackendConfig) (err error) {
	_, exists := dc.Data[backend.Name]
	if !exists {
		err = hes.New("backend isn't exists")
	}
	dc.Data[backend.Name] = backend
	return
}

// RemoveBackend remove backend
func (dc *DirectorConfig) RemoveBackend(name string) {

	delete(dc.Data, name)
}

// NewFileConfig new file basic config
func NewFileConfig() *Config {
	return &Config{
		rw: &FileConfig{
			Type: configType,
			Name: basicConfigFile,
		},
	}
}

// NewFileDirectorConfig new director config
func NewFileDirectorConfig() *DirectorConfig {
	return &DirectorConfig{
		rw: &FileConfig{
			Type: configType,
			Name: directorConfigFile,
		},
	}
}
