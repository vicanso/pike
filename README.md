# pike

HTTP cache server like `varnish`.


## 模块

- cache 缓存模块
  - brotli 提供brotli压缩
  - gzip 提供gzip压缩
  - cache 根据key获取相应的缓存数据，以分区的形式保存，减少锁的场景提升性能
  - lru LRU缓存
- config 应用程序配置模块
  - config 程序配置
  - etcd_config 从ETCD中读写配置
  - file_config 从文件中读写配置
- middleware 相关中间件
  - initialization 初始请求，主要判断是否超出最大请求数，设置请求头以及成功处理后设置响应头
  - cache_identifier 判断请求是否可缓存，如果不可缓存，直接转至下一中间件。如果获取不至缓存，则设置为fetch状态，根据最终响应数据转换为hit for pass或cacheable。如果获取到缓存，则设置缓存并转至下一中间件。
  - proxy 如果该请求已获取到相应的缓存数据，则转至下一中间件，否则转发至相应的upstream
  - responder 将缓存转换为相应的HTTP响应数据

## 测试场景

- [x] 非GET、HEAD直接请求pass至后端
- [x] 非文本类数据不压缩
- [x] POST不可缓存请求，后端返回数据未压缩
- [x] POST不可缓存请求，后端返回数据已压缩
- [x] GET不可缓存请求，后端返回数据未压缩
- [x] GET不可缓存请求，后端返回数据已压缩
- [x] GET可缓存请求，后端返回数据未压缩
- [x] GET可缓存请求，后端返回数据已压缩
- [x] 后端返回数据未添加ETag
- [x] 后端返回数据已添加ETag
- [x] 304的处理


## ETCD

```bash
docker run -d --restart=always \
  -p 2379:2379 \
  --name etcd \
  vicanso/etcd etcd \
  --listen-client-urls 'http://0.0.0.0:2379' \
  --advertise-client-urls 'http://0.0.0.0:2379'
```