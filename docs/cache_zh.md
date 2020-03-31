---
description: 缓存
---

HTTP缓存使用内存缓存桶(lru)，为了提升性能可以配置使用多个缓存桶，在使用时根据应用选择适当的配置，一般设置256个缓存桶，每个缓存桶的大小设置为1024则可满足各类应用场景。

## 缓存的获取

- 根据请求的URL生成识别串(Method + Host + RequsetURI)
- 通过MemHash生成hash值，根据缓存桶的数据取余获取对应的缓存桶
- 从缓存桶中获取缓存数据

<p align="center">
<img src="./docs/cache-flow.jpg"/>
</p>

HTTP缓存的处理非常简单，使用lru保证了常用缓存的时效性，也避免了过多的缓存占用太多的内存空间。

## 缓存有效期

HTTP缓存的有效期从`Cache-Control`响应头中获取，获取有效期的流程如下：

<p align="center">
<img src="./docs/cache-age.jpg"/>
</p>

## 缓存建议

Pike的设计保证了当缓存不存在时，相同的请求只会有一个请求至upstream，整体设计主要是为了应对高并发时系统性能下降，并不建议使用它来提升一个本来响应慢的请求。在使用时，也建议使用短缓存(Cache-Control中设置max-age或s-maxage)，避免需要手工删除数据，建议缓存时长不超过5分钟即可。

