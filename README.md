# pike

HTTP缓存服务，提供高效简单的HTTP缓存服务。

一直以来都使用`varnish`来做HTTP缓存，喜欢它的性能高效与vcl配置的强大。在规范化缓存的配置之后，强大的vcl对于我也没有太多的作用了，此时我更希望易上手，更简洁的配置，`Pike`则由此诞生。

`Pike`主要基于`fasthttp`与`badger`两个开源库开发


[![Build Status](https://img.shields.io/travis/vicanso/pike.svg?label=linux+build)](https://travis-ci.org/vicanso/pike)

## 特性

- 基于yaml的配置，简洁易懂
- WEB管理后台，提供系统性能、黑名单IP、缓存清理功能
- 标准化的基于HTTP头Cache-Control缓存控制
- 压缩保存的响应数据，避免每次响应时重新压缩（如果客户端不支持压缩则解压）
- 自定义日志格式，支持二十多种placeholder，如：cookie，请求头字段，响应头字段，响应时间等。
- 访问日志支持以文件（按天分隔）或者UDP的形式输出
- 支持自定义HTTP响应头配置
- 支持自定义最小压缩长度，对于内网之间的访问，避免压缩、解压的时间损耗

## 配置

使用yaml定义配置文件，更精简易懂，下面是示例配置：

```yaml
# 程序监听的端口，默认为 :3015
listen: :3015
# 数据缓存的db目录，默认为 /tmp/pike 
db: /tmp/pike
# 设置upstream的连接超时，默认为0，0表示无限制(time.Duration)，不建议使用默认值
connectTimeout: 3s 
directors:
  -
    # 名称
    name: tiny
    # backend的健康检测，如果不配置，则不做检测，默认backend都为正常
    ping: /ping
    # backend列表
    backends:
      - 127.0.0.1:5018
      - 192.168.31.3:3001
      - 192.168.31.3:3002
```

## 性能

Pike默认是会将缓存的数据压缩（大于配置的最小压缩长度），在响应时会根据客户端是否支持压缩而解压数据，因此为了避免解压影响性能，请求头带上支持gzip压缩头

### 对pike的压测

```
wrk -H 'Accept-Encoding: gzip, deflate' -t10 -c200 -d1m 'http://127.0.0.1:3015/ping' --latency

Running 1m test @ http://127.0.0.1:3015/ping
  10 threads and 200 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     1.12ms    1.53ms  46.46ms   92.87%
    Req/Sec    23.71k     2.29k   36.60k    74.70%
  Latency Distribution
     50%  770.00us
     75%    1.13ms
     90%    1.94ms
     99%    8.27ms
  14168320 requests in 1.00m, 2.23GB read
Requests/sec: 235773.39
```

### 可缓存请求的压测

```
wrk -H 'Accept-Encoding: gzip, deflate' -t10 -c200 -d1m 'http://127.0.0.1:3015/api/sys/status' --latency

Running 1m test @ http://127.0.0.1:3015/api/sys/status
  10 threads and 200 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     2.17ms    2.39ms  64.81ms   90.41%
    Req/Sec    11.27k     2.25k   22.45k    73.91%
  Latency Distribution
     50%    1.63ms
     75%    2.57ms
     90%    4.44ms
     99%   12.08ms
  6732203 requests in 1.00m, 32.80GB read
Requests/sec: 112036.59
Transfer/sec:    558.91MB
```

### 不可缓存请求的压测

```
# 通过pike转发
wrk -H 'Accept-Encoding: gzip, deflate' -t10 -c200 -d1m 'http://127.0.0.1:3015/api/users/me' --latency

Running 1m test @ http://127.0.0.1:3015/api/users/me
  10 threads and 200 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    13.86ms    2.30ms  92.20ms   88.69%
    Req/Sec     1.45k   151.81     1.62k    79.18%
  Latency Distribution
     50%   13.20ms
     75%   14.13ms
     90%   16.46ms
     99%   19.77ms
  867906 requests in 1.00m, 191.20MB read
Requests/sec:  14459.01
Transfer/sec:      3.19MB

# 直接压测
wrk -H 'Accept-Encoding: gzip, deflate' -t10 -c200 -d1m 'http://127.0.0.1:5018/api/users/me' --latency

Running 1m test @ http://127.0.0.1:5018/api/users/me
  10 threads and 200 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    12.42ms  834.04us  30.14ms   86.07%
    Req/Sec     1.62k    81.15     1.82k    78.58%
  Latency Distribution
     50%   12.17ms
     75%   12.72ms
     90%   13.34ms
     99%   15.38ms
  966524 requests in 1.00m, 140.11MB read
Requests/sec:  16101.42
Transfer/sec:      2.33MB
```

由上面的测试结果可以看出，得益于`fasthttp`与`badger`的高性能，对于可缓存的请求，`pike`的处理可以达到100k Requests/sec（数据量500多MB每秒）。而对于不可缓存的转发请求，平均大概增加了1ms的处理时间。


## 启动方式

### docker

```bash
docker run --restart=always -p 3015:3015 -v ~/pike/config.yml:/etc/pike/config.yml vicanso/pike
```

### bin

```bash
./pike -c ~/pike/config.yml
```
