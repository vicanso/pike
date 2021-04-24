// MIT License

// Copyright (c) 2021 Tree Xie

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package store

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/vicanso/pike/log"
	"go.uber.org/zap"
)

type redisStore struct {
	client redis.UniversalClient
	// timeout 超时设置
	timeout time.Duration
	// prefix key的前缀
	prefix string
}

type redisLogger struct{}

func (rl *redisLogger) Printf(ctx context.Context, format string, v ...interface{}) {
	log.Default().Info(fmt.Sprintf(format, v...),
		zap.String("category", "redisLogger"),
	)
}

func init() {
	redis.SetLogger(&redisLogger{})
}

func newRedisStore(connectionURI string) (store Store, err error) {
	urlInfo, err := url.Parse(connectionURI)
	if err != nil {
		return
	}
	user := ""
	password := ""
	if urlInfo.User != nil {
		user = urlInfo.User.Username()
		password, _ = urlInfo.User.Password()
	}
	// redis选择的db
	db, _ := strconv.Atoi(urlInfo.Query().Get("db"))
	// 设置的超时，如 3s
	timeout, _ := time.ParseDuration(urlInfo.Query().Get("timeout"))
	// 保存的key的前缀
	prefix := urlInfo.Query().Get("prefix")

	addrs := strings.Split(urlInfo.Host, ",")
	master := urlInfo.Query().Get("master")

	client := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:            addrs,
		Username:         user,
		Password:         password,
		DB:               db,
		SentinelPassword: password,
		MasterName:       master,
	})

	// 默认3秒超时
	if timeout == 0 {
		timeout = 3 * time.Second
	}
	store = &redisStore{
		client:  client,
		timeout: timeout,
		prefix:  prefix,
	}
	return
}

func (rs *redisStore) getKey(key []byte) string {
	return rs.prefix + string(key)
}

// Get get data from redis
func (rs *redisStore) Get(key []byte) (data []byte, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), rs.timeout)
	defer cancel()
	k := rs.getKey(key)
	cmd := rs.client.Get(ctx, k)
	data, err = cmd.Bytes()
	if err != nil {
		if err == redis.Nil {
			err = ErrNotFound
		}
		return
	}
	return
}

// Set set data to redis
func (rs *redisStore) Set(key []byte, data []byte, ttl time.Duration) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), rs.timeout)
	defer cancel()
	k := rs.getKey(key)
	cmd := rs.client.Set(ctx, k, data, ttl)
	return cmd.Err()
}

// Delete delete date from redis
func (rs *redisStore) Delete(key []byte) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), rs.timeout)
	defer cancel()
	k := rs.getKey(key)
	cmd := rs.client.Del(ctx, k)
	return cmd.Err()
}

// Close close redis
func (rs *redisStore) Close() error {
	return rs.client.Close()
}
