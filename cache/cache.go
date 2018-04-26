package cache

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/vicanso/pike/util"
	"github.com/vicanso/pike/vars"

	"github.com/akrylysov/pogreb"
)

const (
	// Pass request status: pass
	Pass = iota + 1
	// Fetching request status: fetching
	Fetching
	// Waiting request status: wating
	Waiting
	// HitForPass request status: hitForPass
	HitForPass
	// Cacheable request status: cacheable
	Cacheable
)

type (

	// Client 缓存
	Client struct {
		Path  string
		db    *pogreb.DB
		rsMap map[string]*RequestStatus
		mutex *sync.Mutex
	}
	// Response 响应数据
	Response struct {
		// 创建时间
		CreatedAt uint32 `json:"createdAt"`
		// HTTP状态码
		StatusCode uint16 `json:"statusCode"`
		// 缓存有效时间(最大65535)
		TTL uint16 `json:"ttl"`
		// HTTP响应头
		Header http.Header `json:"header"`
		// HTTP响应数据
		Body []byte `json:"body"`
		// HTTP响应数据(gzip)
		GzipBody []byte `json:"gzip"`
		// HTTP响应数据(br)
		BrBody []byte `json:"br"`
		// 压缩数据级别
		CompressLevel int `json:"compressLevel"`
		// 最小压缩数据
		CompressMinLength int `json:"compressMinLength"`
	}
	// RequestStatus 获取请求状态
	RequestStatus struct {
		createdAt uint32
		ttl       uint16
		// 请求状态 fetching hitForPass 等
		status int
		// 如果此请求为fetching，则此时相同的请求会写入一个chan
		waitingChans []chan int
	}
	// Stats 各状态数量统计
	Stats struct {
		Waiting    int `json:"waiting"`
		Fetching   int `json:"fetching"`
		HitForPass int `json:"hitForPass"`
		Cacheable  int `json:"cacheable"`
	}
)

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

// 判断是否已过期
func isExpired(rs *RequestStatus) bool {
	now := uint32(time.Now().Unix())
	if rs.ttl != 0 && now-rs.createdAt > uint32(rs.ttl) {
		return true
	}
	return false
}

func (r *Response) getRawBody() ([]byte, error) {
	if len(r.Body) != 0 {
		return r.Body, nil
	}
	if len(r.GzipBody) != 0 {
		return util.Gunzip(r.GzipBody)
	}
	return nil, vars.ErrBodyCotentNotFound
}

// GetBody 根据accept encondings 获取数据
func (r *Response) GetBody(acceptEncoding string) (body []byte, encoding string) {
	// 如果是204,直接返回
	if r.StatusCode == http.StatusNoContent {
		return
	}
	compressMinLength := r.CompressMinLength
	if compressMinLength == 0 {
		compressMinLength = vars.CompressMinLength
	}
	rawBodySize := len(r.Body)
	// 如果原始数据小于最低压缩限制，则直接返回
	if rawBodySize != 0 && rawBodySize < compressMinLength {
		body = r.Body
		return
	}
	level := r.CompressLevel
	supportEncondings := []string{
		vars.BrEncoding,
		vars.GzipEncoding,
	}
	for _, enc := range supportEncondings {
		if !strings.Contains(acceptEncoding, enc) {
			continue
		}
		if enc == vars.BrEncoding {
			if len(r.BrBody) != 0 {
				body = r.BrBody
				encoding = enc
				return
			}
			// 获取原始未压缩数据
			raw, err := r.getRawBody()
			if err != nil {
				continue
			}
			// 做br压缩
			brBody, err := util.Brotli(raw, level)
			// 如果压缩出错，使用下一个encoding
			if err != nil {
				continue
			}
			body = brBody
			encoding = enc
			return
		} else if enc == vars.GzipEncoding {
			if len(r.GzipBody) != 0 {
				body = r.GzipBody
				encoding = enc
				return
			}
			// gzip压缩
			gzipBody, err := util.Gzip(r.Body, level)
			if err != nil {
				continue
			}
			body = gzipBody
			encoding = enc
			return
		}
	}

	body = r.Body
	return
}

// Init 初始化缓存
func (c *Client) Init() error {
	os.Remove(c.Path + ".lock")
	db, err := pogreb.Open(c.Path, nil)
	c.db = db
	c.rsMap = make(map[string]*RequestStatus)
	c.mutex = &sync.Mutex{}
	return err
}

// Close 关闭缓存
func (c *Client) Close() error {
	return c.db.Close()
}

