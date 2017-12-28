package director

import (
	"bytes"
)

// Config 服务器配置列表
type Config struct {
	Name     string
	Type     string
	Ping     string
	Prefix   []string
	Host     []string
	Pass     []string
	Backends []string
}

// Director 服务器列表
type Director struct {
	Name     string
	Policy   string
	Ping     string
	Prefixs  [][]byte
	Hosts    [][]byte
	Passes   [][]byte
	Backends []string
	Priority int
}

// Match 判断该director是否符合
func (d *Director) Match(host, uri []byte) bool {
	hosts := d.Hosts
	prefixs := d.Prefixs
	if hosts == nil && prefixs == nil {
		return true
	}
	match := false
	if hosts != nil {
		for _, item := range hosts {
			if !match && bytes.Compare(host, item) == 0 {
				match = true
			}
		}
		// 如果host不匹配，直接返回
		if !match {
			return match
		}
	}
	if prefixs != nil {
		// 重置match状态，判断prefix
		match = false
		for _, item := range prefixs {
			if !match && bytes.Contains(uri, item) {
				match = true
			}
		}
	}
	return match
}

func strListToByteList(original []string) [][]byte {
	items := make([][]byte, len(original))
	for index, item := range original {
		items[index] = []byte(item)
	}
	return items
}

// CreateDirector 创建一个director
func CreateDirector(config *Config) *Director {
	d := &Director{
		Name:     config.Name,
		Policy:   config.Type,
		Ping:     config.Ping,
		Backends: config.Backends,
	}
	priority := 8
	if len(config.Prefix) != 0 {
		priority -= 2
		d.Prefixs = strListToByteList(config.Prefix)
	}
	if len(config.Host) != 0 {
		priority -= 4
		d.Hosts = strListToByteList(config.Host)
	}
	if len(config.Pass) != 0 {
		d.Passes = strListToByteList(config.Pass)
	}
	d.Priority = priority
	return d
}
