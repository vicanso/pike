package cache

import (
	"bytes"
	"sync"

	"../util"
	"../vars"
	"github.com/boltdb/bolt"
	"github.com/valyala/fasthttp"
)

var rsMap = make(map[string]*RequestStatus)
var rsMutex = sync.Mutex{}

var client *bolt.DB

// ResponseData 记录响应数据
type ResponseData struct {
	CreatedAt uint32
	TTL       uint16
	Header    []byte
	Body      []byte
}

// RequestStatus 请求状态
type RequestStatus struct {
	createdAt   int64
	ttl         uint16
	status      string
	waitingList []*fasthttp.RequestCtx
}

func initRequestStatus(key string, ttl uint16) *RequestStatus {
	rs := &RequestStatus{
		createdAt: util.GetSeconds(),
		ttl:       ttl,
	}
	rsMap[key] = rs
	return rs
}

func isExpired(rs *RequestStatus) bool {
	if rs.ttl != 0 && util.GetSeconds()-rs.createdAt > int64(rs.ttl) {
		return true
	}
	return false
}

// GetRequestStatus 获取请求的状态
func GetRequestStatus(ctx *fasthttp.RequestCtx) string {
	rsMutex.Lock()
	defer rsMutex.Unlock()
	key := string(util.GenRequestKey(ctx))
	rs := rsMap[key]
	status := ""
	if rs == nil || isExpired(rs) {
		status = vars.Fetching
		rs = initRequestStatus(key, 0)
		rs.status = status
	} else if rs.status == vars.Fetching {
		status = vars.Waiting
		rs.waitingList = append(rs.waitingList, ctx)
	} else {
		status = rs.status
	}
	return status
}

// GetWaitingRequests 获取等待中的请求，并重置
func GetWaitingRequests(key []byte) []*fasthttp.RequestCtx {
	rsMutex.Lock()
	defer rsMutex.Unlock()
	k := string(key)
	rs := rsMap[k]
	if rs == nil {
		return nil
	}
	return rs.waitingList
}

// GetWatingRequstAndSetStatus 获取等待中的请求，并设置状态和有效期
func GetWatingRequstAndSetStatus(key []byte, status string, ttl uint16) []*fasthttp.RequestCtx {
	rsMutex.Lock()
	defer rsMutex.Unlock()
	k := string(key)
	rs := rsMap[k]
	if rs == nil {
		return nil
	}
	rs.status = status
	rs.ttl = ttl
	waitingList := rs.waitingList
	rs.waitingList = nil
	return waitingList
}

// InitDB 初始化db
func InitDB(file string) (*bolt.DB, error) {
	if client != nil {
		return client, nil
	}
	db, err := bolt.Open(file, 0600, nil)
	if err != nil {
		return nil, err
	}
	client = db
	return db, nil
}

// InitBucket 初始化bucket
func InitBucket(bucket []byte) error {
	if client == nil {
		return vars.ErrDbNotInit
	}
	return client.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucket)
		return err
	})
}

// SaveResponseData 保存Response
func SaveResponseData(bucket, key, buf, header []byte, ttl uint16) error {
	// 前四个字节保存创建时间
	// 接着后面两个字节保存ttl
	// 接着后面两个字节保存header的长度
	// 接着是header
	// 最后才是body
	createdAt := util.GetNowSecondsBytes()

	s := [][]byte{
		createdAt,
		util.ConvertUint16ToBytes(ttl),
		util.ConvertUint16ToBytes(uint16(len(header))),
		header,
		buf,
	}
	data := bytes.Join(s, []byte(""))
	return Save(bucket, key, data)
}

// GetResponse 获取response
func GetResponse(bucket, key []byte) (*ResponseData, error) {
	data, err := Get(bucket, key)
	if err != nil {
		return nil, err
	}
	headerLength := util.ConvertBytesToUint16(data[6:8])
	return &ResponseData{
		CreatedAt: util.ConvertBytesToUint32(data[0:4]),
		TTL:       util.ConvertBytesToUint16(data[4:6]),
		Header:    data[8 : 8+headerLength],
		Body:      data[8+headerLength:],
	}, nil
}

// Save 保存数据
func Save(bucket, key, buf []byte) error {
	if client == nil {
		return vars.ErrDbNotInit
	}
	return client.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		return b.Put(key, buf)
	})
}

// Get 获取数据
func Get(bucket, key []byte) ([]byte, error) {
	if client == nil {
		return nil, vars.ErrDbNotInit
	}
	var buf []byte
	client.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		buf = b.Get(key)
		return nil
	})
	return buf, nil
}
