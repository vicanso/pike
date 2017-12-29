package cache

import (
	"sync"
	"time"

	"github.com/boltdb/bolt"

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

// Init 初始化db
func Init(file string) (*bolt.DB, error) {
	if client != nil {
		return client, nil
	}
	db, err := bolt.Open(file, 0600, nil)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists(vars.ResponseBucket)
		return err
	})
	if err != nil {
		db.Close()
		return nil, err
	}
	client = db
	return db, err
}

// Save 保存数据
func Save(key string, buf []byte) error {
	if client == nil {
		return vars.ErrDbNotInit
	}
	return client.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(vars.ResponseBucket)
		return b.Put([]byte(key), buf)
	})
}

// Get 获取数据
func Get(key string) ([]byte, error) {
	if client == nil {
		return nil, vars.ErrDbNotInit
	}
	var buf []byte
	client.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(vars.ResponseBucket)
		buf = b.Get([]byte(key))
		return nil
	})
	return buf, nil
}
