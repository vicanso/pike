package director

import (
	"bytes"
	"sort"
	"time"

	"github.com/vicanso/pike/proxy"
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
	Name     string          `json:"name"`
	Policy   string          `json:"policy"`
	Ping     string          `json:"ping"`
	Prefixs  [][]byte        `json:"prefixs"`
	Hosts    [][]byte        `json:"hosts"`
	Passes   [][]byte        `json:"passes"`
	Backends []string        `json:"backends"`
	Priority int             `json:"priority"`
	Upstream *proxy.Upstream `json:"upstream"`
}

// DirectorSlice 用于director排序
type DirectorSlice []*Director

// Len 获取director slice的长度
func (s DirectorSlice) Len() int {
	return len(s)
}

// Swap 元素互换
func (s DirectorSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s DirectorSlice) Less(i, j int) bool {
	return s[i].Priority < s[j].Priority
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
			if match {
				break
			}
			// 如果配置的host以~开头，表示只要包含则符合
			if item[0] == byte('~') {
				if bytes.Contains(host, item[1:]) {
					match = true
				}
			} else if bytes.Compare(host, item) == 0 {
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
			if !match && bytes.HasPrefix(uri, item) {
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
		priority -= 4
		d.Prefixs = strListToByteList(config.Prefix)
	}
	if len(config.Host) != 0 {
		priority -= 2
		d.Hosts = strListToByteList(config.Host)
	}
	if len(config.Pass) != 0 {
		d.Passes = strListToByteList(config.Pass)
	}
	d.Priority = priority
	return d
}

// GetDirectors 获取 directors
func GetDirectors(ds []*Config) DirectorSlice {
	length := len(ds)
	directorList := make(DirectorSlice, 0, length)
	for _, directorConf := range ds {
		d := CreateDirector(directorConf)
		name := d.Name
		up := proxy.CreateUpstream(name, d.Policy, "")
		for _, backend := range d.Backends {
			up.AddBackend(backend)
		}
		up.StartHealthcheck(d.Ping, time.Second)
		d.Upstream = up
		directorList = append(directorList, d)
	}
	sort.Sort(directorList)
	return directorList
}
