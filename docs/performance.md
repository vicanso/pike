---
description: performance
---

The most impotant indicator of cache is performance, varnish is is usually bound by the speed of the network, pike is a little slower than varnish, but the test result is satisfatory.

server: 8U 8GB meory, wrk and pike run in the same server
data: around 140KB
client: support gzip, br or not support compress, 1000 connections, 1 minute

## gzip

Client support gzip, pike will response gzip's data(wrk uses 2U of cpus during the test).The test result below, 73028.57 Requests/sec and 1.03GB Transfer/sec, shows it can meet demands of most scenarios.

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

Client support brotli, pike will response br's data(wrk uses 2U of cpus during the test).The test result below, 74003.72 Requests/sec(same as gzip) and 0.91GB Transfer/sec(compression ratio of br is higher), shows it can meet demands of most scenarios.


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


## Not Support Compression

Client does not support compression, pike will decompress the gzip's data(wrk uses 1U of cpus during the test)。The test result below, 5321.82 Requests/sec and 795.24MBTransfer/sec is bad. Because most clients support gzip(br), so pike compresses gzip(br) data to cache. If the client does not support compress, pike will decompress for response, which uses most of cpu.

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