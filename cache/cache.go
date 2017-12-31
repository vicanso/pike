package cache

import (
	"bytes"
	"sync"
	"time"

	"github.com/boltdb/bolt"

	"../util"
	"../vars"
)

// 保存请求状态
var requestStatsMap sync.Map
var client *bolt.DB

// Status 请求状态
type Status struct {
	Name      string
	CreatedAt int64
}

// ResponseData 记录响应数据
type ResponseData struct {
	CreatedAt uint32
	TTL       uint16
	Header    []byte
	Body      []byte
}

// GetStatus 获取该key对应的请求状态
func GetStatus(key string) string {
	v, loaded := requestStatsMap.LoadOrStore(key, &Status{
		Name:      vars.Fetching,
		CreatedAt: time.Now().Unix(),
	})
	if !loaded {
		return vars.None
	}
	return v.(*Status).Name
}

// DeleteStatus 删除该key对应的状态
func DeleteStatus(key string) {
	requestStatsMap.Delete(key)
}

// SetHitForPass 判断该key是否hit for pass
func SetHitForPass(key string) {
	requestStatsMap.Store(key, &Status{
		Name:      vars.HitForPass,
		CreatedAt: time.Now().Unix(),
	})
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
