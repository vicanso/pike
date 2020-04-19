---
description: 性能测试 
---

缓存服务最重要的是性能指标，varnish号称的是性能瓶颈在于服务器网卡，pike虽然达不到varnish性能指标，但测试结果也是十分可观。

测试机器：8U 8GB内存，wrk与测试程序均在同一机器上运行
测试数据：数据原始长度约为140KB
测试环境：客户端支持gzip、 br以及不支持压缩三种场景，并发请求数设置为1000，测试时长为1分钟

## gzip

客户端支持gzip压缩，pike返回已压缩的gzip数据(wrk运行于同一台机器中，大概占用2U的资源)。从下面的测试结果可以看出，每秒的处理请求数与传输的数据量，这已能满足大部分的应用场景。

```bash
wrk -c1000 -t10 -d1m -H 'Accept-Encoding: gzip, deflate' --latency 'http://127.0.0.1:8080/'
Running 1m test @ http://127.0.0.1:8080/
  10 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    13.89ms   10.33ms 192.87ms   79.01%
    Req/Sec     7.47k     1.53k   30.82k    71.94%
  Latency Distribution
     50%   12.83ms
     75%   16.84ms
     90%   23.29ms
     99%   51.84ms
  4453517 requests in 1.00m, 44.42GB read
Requests/sec:  74102.53
Transfer/sec:    756.79MB
```

## br

客户端支持br压缩，pike返回已压缩的br数据(wrk运行于同一台机器中，大概占用2U的资源)。从下面的测试结果可以看出，每秒的处理请求数为与gzip基本一致且数据量更少，这已能满足大部分的应用场景。

```bash
wrk -c1000 -t10 -d1m -H 'Accept-Encoding: br, gzip, deflate' --latency 'http://127.0.0.1:8080/'
Running 1m test @ http://127.0.0.1:8080/
  10 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    13.54ms    9.83ms 173.94ms   77.98%
    Req/Sec     7.60k     1.56k   15.71k    70.23%
  Latency Distribution
     50%   12.48ms
     75%   16.56ms
     90%   23.14ms
     99%   49.92ms
  4527025 requests in 1.00m, 37.32GB read
Requests/sec:  75352.10
Transfer/sec:    636.17MB
```


## 不支持压缩

客户端不支持压缩，pike解压gzip数据返回(wrk运行于同一台机器中，大概占用1U的资源)。从下面的测试结果可以看出，每秒处理的请求数下降比较明显。由于基本所有的客户端都能支持`gzip`压缩，大部分的也支持`br`，所以pike在缓存数据时，对于可压缩数据则会预压缩生成gzip与br数据，对于不支持压缩的从gzip数据中解压获取，因此解压会占用大量CPU资源导致响应更慢。

```bash
wrk -c1000 -t10 -d1m --latency 'http://127.0.0.1:8080/'
Running 1m test @ http://127.0.0.1:8080/
  10 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   313.85ms  397.12ms   2.00s    82.40%
    Req/Sec   581.74    163.59     1.55k    69.50%
  Latency Distribution
     50%   67.44ms
     75%  555.22ms
     90%  902.45ms
     99%    1.60s
  347280 requests in 1.00m, 52.49GB read
  Socket errors: connect 0, read 0, write 0, timeout 2159
Requests/sec:   5783.48
Transfer/sec:      0.87GB
```