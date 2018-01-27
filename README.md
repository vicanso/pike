# pike

HTTP缓存服务，提供高效简单的HTTP缓存服务。

一直以来都使用`varnish`来做HTTP缓存，喜欢它的性能高效与vcl配置的强大。在规范化缓存的配置之后，强大的vcl对于我也没有太多的作用了，此时我更希望易上手，更简洁的配置，`Pike`则由此诞生。

`Pike`主要基于`fasthttp`与`badger`两个开源库开发


[![Build Status](https://img.shields.io/travis/vicanso/pike.svg?label=linux+build)](https://travis-ci.org/vicanso/pike)


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

### 对pike的压测

```
wrk -H 'Accept-Encoding:gzip, deflate' -t10 -c200 -d1m 'http://127.0.0.1:3015/ping' --latency

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
# 因为缓存的数据是做压缩的，客户端支持压缩则不需要重新解压
wrk -H 'Accept-Encoding:gzip, deflate' -t10 -c200 -d1m 'http://127.0.0.1:3015/api/sys/status' --latency

Running 1m test @ http://127.0.0.1:3015/api/sys/status
  10 threads and 200 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    16.82ms   21.67ms 272.58ms   87.60%
    Req/Sec     1.89k   535.73     4.61k    69.45%
  Latency Distribution
     50%   10.86ms
     75%   22.94ms
     90%   43.45ms
     99%  100.49ms
  1128495 requests in 1.00m, 24.84GB read
Requests/sec:  18784.70
```

### 不可缓存请求的压测

```
# 通过pike转发
wrk -H 'Accept-Encoding:gzip, deflate' -t10 -c200 -d1m 'http://127.0.0.1:3015/api/users/me' --latency

Running 1m test @ http://127.0.0.1:3015/api/users/me
  10 threads and 200 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   187.35ms   16.34ms 337.80ms   70.63%
    Req/Sec   107.18     39.39   202.00     75.45%
  Latency Distribution
     50%  180.62ms
     75%  200.57ms
     90%  210.80ms
     99%  228.15ms
  64001 requests in 1.00m, 1.41GB read
Requests/sec:   1065.84

# 直接压测
wrk -H 'Accept-Encoding:gzip, deflate' -t10 -c200 -d1m 'http://127.0.0.1:5018/api/users/me' --latency

Running 1m test @ http://127.0.0.1:5018/api/users/me
  10 threads and 200 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   154.60ms    7.37ms 239.32ms   80.22%
    Req/Sec   131.58     61.31   202.00     52.45%
  Latency Distribution
     50%  153.79ms
     75%  157.45ms
     90%  162.34ms
     99%  176.92ms
  77588 requests in 1.00m, 1.70GB read
Requests/sec:   1291.99
```


## 启动方式

### docker

```bash
docker run --restart=always -p 3015:3015 -v ~/pike/config.yml:/etc/pike/config.yml vicanso/pike
```

### bin

```bash
./pike -c ~/pike/config.yml
```
