// Copyright 2019 tree xie
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// The client to read data of config,
// such as etcd and file.

package config

import (
	"context"
	"net/url"
	"strings"
	"time"

	"go.etcd.io/etcd/clientv3"
)

type (
	// Client client interface
	Client interface {
		// Get get the data of key
		Get(key string) (data []byte, err error)
		// Set set the data of key
		Set(key string, data []byte) (err error)
		// Delete delete the data of key
		Delete(key string) (err error)
		// List list all sub keys of key
		List(prefix string) (keys []string, err error)
	}
	// EtcdClient etcd client
	EtcdClient struct {
		c       *clientv3.Client
		Timeout time.Duration
	}
)

const (
	defaultEtcdTimeout = 5 * time.Second
)

// NewEtcdClient create a new etcd client
func NewEtcdClient(uri string) (client *EtcdClient, err error) {
	u, err := url.Parse(uri)
	if err != nil {
		return
	}
	conf := clientv3.Config{
		Endpoints: strings.Split(u.Host, ","),
	}
	if u.User != nil {
		conf.Username = u.User.Username()
	}
	c, err := clientv3.New(conf)
	if err != nil {
		return
	}
	client = &EtcdClient{
		c: c,
	}
	return
}

func (ec *EtcdClient) context() (context.Context, context.CancelFunc) {
	d := ec.Timeout
	if d == 0 {
		d = defaultEtcdTimeout
	}
	return context.WithTimeout(context.Background(), d)
}

// Get get data from etcd
func (ec *EtcdClient) Get(key string) (data []byte, err error) {
	ctx, cancel := ec.context()
	defer cancel()
	resp, err := ec.c.Get(ctx, key)
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
func (ec *EtcdClient) Set(key string, data []byte) (err error) {
	ctx, cancel := ec.context()
	defer cancel()
	_, err = ec.c.Put(ctx, key, string(data))
	return
}

// List list all key of the prefix from etcd
func (ec *EtcdClient) List(prefix string) (keys []string, err error) {
	ctx, cancel := ec.context()
	defer cancel()
	resp, err := ec.c.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return
	}
	keys = make([]string, len(resp.Kvs))
	for index, item := range resp.Kvs {
		keys[index] = string(item.Key)
	}
	return
}

// Delete delete the data of key
func (ec *EtcdClient) Delete(key string) (err error) {
	ctx, cancel := ec.context()
	defer cancel()
	_, err = ec.c.Delete(ctx, key)
	return
}
