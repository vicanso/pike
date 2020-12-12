// MIT License

// Copyright (c) 2020 Tree Xie

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

// etcd client for config

package config

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
)

// etcdClient etcd client
type etcdClient struct {
	c       *clientv3.Client
	key     string
	Timeout time.Duration
}

const (
	defaultEtcdTimeout = 5 * time.Second
)

// NewEtcdClient create a new etcd client
func NewEtcdClient(uri string) (client *etcdClient, err error) {
	u, err := url.Parse(uri)
	if err != nil {
		return
	}
	conf := clientv3.Config{
		Endpoints:   strings.Split(u.Host, ","),
		DialTimeout: defaultEtcdTimeout,
	}
	if u.User != nil {
		conf.Username = u.User.Username()
		conf.Password, _ = u.User.Password()
	}
	// TODO 后续有需要添加支持tls
	// TODO 后续支持从querystring中配置更多的参数
	c, err := clientv3.New(conf)
	if err != nil {
		return
	}
	client = &etcdClient{
		c:   c,
		key: u.Path,
	}
	return
}

func (ec *etcdClient) context() (context.Context, context.CancelFunc) {
	d := ec.Timeout
	if d == 0 {
		d = defaultEtcdTimeout
	}
	return context.WithTimeout(context.Background(), d)
}

// Get get data from etcd
func (ec *etcdClient) Get() (data []byte, err error) {
	ctx, cancel := ec.context()
	defer cancel()
	resp, err := ec.c.Get(ctx, ec.key)
	if err != nil {
		return
	}
	kvs := resp.Kvs
	if len(kvs) == 0 {
		return
	}
	data = kvs[0].Value
	return
}

// Set set data to etcd
func (ec *etcdClient) Set(data []byte) (err error) {
	ctx, cancel := ec.context()
	defer cancel()
	_, err = ec.c.Put(ctx, ec.key, string(data))
	return
}

// Watch watch config change
func (ec *etcdClient) Watch(onChange OnChange) {
	ch := ec.c.Watch(context.Background(), ec.key)
	// 只监听有变化则可
	for range ch {
		onChange()
	}
}

// Close close etcd client
func (ec *etcdClient) Close() error {
	return ec.c.Close()
}
