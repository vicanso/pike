package director

import (
	"bytes"
	"sort"

	"github.com/vicanso/pike/config"
)

// Director 服务器列表
type Director struct {
	Name     string   `json:"name"`
	Policy   string   `json:"policy"`
	Ping     string   `json:"ping"`
	Prefixs  [][]byte `json:"prefixs"`
	Hosts    [][]byte `json:"hosts"`
	Passes   [][]byte `json:"passes"`
	Backends []string `json:"backends"`
	Priority int      `json:"priority"`
}

// Directors 用于director排序
type Directors []*Director

// 保证director列表
var directorList = make(Directors, 0)

// Len 获取director slice的长度
func (s Directors) Len() int {
	return len(s)
}

// Swap 元素互换
func (s Directors) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s Directors) Less(i, j int) bool {
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

// createDirector 创建一个director
func createDirector(directorConfig *config.Director) *Director {
	d := &Director{
		Name:     directorConfig.Name,
		Policy:   directorConfig.Type,
		Ping:     directorConfig.Ping,
		Backends: directorConfig.Backends,
	}
	priority := 8
	if len(directorConfig.Prefix) != 0 {
		priority -= 4
		d.Prefixs = strListToByteList(directorConfig.Prefix)
	}
	if len(directorConfig.Host) != 0 {
		priority -= 2
		d.Hosts = strListToByteList(directorConfig.Host)
	}
	if len(directorConfig.Pass) != 0 {
		d.Passes = strListToByteList(directorConfig.Pass)
	}
	d.Priority = priority
	return d
}

// Append 添加director
// 在程序启动时做初始化添加，非线程安全，若后期需要动态修改再调整
func Append(directorConfig *config.Director) {
	d := createDirector(directorConfig)
	directorList = append(directorList, d)
	sort.Sort(directorList)
}

// GetMatch 获取匹配的director
func GetMatch(host, uri []byte) *Director {
	var found *Director
	// 查找可用的director
	for _, d := range directorList {
		if found == nil && d.Match(host, uri) {
			found = d
		}
	}
	return found
}

// List 获取所有的director
func List() Directors {
	return directorList
}
