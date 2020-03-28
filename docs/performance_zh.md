---
description: performance
---

缓存服务最重要的是性能指标，varnish号称的是性能瓶颈在于服务器网卡，pike虽然达不到varnish性能指标，但测试结果也是十分可观。

测试机器：8U 8GB内存，wrk与测试程序均在同一机器上运行
测试数据：数据原始长度约为140KB
测试环境：客户端支持gzip、 br以及不支持压缩三种场景，并发请求数设置为1000，测试时长为1分钟

## gzip

客户端支持gzip压缩，pike返回已压缩的gzip数据(wrk运行于同一台机器中，大概占用2U的资源)。从下面的测试结果可以看出，每秒的处理请求数为73k，1GB的数据传输，这已能满足大部分的应用场景。

```bash
wrk -c1000 -t10 -d1m -H 'Accept-Encoding: gzip, deflate' --latency 'http://127.0.0.1:8080/'
Running 1m test @ http://127.0.0.1:8080/
  10 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    13.73ms   10.19ms 174.52ms   83.55%
    Req/Sec     7.37k     1.43k   14.14k    70.48%
  Latency Distribution
     50%   11.77ms
     75%   15.84ms
     90%   23.45ms
     99%   54.41ms
  4389008 requests in 1.00m, 61.85GB read
Requests/sec:  73028.57
Transfer/sec:      1.03GB
```

## br

客户端支持br压缩，pike返回已压缩的br数据(wrk运行于同一台机器中，大概占用2U的资源)。从下面的测试结果可以看出，每秒的处理请求数为74k(与gzip基本相同)，0.91GB的数据传输(br的压缩率比gzip高），这已能满足大部分的应用场景。

```bash
wrk -c1000 -t10 -d1m -H 'Accept-Encoding: br, gzip, deflate' --latency 'http://127.0.0.1:8080/'
Running 1m test @ http://127.0.0.1:8080/
  10 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    13.57ms    9.85ms 163.54ms   83.15%
    Req/Sec     7.47k     1.44k   17.56k    71.21%
  Latency Distribution
     50%   11.69ms
     75%   15.60ms
     90%   23.15ms
     99%   53.68ms
  4447442 requests in 1.00m, 54.41GB read
Requests/sec:  74003.72
Transfer/sec:      0.91GB
```


## 不支持压缩

客户端不支持压缩，pike解压gzip数据返回(wrk运行于同一台机器中，大概占用1U的资源)。从下面的测试结果可以看出，每秒仅能处理请求数为5k，790MB的数据传输。由于基本所有的客户端都能支持`gzip`压缩，大部分的也支持`br`，所以pike在缓存数据时，对于可压缩数据则会预压缩生成gzip与br数据，对于不支持压缩的从gzip数据中解压获取，因此解压会占用大量CPU资源导致响应更慢。

```bash
wrk -c1000 -t10 -d1m --latency 'http://127.0.0.1:8080/'
Running 1m test @ http://127.0.0.1:8080/
  10 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   319.58ms  395.73ms   2.00s    82.30%
    Req/Sec   536.33    156.85     1.40k    70.79%
  Latency Distribution
     50%   76.09ms
     75%  569.69ms
     90%  905.92ms
     99%    1.57s
  319657 requests in 1.00m, 46.65GB read
  Socket errors: connect 0, read 0, write 0, timeout 2105
Requests/sec:   5321.82
Transfer/sec:    795.24MB
```