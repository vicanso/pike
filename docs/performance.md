---
description: performance
---

The most impotant indicator of cache is performance, varnish is is usually bound by the speed of the network, pike is a little slower than varnish, but the test result is satisfatory.

server: 8U 8GB meory, wrk and pike run in the same server
data: around 140KB
client: support gzip, br or not support compress, 1000 connections, 1 minute

## gzip

Client support gzip, pike will response gzip's data(wrk uses 2U of cpus during the test).The test result below can meet demands of most scenarios.

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

Client support brotli, pike will response br's data(wrk uses 2U of cpus during the test).The test result below, Requests/sec the same as gzip and Tansfer/sec is less(compression ratio of br is higher), shows it can meet demands of most scenarios.


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


## Not Support Compression

Client does not support compression, pike will decompress the gzip's data(wrk uses 1U of cpus during the test)。The test result below Requests/sec and Transfer/sec is a little bad. Because most clients support gzip(br), so pike compresses gzip(br) data to cache. If the client does not support compress, pike will decompress for response, which uses most of cpu.

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