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
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const defaultMongoCacheColletion = "caches"
const defaultMongoDatabase = "pike"

type mongoStore struct {
	client  *mongo.Client
	db      string
	timeout time.Duration
}

type mongoCache struct {
	Key       string    `json:"key,omitempty" bson:"key,omitempty" `
	Data      []byte    `json:"data,omitempty" bson:"data,omitempty"`
	ExpiredAt time.Time `json:"expiredAt,omitempty" bson:"expiredAt,omitempty"`
}

func fillMongoStoreOptions(connectionURI string, ms *mongoStore) {
	ms.db = defaultMongoDatabase
	ms.timeout = 3 * time.Second
	urlInfo, _ := url.Parse(connectionURI)
	if urlInfo == nil {
		return
	}
	arr := strings.Split(urlInfo.Path, "/")
	if len(arr) >= 2 {
		ms.db = arr[1]
	}
	// 设置的超时，如 3s
	timeout, _ := time.ParseDuration(urlInfo.Query().Get("timeout"))
	if timeout != 0 {
		ms.timeout = timeout
	}
}

func newMongoStore(connectionURI string) (store Store, err error) {
	clientOptions := options.Client().ApplyURI(connectionURI)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return
	}
	ms := &mongoStore{
		client: client,
	}
	fillMongoStoreOptions(connectionURI, ms)

	err = client.Ping(ctx, nil)
	if err != nil {
		return
	}

	// 创建索引
	expireAfterSeconds := int32(0)
	background := true
	unique := true
	_, err = ms.collection().Indexes().CreateMany(context.TODO(), []mongo.IndexModel{
		// key索引
		{
			Keys: bson.M{
				"key": 1,
			},
			Options: &options.IndexOptions{
				Unique:     &unique,
				Background: &background,
			},
		},
		// 数据自助过期索引
		{
			Keys: bson.M{
				"expiredAt": 1,
			},
			Options: &options.IndexOptions{
				Background:         &background,
				ExpireAfterSeconds: &expireAfterSeconds,
			},
		},
	})
	if err != nil {
		return
	}

	store = ms
	return
}

func (ms *mongoStore) collection() *mongo.Collection {
	return ms.client.Database(ms.db).Collection(defaultMongoCacheColletion)
}

// Get gets data from mongo
func (ms *mongoStore) Get(key []byte) (data []byte, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), ms.timeout)
	defer cancel()
	result := mongoCache{}
	err = ms.collection().FindOne(ctx, &mongoCache{
		Key: string(key),
	}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			err = ErrNotFound
		}
		return
	}
	data = result.Data
	return
}

// Set sets data to mongo
func (ms *mongoStore) Set(key []byte, data []byte, ttl time.Duration) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), ms.timeout)
	defer cancel()
	upsert := true
	fmt.Println(time.Now().Add(ttl))
	_, err = ms.collection().UpdateOne(ctx, &mongoCache{
		Key: string(key),
	}, bson.M{
		"$set": &mongoCache{
			Key:       string(key),
			Data:      data,
			ExpiredAt: time.Now().Add(ttl),
		},
	}, &options.UpdateOptions{
		Upsert: &upsert,
	})
	if err != nil {
		return
	}
	return
}

// Delete deletes data from mongo
func (ms *mongoStore) Delete(key []byte) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), ms.timeout)
	defer cancel()
	_, err = ms.collection().DeleteOne(ctx, &mongoCache{
		Key: string(key),
	})
	if err != nil {
		return
	}
	return
}

// Close closes mongo
func (ms *mongoStore) Close() error {
	return ms.client.Disconnect(context.TODO())
}
