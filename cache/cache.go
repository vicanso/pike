package cache

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/vicanso/pike/vars"
)

var rsMap = make(map[string]*RequestStatus)
var rsMutex = sync.Mutex{}

var client *badger.DB

// ResponseData 记录响应数据
type ResponseData struct {
	CreatedAt      uint32
	StatusCode     uint16
	Compress       uint8
	ShouldCompress bool
	TTL            uint32
	Header         []byte
	Body           []byte
}

const (
	createIndex         = 0
	statusCodeIndex     = 4
	compressIndex       = 6
	shouldCompressIndex = 7
	ttlIndex            = 8
	headerLengthIndex   = 12
	headerIndex         = 14
)

// RequestStatus 请求状态
type RequestStatus struct {
	createdAt    uint32
	ttl          uint32
	status       int
	waitingChans []chan int
}

// 初始化请求状态
func initRequestStatus(ttl uint32) *RequestStatus {
	rs := &RequestStatus{
		createdAt: uint32(time.Now().Unix()),
		ttl:       ttl,
	}
	return rs
}

// 判断是否已过期
func isExpired(rs *RequestStatus) bool {
	now := uint32(time.Now().Unix())
	if rs.ttl != 0 && now-rs.createdAt > uint32(rs.ttl) {
		return true
	}
	return false
}

// 将uint16转换为字节
func uint16ToBytes(v uint16) []byte {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, v)
	return buf
}

// 将字节转换为uint16
func bytesToUint16(buf []byte) uint16 {
	return binary.LittleEndian.Uint16(buf)
}

// 将uint32转换为字节
func uint32ToBytes(v uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, v)
	return buf
}

// 将字节转换为uint32
func bytesToUint32(buf []byte) uint32 {
	return binary.LittleEndian.Uint32(buf)
}

// trimHeader 将无用的头属性删除（如Date Connection等）
func trimHeader(header []byte) []byte {
	arr := bytes.Split(header, vars.LineBreak)
	data := make([][]byte, 0, len(arr))
	ignoreList := []string{
		"date",
		"connection",
	}
	for _, item := range arr {
		index := bytes.IndexByte(item, vars.Colon)
		if index == -1 {
			continue
		}
		k := strings.ToLower(string(item[:index]))
		found := false
		for _, ignore := range ignoreList {
			if found {
				break
			}
			if k == ignore {
				found = true
			}
		}
		if found {
			continue
		}
		data = append(data, item)
	}
	return bytes.Join(data, vars.LineBreak)
}

// Size 获取缓存记录的总数
func Size() int {
	return len(rsMap)
}

// DataSize 获取数据大小
func DataSize() (int, int) {
	if client == nil {
		return -1, -1
	}
	lsm, vLog := client.Size()
	mb := int64(1024 * 1024)
	return int(lsm / mb), int(vLog / mb)
}

// Stats 获取请求状态的统计
func Stats() (int, int, int, int) {
	fetchingCount := 0
	waitingCount := 0
	cacheableCount := 0
	hitForPassCount := 0
	rsMutex.Lock()
	for _, v := range rsMap {
		switch v.status {
		case vars.Fetching:
			fetchingCount++
			waitingCount += len(v.waitingChans)
		case vars.HitForPass:
			hitForPassCount++
		case vars.Cacheable:
			cacheableCount++
		}
	}
	rsMutex.Unlock()
	return fetchingCount, waitingCount, cacheableCount, hitForPassCount
}

// GetCachedList 获取已缓存的请求列表
func GetCachedList() []byte {
	rsMutex.Lock()
	defer rsMutex.Unlock()
	type cacheData struct {
		Key       string `json:"key"`
		TTL       uint32 `json:"ttl"`
		CreatedAt uint32 `json:"createdAt"`
	}
	cacheDatas := make([]*cacheData, 0)
	for key, v := range rsMap {
		if v.status != vars.Cacheable {
			continue
		}
		cacheDatas = append(cacheDatas, &cacheData{
			Key:       key,
			TTL:       v.ttl,
			CreatedAt: v.createdAt,
		})
	}
	data, _ := json.Marshal(cacheDatas)
	return data
}

// GetRequestStatus 获取请求的状态
func GetRequestStatus(key []byte) (int, chan int) {
	rsMutex.Lock()
	var c chan int
	k := string(key)
	rs := rsMap[k]
	status := vars.Fetching
	if rs == nil || isExpired(rs) {
		status = vars.Fetching
		rs = initRequestStatus(0)
		rsMap[k] = rs
		rs.status = status
	} else if rs.status == vars.Fetching {
		status = vars.Waiting
		c = make(chan int, 1)
		rs.waitingChans = append(rs.waitingChans, c)
	} else {
		status = rs.status
	}
	rsMutex.Unlock()
	return status, c
}