// SaveResponse 保存response
func (c *Client) SaveResponse(key []byte, resp *Response) error {

	createdAt := resp.CreatedAt
	if createdAt == 0 {
		createdAt = uint32(time.Now().Unix())
	}
	header, err := json.Marshal(resp.Header)
	if err != nil {
		return err
	}
	// 将要保存的数据转换为bytes
	body := resp.Body
	gzipBody := resp.GzipBody
	brBody := resp.BrBody
	s := [][]byte{
		uint32ToBytes(createdAt),
		uint16ToBytes(resp.StatusCode),
		uint16ToBytes(resp.TTL),
		uint32ToBytes(uint32(len(header))),
		uint32ToBytes(uint32(len(body))),
		uint32ToBytes(uint32(len(gzipBody))),
		uint32ToBytes(uint32(len(brBody))),
		header,
		body,
		gzipBody,
		brBody,
	}
	data := bytes.Join(s, nil)
	return c.db.Put(key, data)
}

// GetResponse 从缓存中获取Response
func (c *Client) GetResponse(key []byte) (resp *Response, err error) {
	data, err := c.db.Get(key)
	if err != nil {
		return
	}
	if data == nil || len(data) == 0 {
		return
	}
	resp = &Response{}
	resp.CreatedAt = bytesToUint32(data[0:4])
	resp.StatusCode = bytesToUint16(data[4:6])
	resp.TTL = bytesToUint16(data[6:8])
	headerLength := bytesToUint32(data[8:12])
	bodyLength := bytesToUint32(data[12:16])
	gzipLength := bytesToUint32(data[16:20])
	brLength := bytesToUint32(data[20:24])
	var offset uint32 = 24
	header := make(http.Header)
	err = json.Unmarshal(data[offset:offset+headerLength], &header)
	offset += headerLength
	if err != nil {
		return
	}
	resp.Header = header

	resp.Body = data[offset : offset+bodyLength]
	offset += bodyLength

	resp.GzipBody = data[offset : offset+gzipLength]
	offset += gzipLength

	resp.BrBody = data[offset : offset+brLength]

	return
}

// GetRequestStatus 获取key对应的请求status
func (c *Client) GetRequestStatus(key []byte) (status int, ch chan int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	k := string(key)
	rs := c.rsMap[k]
	// 如果该key对应的状态为空或者已过期
	if rs == nil || isExpired(rs) {
		status = Fetching
		rs = &RequestStatus{
			createdAt:    uint32(time.Now().Unix()),
			ttl:          0,
			waitingChans: make([]chan int, 0),
			status:       Fetching,
		}
		c.rsMap[k] = rs
	} else if rs.status == Fetching {
		// 如果该key对应的请求正在处理中，添加chan
		status = Waiting
		ch = make(chan int)
		rs.waitingChans = append(rs.waitingChans, ch)
	} else {
		// hit for pass 或者 cacheable
		status = rs.status
	}
	return
}

// UpdateRequestStatus 更新状态，获取等待中的请求，并设置状态和有效期
func (c *Client) UpdateRequestStatus(key []byte, status int, ttl uint16) {

	c.mutex.Lock()
	defer c.mutex.Unlock()
	k := string(key)
	rs := c.rsMap[k]
	if rs == nil {
		return
	}
	rs.status = status
	rs.ttl = ttl
	waitingChans := rs.waitingChans
	// 对所有等待中的请求触发channel
	for _, c := range waitingChans {
		c <- status
	}
	rs.waitingChans = nil
}

// HitForPass 设置为hit for pass
func (c *Client) HitForPass(key []byte, ttl uint16) {
	c.UpdateRequestStatus(key, HitForPass, ttl)
}

// Cacheable 设置状态为cacheable
func (c *Client) Cacheable(key []byte, ttl uint16) {
	c.UpdateRequestStatus(key, Cacheable, ttl)
}

// ClearExpired 清除过期数据
func (c *Client) ClearExpired(delay int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	now := uint32(time.Now().Unix())
	// 为了避免删除数据之后，如果并发在请求rsmap为cacheable之后有可能导致获取数据失败，需要设置delay
	if delay < 0 {
		delay = 60
	}
	for k, v := range c.rsMap {
		ttl := v.ttl
		if ttl != 0 && now-v.createdAt > uint32(ttl)+uint32(delay) {
			delete(c.rsMap, k)
			c.db.Delete([]byte(k))
		}
	}
}

// Size 获取缓存数量
func (c *Client) Size() int {
	return len(c.rsMap)
}

// GetStats 获取缓存状态统计
func (c *Client) GetStats() (stats *Stats) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	stats = &Stats{}
	for _, v := range c.rsMap {
		switch v.status {
		case Fetching:
			stats.Fetching++
			stats.Waiting += len(v.waitingChans)
			break
		case HitForPass:
			stats.HitForPass++
			break
		case Cacheable:
			stats.Cacheable++
			break
		}
	}
	return
}
