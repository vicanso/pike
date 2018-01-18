package server

import (
	"sort"
	"strings"

	"../cache"
)

// TODO 是否将黑名单保存

// BlockIP 黑名单IP
// 黑名单IP的操作由管理员操作，并不常调用
// 因此没有锁的处理
type BlockIP struct {
	IPList []string `json:"ipList"`
}

// 保存到bucket中对应的key
var blockIPKey = []byte("config-blockIP")

// Add 添加黑名单IP
func (b *BlockIP) Add(ip string) {
	index := b.FindIndex(ip)
	if index == -1 {
		b.IPList = append(b.IPList, ip)
		sort.Strings(b.IPList)
		data := []byte(strings.Join(b.IPList, ","))
		cache.Save(blockIPKey, data, 365*24*3600)
	}
}

// FindIndex 获取该IP所在的index
func (b *BlockIP) FindIndex(ip string) int {
	ipList := b.IPList
	i := sort.Search(len(ipList), func(i int) bool {
		return ipList[i] >= ip
	})
	if i >= len(ipList) || ipList[i] != ip {
		return -1
	}
	return i
}

// Remove 删除黑名单IP
func (b *BlockIP) Remove(ip string) {
	index := b.FindIndex(ip)
	if index != -1 {
		ipList := b.IPList
		b.IPList = append(ipList[:index], ipList[index+1:]...)
	}
}

// InitFromCache 从缓存中初始化黑名单IP
func (b *BlockIP) InitFromCache() {
	data, err := cache.Get(blockIPKey)
	if err != nil {
		return
	}
	ips := strings.Split(string(data), ",")
	b.IPList = append(b.IPList, ips...)
}
