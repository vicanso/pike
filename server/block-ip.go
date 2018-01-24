package server

import (
	"sort"
	"strings"
	"sync"

	"github.com/vicanso/pike/cache"
)

// BlockIP 屏蔽IP
type BlockIP struct {
	// ip列表
	List []string `json:"ipList"`
	// 读写锁
	m *sync.RWMutex
}

// 保存到bucket中对应的key
var blockIPKey = []byte("config-blockIP")

// 获取该ip对应的index
func findIndex(list []string, ip string) int {
	i := sort.Search(len(list), func(i int) bool {
		return list[i] >= ip
	})
	if i >= len(list) || list[i] != ip {
		return -1
	}
	return i
}

// Add 添加黑名单IP
func (b *BlockIP) Add(ip string) {
	b.m.Lock()
	defer b.m.Unlock()
	index := findIndex(b.List, ip)
	if index == -1 {
		b.List = append(b.List, ip)
		sort.Strings(b.List)
		data := []byte(strings.Join(b.List, ","))
		cache.Save(blockIPKey, data, 365*24*3600)
	}
}

// FindIndex 获取该IP所在的index
func (b *BlockIP) FindIndex(ip string) int {
	b.m.RLock()
	defer b.m.RUnlock()
	return findIndex(b.List, ip)
}

// Remove 删除黑名单IP
func (b *BlockIP) Remove(ip string) {
	b.m.Lock()
	defer b.m.Unlock()
	index := findIndex(b.List, ip)
	if index != -1 {
		ipList := b.List
		b.List = append(ipList[:index], ipList[index+1:]...)
	}
}

// InitFromCache 从缓存中初始化黑名单IP
func (b *BlockIP) InitFromCache() {
	b.m.Lock()
	defer b.m.Unlock()
	data, err := cache.Get(blockIPKey)
	if err != nil {
		return
	}
	ips := strings.Split(string(data), ",")
	b.List = append(b.List, ips...)
}
