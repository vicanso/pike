package director

import (
	"bytes"
	"sort"

	"github.com/vicanso/pike/config"
)

// Director 服务器列表
type Director struct {
	// 名称
	Name string `json:"name"`
	// 选择backend的方式
	Policy string `json:"policy"`
	// 健康检测的url
	Ping string `json:"ping"`
	// 对应的url前缀
	Prefixs [][]byte `json:"prefixs"`
	// 对应的host
	Hosts [][]byte `json:"hosts"`
	// 设置直接pass的url规则
	Passes [][]byte `json:"passes"`
	// 后端服务列表
	Backends []string `json:"backends"`
	// 优先级（根据host prefix计算），低的优先选择
	Priority int `json:"priority"`
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
	// 判断host是否符合
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
	// 判断prefix是否符合
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

// []string 转换为 [][]byte
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
	// 如果有配置host，优先前提升4
	if len(directorConfig.Host) != 0 {
		priority -= 4
		d.Hosts = strListToByteList(directorConfig.Host)
	}
	// 如果有配置prefix，优先级提升2
	if len(directorConfig.Prefix) != 0 {
		priority -= 2
		d.Prefixs = strListToByteList(directorConfig.Prefix)
	}
	// 如果配置子pass，生成pass列表
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
