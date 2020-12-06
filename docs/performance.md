---
description: 性能测试
---

测试机器：8U 8GB内存
测试数据：数据原始长度约为140KB
测试场景：客户端支持gzip、 br以及不支持压缩三种场景，并发请求数设置为1000，测试时长为1分钟
测试结论：当客户端可接受gzip或br压缩时，测试的结果均非常接近，而客户端不接受压缩时，需要先解压数据，性能则大幅度下降

## GZIP

```bash
wrk -c1000 -t10 -d1m -H 'Accept-Encoding: gzip' --latency http://127.0.0.1:3015/repos
Running 1m test @ http://127.0.0.1:3015/repos
  10 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    17.12ms   17.05ms 241.93ms   83.49%
    Req/Sec     6.96k     1.17k   12.88k    73.49%
  Latency Distribution
     50%   16.78ms
     75%   29.72ms
     90%   38.34ms
     99%   66.55ms
  4153511 requests in 1.00m, 40.89GB read
Requests/sec:  69123.45
Transfer/sec:    696.85MB
```

## BR

```bash
wrk -c1000 -t10 -d1m -H 'Accept-Encoding: br' --latency http://127.0.0.1:3015/repos
Running 1m test @ http://127.0.0.1:3015/repos
  10 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    16.65ms   16.27ms 181.87ms   64.06%
    Req/Sec     7.08k     1.06k   17.58k    70.18%
  Latency Distribution
     50%   16.56ms
     75%   29.11ms
     90%   37.47ms
     99%   62.46ms
  4223664 requests in 1.00m, 36.57GB read
Requests/sec:  70302.18
Transfer/sec:    623.26MB
```

## 不支持压缩

```bash
wrk -c1000 -t10 -d1m  --latency http://127.0.0.1:3015/repos
Running 1m test @ http://127.0.0.1:3015/repos
  10 threads and 1000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   319.12ms  421.34ms   2.00s    81.56%
    Req/Sec   555.57    193.17     2.54k    69.97%
  Latency Distribution
     50%   32.36ms
     75%  589.84ms
     90%  965.91ms
     99%    1.61s
  331687 requests in 1.00m, 44.85GB read
  Socket errors: connect 0, read 0, write 0, timeout 4348
Requests/sec:   5518.90
Transfer/sec:    764.11MB
```