// triggerWatingRequstAndSetStatus 获取等待中的请求，并设置状态和有效期
func triggerWatingRequstAndSetStatus(key []byte, status int, ttl uint32) {
	rsMutex.Lock()
	defer rsMutex.Unlock()
	k := string(key)
	rs := rsMap[k]
	if rs == nil {
		return
	}
	rs.status = status
	rs.ttl = ttl
	waitingChans := rs.waitingChans
	for _, c := range waitingChans {
		c <- status
		close(c)
	}
	rs.waitingChans = nil
}

// HitForPass 触发等待中的请求，并设置状态为hit for pass
func HitForPass(key []byte, ttl uint32) {
	triggerWatingRequstAndSetStatus(key, vars.HitForPass, ttl)
}

// Cacheable 触发等待中的请求，并设置状态为 cacheable
func Cacheable(key []byte, ttl uint32) {
	triggerWatingRequstAndSetStatus(key, vars.Cacheable, ttl)
}

// InitDB 初始化db
func InitDB(dbPath string) (*badger.DB, error) {
	if client != nil {
		return client, nil
	}

	opts := badger.DefaultOptions
	opts.Dir = dbPath
	opts.ValueDir = dbPath
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	client = db
	return db, nil
}

// SaveResponseData 保存Response
func SaveResponseData(key []byte, respData *ResponseData) error {
	// 前四个字节保存创建时间
	// 接着后面两个字节保存ttl
	// 接着后面两个字节保存header的长度
	// 接着是header
	// 最后才是body
	createdAt := respData.CreatedAt
	if createdAt == 0 {
		createdAt = uint32(time.Now().Unix())
	}
	header := trimHeader(respData.Header)
	ttl := respData.TTL
	var shouldCompressData uint8
	if respData.ShouldCompress {
		shouldCompressData = 1
	}
	s := [][]byte{
		uint32ToBytes(createdAt),
		uint16ToBytes(respData.StatusCode),
		[]byte{respData.Compress},
		[]byte{shouldCompressData},
		uint32ToBytes(respData.TTL),
		uint16ToBytes(uint16(len(header))),
		header,
		respData.Body,
	}
	data := bytes.Join(s, []byte(""))
	return Save(key, data, ttl)
}

// GetResponse 获取response
func GetResponse(key []byte) (*ResponseData, error) {
	data, err := Get(key)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}

	headerLength := bytesToUint16(data[headerLengthIndex:headerIndex])
	resData := &ResponseData{
		CreatedAt:  bytesToUint32(data[createIndex:statusCodeIndex]),
		StatusCode: bytesToUint16(data[statusCodeIndex:compressIndex]),
		Compress:   data[compressIndex],
		TTL:        bytesToUint32(data[ttlIndex:headerLengthIndex]),
		Header:     data[headerIndex : headerIndex+headerLength],
		Body:       data[headerIndex+headerLength:],
	}
	if data[shouldCompressIndex] == 1 {
		resData.ShouldCompress = true
	}

	return resData, nil
}

// Save 保存数据
func Save(key, buf []byte, ttl uint32) error {
	if client == nil {
		return vars.ErrDbNotInit
	}
	var stale uint64 = 5
	return client.Update(func(tx *badger.Txn) error {
		return tx.SetEntry(&badger.Entry{
			Key:   key,
			Value: buf,
			// 缓存的数据延期5秒过期
			ExpiresAt: uint64(time.Now().Unix()) + uint64(ttl) + stale,
		})
	})
}

// Get 获取数据
func Get(key []byte) ([]byte, error) {
	if client == nil {
		return nil, vars.ErrDbNotInit
	}
	var buf []byte

	err := client.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		val, err := item.Value()
		if err != nil {
			return err
		}
		buf = val
		return nil
	})

	return buf, err
}

// ClearExpired 清除过期数据数据
func ClearExpired() error {
	rsMutex.Lock()
	now := uint32(time.Now().Unix())
	for k, v := range rsMap {
		if v.createdAt+v.ttl < now {
			delete(rsMap, k)
		}
	}
	rsMutex.Unlock()

	if client == nil {
		return vars.ErrDbNotInit
	}
	err := client.PurgeOlderVersions()
	if err != nil {
		return err
	}
	return client.RunValueLogGC(0.5)
}

// Close 关闭数据库
func Close() error {
	if client == nil {
		return vars.ErrDbNotInit
	}
	return client.Close()
}
