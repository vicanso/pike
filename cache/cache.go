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
	CreatedAt uint32
	// HTTP状态码
	StatusCode uint16
	// 数据是否压缩
	Compress uint8
	// 数据是否应该压缩
	ShouldCompress bool
	// 缓存有效时间
	TTL uint32
	// HTTP响应头
	Header []byte
	// HTTP响应数据
	Body []byte
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
	createdAt uint32
	ttl       uint32
	// 请求状态 fetching hitForPass 等
	status int
	// 如果此请求为fetching，则此时相同的请求会写入一个chan
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
	// 需要清除的http头
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
		// 需要忽略的http头
		if found {
			continue
		}
		data = append(data, item)
	}
	return bytes.Join(data, vars.LineBreak)
}

// GetRequestStatus 获取请求的状态
func GetRequestStatus(key []byte) (int, chan int) {
	rsMutex.Lock()
	defer rsMutex.Unlock()
	var c chan int
	k := string(key)
	rs := rsMap[k]
	status := vars.Fetching
	// 如果该key对应的状态为空或者已过期
	if rs == nil || isExpired(rs) {
		status = vars.Fetching
		rs = initRequestStatus(0)
		rsMap[k] = rs
		rs.status = status
	} else if rs.status == vars.Fetching {
		// 如果该key对应的请求正在处理中，添加chan
		status = vars.Waiting
		c = make(chan int, 1)
		rs.waitingChans = append(rs.waitingChans, c)
	} else {
		// hit for pass 或者 cacheable
		status = rs.status
	}
	return status, c
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
	defer rsMutex.Unlock()
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
	now := uint32(time.Now().Unix())
	for key, v := range rsMap {
		// 对于非已缓存的忽略
		if v.status != vars.Cacheable || v.createdAt+v.ttl < now {
			continue
		}
		// 保存缓存的记录
		cacheDatas = append(cacheDatas, &cacheData{
			Key:       key,
			TTL:       v.ttl,
			CreatedAt: v.createdAt,
		})
	}
	data, _ := json.Marshal(cacheDatas)
	return data
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
	// 对所有等待中的请求触发channel
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

// Expire 将key对应的状态设置为过期
func Expire(key []byte) {
	rsMutex.Lock()
	defer rsMutex.Unlock()
	k := string(key)
	rs := rsMap[k]
	if rs == nil {
		return
	}
	rs.createdAt = 0
}

// InitDB 初始化db
func InitDB(dbPath string) (*badger.DB, error) {
	if client != nil {
		return client, nil
	}

	opts := badger.DefaultOptions
	// 暂时未分开两个目录，如果需要更高的性能，可以考虑再调整
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
	// 将要保存的数据转换为bytes
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
	// 因为数据的缓存比rs map的更晚删除，因为肯定有数据，无需要对data检测
	headerLength := bytesToUint16(data[headerLengthIndex:headerIndex])
	// 将bytes转换为ResponseData
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
	// 缓存延期删除，为了避免在判断请求可读取缓存时，刚好缓存过期
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
	// 从数据库中读取数据
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
	defer rsMutex.Unlock()
	now := uint32(time.Now().Unix())
	// 对保存请求状态的map清除
	for k, v := range rsMap {
		if v.createdAt+v.ttl < now {
			delete(rsMap, k)
		}
	}

	if client == nil {
		return vars.ErrDbNotInit
	}
	// 清除旧版数据
	err := client.PurgeOlderVersions()
	if err != nil {
		return err
	}
	// 清除日志数据
	return client.RunValueLogGC(0.5)
}

// Close 关闭数据库
func Close() error {
	if client == nil {
		return vars.ErrDbNotInit
	}
	return client.Close()
}
