---
description: 缓存处理
---

HTTP缓存使用了LRU缓存，能提供高效的缓存读取能力及有效的控制缓存过大。为了减少缓存中锁的影响，使用了128个LRU组成缓存桶，每次根据hash选择对应的LRU，提升性能。需要注意，缓存只针对`GET`与`HEAD`请求，其它的请求都是不可缓存请求。

## 缓存的获取

- 根据请求的URL生成key(Method + Host + RequestURI)
- 使用该key通过MemHash生成hash值取余获取对应的缓存桶
- 从缓存桶中获取缓存数据

## 缓存有效期

HTTP缓存的有效期仅支持从`Cache-Control`响应头中获取，获取有效期的流程如下：

- 如果响应头有`Set-Cookie`，则返回缓存有效期为0
- 如果响应头无`Cache-Control`，则返回缓存有效期为0
- 如果响应头中`Cache-Control`包含`no-cache`，`no-store`或者`private`，则返回有效期为0
- 如果响应头中`Cache-Control`包含`s-maxage`，则根据`s-maxage`获取缓存有效期
- 如果响应头中`Cache-Control`包含`max-age`，则根据`max-age`获取缓存有效期
- 如果响应头中有`Age`字段，则最终的缓存有效期需减去`Age`

## 缓存状态

- `passed` 如果请求非HEAD与GET请求，其缓存状态则为passed，直接跳过缓存转发至后端服务
- `fetching` 当请求对应的key无法查找到缓存时，其缓存状态则为fetching，表示无缓存转发至后端服务。当获取该请求响应时，如果可缓存，则将相关数据缓存。如果不可缓存时，则缓存hit for pass（只缓存状态不需要缓存数据）
- `cacheable` 当请求对应的key可以获取到缓存数据，且该数据是可缓存，则直接返回
- `hitForPass` 当请求对应的key获取到缓存数据，且该数据是hit for pass时，则直接转发至后端服务

## 缓存建议

Pike的设计保证了当缓存不存在时，相同的请求只会有一个请求至upstream，整体设计主要是为了应对高并发时系统性能下降，并不建议使用它来提升一个本来响应慢的请求。在使用时，也建议使用短缓存(Cache-Control中设置max-age或s-maxage)，避免需要手工删除数据，非静态文件建议缓存时长不超过5分钟，静态文件（url中有相应的版本号）建议不要超过1小时。