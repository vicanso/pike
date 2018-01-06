package server

import "sort"

// TODO 是否将黑名单保存

// BlackIP 黑名单IP
// 黑名单IP的操作由管理员操作，并不常调用
// 因此没有锁的处理
type BlackIP struct {
	IPList []string `json:"ipList"`
}

// Add 添加黑名单IP
func (b *BlackIP) Add(ip string) {
	index := b.FindIndex(ip)
	if index == -1 {
		b.IPList = append(b.IPList, ip)
		sort.Strings(b.IPList)
	}
}

// FindIndex 获取该IP所在的index
func (b *BlackIP) FindIndex(ip string) int {
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
func (b *BlackIP) Remove(ip string) {
	index := b.FindIndex(ip)
	if index != -1 {
		ipList := b.IPList
		b.IPList = append(ipList[:index], ipList[index+1:]...)
	}
}
