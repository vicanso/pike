package cache

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCacheStatusString(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("fetching", StatusFetching.String())
	assert.Equal("hitForPass", StatusHitForPass.String())
	assert.Equal("hit", StatusHit.String())
	assert.Equal("passed", StatusPassed.String())
	assert.Equal("unknown", StatusUnknown.String())
}

func TestHTTPCacheBytes(t *testing.T) {
	assert := assert.New(t)
	hc := httpCache{
		status: StatusFetching,
		response: &HTTPResponse{
			CompressSrv: "compress",
		},
		createdAt: 1,
		expiredAt: 2,
	}
	data, err := hc.Bytes()
	assert.Nil(err)

	newHC := NewHTTPCache()
	err = newHC.FromBytes(data)
	assert.Nil(err)
	assert.Equal(hc.status, newHC.status)
	assert.Equal(hc.response.CompressSrv, newHC.response.CompressSrv)
	assert.Equal(hc.createdAt, newHC.createdAt)
	assert.Equal(hc.expiredAt, newHC.expiredAt)
}

func TestHTTPCacheGet(t *testing.T) {
	assert := assert.New(t)
	cacheResp, err := NewHTTPResponse(200, nil, "", []byte("Hello world!"))
	// 避免压缩，方便后面对数据检测
	cacheResp.CompressMinLength = 1024
	assert.Nil(err)
	expiredHC := NewHTTPCache()
	expiredHC.expiredAt = 1
	expiredHC.status = StatusHitForPass

	tests := []struct {
		status Status
		hc     *httpCache
		resp   *HTTPResponse
	}{
		{
			status: StatusHit,
			hc:     expiredHC,
			resp:   cacheResp,
		},
		{
			status: StatusHitForPass,
			hc:     NewHTTPCache(),
		},
	}
	type testResult struct {
		status Status
		resp   *HTTPResponse
	}
	for _, tt := range tests {
		mu := sync.Mutex{}
		wg := sync.WaitGroup{}
		results := make([]*testResult, 0)
		max := 10
		for i := 0; i < max; i++ {
			wg.Add(1)
			go func() {
				status, resp := tt.hc.Get()
				mu.Lock()
				defer mu.Unlock()
				results = append(results, &testResult{
					status: status,
					resp:   resp,
				})
				wg.Done()
			}()
		}
		// 简单等待10ms，让所有for中的goroutine都已执行
		time.Sleep(10 * time.Millisecond)
		switch tt.status {
		case StatusHit:
			tt.hc.Cacheable(tt.resp, 300)
		case StatusHitForPass:
			tt.hc.HitForPass(-1)
		}

		wg.Wait()
		count := 0
		for _, result := range results {
			// 如果不相等的，只能是fetching
			if result.status != tt.status {
				assert.Equal(StatusFetching, result.status)
			} else {
				assert.Equal(tt.status, result.status)
				count++
				// fetching的数据由fetching取，不使用缓存返回
				assert.Equal(tt.resp, result.resp)
			}
		}
		// 其它状态的都相同
		assert.Equal(max-1, count)

		// 在后续已设置缓存状态之后，再次获取直接返回
		status, resp := tt.hc.Get()
		assert.Equal(tt.status, status)
		assert.Equal(tt.resp, resp)
	}
}

func TestHTTPCacheAge(t *testing.T) {
	assert := assert.New(t)
	hc := httpCache{
		createdAt: nowUnix() - 1,
		mu:        &sync.RWMutex{},
	}
	assert.GreaterOrEqual(hc.Age(), 1)
}

func TestHTTPCacheGetStatus(t *testing.T) {
	assert := assert.New(t)
	hc := httpCache{
		status: StatusFetching,
		mu:     &sync.RWMutex{},
	}
	assert.Equal(StatusFetching, hc.GetStatus())
}

func TestHTTPCacheIsExpired(t *testing.T) {
	assert := assert.New(t)
	hc := httpCache{
		expiredAt: 0,
		mu:        &sync.RWMutex{},
	}
	assert.False(hc.IsExpired())
	hc.expiredAt = 1
	assert.True(hc.IsExpired())
	hc.expiredAt = nowUnix() + 10
	assert.False(hc.IsExpired())
}
