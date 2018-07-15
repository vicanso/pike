# pike [![Build Status](https://img.shields.io/travis/vicanso/pike.svg?label=linux+build)](https://travis-ci.org/vicanso/pike)


HTTP缓存服务，提供高效简单的HTTP缓存服务。

一直以来都使用`varnish`来做HTTP缓存，喜欢它的性能高效与vcl配置的强大。在规范化缓存的配置之后，强大的vcl对于我也没有太多的作用了，此时我更希望易上手，更简洁的配置，`Pike`则由此诞生。

## 测试命令

go test -race -coverprofile=test.out ./... && go tool cover --html=test.out

go test -v -bench=".*" ./benchmark

## 构建命令

生成后台管理文件

```bash
cd admin \
  && yarn \
  && yarn build \
  && rm ./dist/js/*.map \
  && cd .. \
  && packr -z
```

## 中间件

中间件自定义Cotenxt新增参数有：

- `status`: 该请求对应的状态（必须）
- `identity`: 该请求对应的id（如果是Pass的请求则无此参数）
- `director`: 该请求对应的director
- `resp`: 该请求的响应数据（包括HTTP状态码，响应头，响应内容）
- `createdAt`: 记录context的创建时间
- `serverTiming`: 用于记录生成各中间件处理时长的server timing
- `fresh`: 根据HTTP请求头与响应头判断数据是否为fresh

### Initialization

- 设置公共请求、响应头
- 处理请求数+1，当前处理并发数+1
- 如果当前处理并发数大于最大值（默认为256 * 1000），则返回出错
- 将请求交至下一中间件（在所有中间件处理完成时，当前处理并发数-1）

### Identifier

- 判断该请求是否为Pass(非GET与HEAD请求)，如果是则跳至下一中间件
- 生成请求的唯一ID (method + host + requestURI)
- 获取该ID对应的请求状态（fetching, hitforpas cacheable）
- 设置ID与Status至Context中


## DirectorPicker

- 如果请求是`Cacheable`，直接从缓存中读取，跳过director picker
- 根据Host与Request从配置的director列表中选择符合的director
- 设置符合的director至Context中

## CacheFetcher

- 如果该请求对应的状态不是`cacheable`，则跳至下一中间件
- 从缓存数据库中读取该请求对应的响应数据
- 设置响应数据至Context中

## Proxy

- 如果该请求已经从缓存中获取数据，则跳至下一中间件
- 根据director配置的backend选择算法，选择符合的可用backend
- 将当前请求转发至backend，获取响应数据，并设置至Context中

## HeaderSetter

- 从响应数据中获取响应的响应头，设置至Response.Header中

## FreshChecker

- 判断请求是否GET或者HEAD，如果否，则跳至下一中间件
- 判断响应状态码是否 < 200 或者 >= 400，如果是，则跳至下一中间件
- 根据请求头与响应头，判断客户端缓存的数据是否为`fresh`
- 设置`fresh`状态至Context中

## Dispatcher

- 从Context中获取Response
- 如果该请求状态为cacheable，设置HTTP Response Header:Age
- 设置HTTP Response Header:X-Status 
- 判断该请求是否非(cacheable与pass)，根据TTL生成hitForPass或者Cacheable状态写入缓存数据库（新的goroutine）
- 判断`fresh`状态，如果是则直接返回NotModified
- 设置HTTP状态码，根据AcceptEncoding生成响应数据并返回


## 性能测试

测试机器：8核 8GB内存，测试环境有限，wrk与测试程序均在同一机器上运行

### Ping测试

Pike的health check，无其它处理逻辑，返回200

```bash
wrk -H 'Accept-Encoding: gzip, deflate' -t10 -c2000 \
-d1m 'http://127.0.0.1:3015/ping' --latency --timeout 3s

Running 1m test @ http://127.0.0.1:3015/ping
  10 threads and 2000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    33.29ms   50.93ms   1.56s    99.22%
    Req/Sec     6.54k   656.09    11.02k    84.68%
  Latency Distribution
     50%   29.03ms
     75%   32.89ms
     90%   37.99ms
     99%   70.45ms
  3906289 requests in 1.00m, 447.04MB read
Requests/sec:  65003.93
Transfer/sec:      7.44MB
```

### 获取可缓存请求

对于可缓存请求的压测，主要三种情况，客户端不支持压缩、支持gzip压缩、支持br压缩

#### 客户端不支持压缩

由于默认缓存的数据只有gzip与br两份数据（现在的客户端都支持两种压缩之一），因此如果客户端不支持，需要将gzip解压返回，性能有所损耗，平均每个请求的数据量为：24KB。

```bash
wrk -t10 -c2000 -d1m 'http://127.0.0.1:3015/api/i18ns' --latency --timeout 3s

Running 1m test @ http://127.0.0.1:3015/api/i18ns
  10 threads and 2000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   373.53ms  460.48ms   3.00s    85.21%
    Req/Sec   832.91    209.89     2.04k    71.07%
  Latency Distribution
     50%  246.35ms
     75%  603.95ms
     90%    1.01s
     99%    1.94s
  497138 requests in 1.00m, 11.14GB read
Requests/sec:   8273.95
Transfer/sec:    189.79MB
```

### 客户端支持gzip压缩

缓存数据中有gzip与br数据，因此无需要重新做压缩，性能较高，平均每个请求的数据量为：5KB。

```bash
wrk -H 'Accept-Encoding: gzip, deflate' -t10 -c2000 \
-d1m 'http://127.0.0.1:3015/api/i18ns' --latency --timeout 3s

Running 1m test @ http://127.0.0.1:3015/api/i18ns
  10 threads and 2000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   203.61ms  226.35ms   3.00s    85.52%
    Req/Sec     1.31k   240.96     3.15k    74.11%
  Latency Distribution
     50%  157.26ms
     75%  315.03ms
     90%  506.21ms
     99%  963.79ms
  785711 requests in 1.00m, 4.13GB read
Requests/sec:  13075.99
Transfer/sec:     70.40MB
```

### 客户端支持br压缩

缓存数据中有gzip与br数据，因此无需要重新做压缩，性能较高，平均每个请求的数据量为：4KB。

```bash
wrk -H 'Accept-Encoding: br, gzip, deflate' -t10 -c2000 \
-d1m 'http://127.0.0.1:3015/api/i18ns' --latency --timeout 3s

Running 1m test @ http://127.0.0.1:3015/api/i18ns
  10 threads and 2000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   199.74ms  222.63ms   2.30s    85.41%
    Req/Sec     1.33k   292.32     3.47k    79.24%
  Latency Distribution
     50%  154.28ms
     75%  309.51ms
     90%  498.98ms
     99%  946.04ms
  795014 requests in 1.00m, 3.58GB read
Requests/sec:  13230.77
Transfer/sec:     61.09MB
```

上面的测试中，`pike`是以docker的形式运行，`docker-proxy`会占用了部分的CPU资源，下面再测试非docker下的性能测试（只测试gzip）:

```bash
wrk -H 'Accept-Encoding: gzip, deflate' -t10 -c2000 \
-d1m 'http://127.0.0.1:3015/api/i18ns' --latency --timeout 3s

Running 1m test @ http://127.0.0.1:3015/api/i18ns
  10 threads and 2000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   212.12ms  277.83ms   2.99s    86.40%
    Req/Sec     1.71k   441.18     4.68k    70.75%
  Latency Distribution
     50%  130.96ms
     75%  330.50ms
     90%  569.07ms
     99%    1.19s
  1018914 requests in 1.00m, 5.36GB read
Requests/sec:  16965.24
Transfer/sec:     91.34MB
```

可以看出，`Pike`的性能已经能满足大部分的网站了，虽然达不到`varnish`那么强悍(docker形式运行，同样的请求可达30K Requests/sec)，但是配置简单更多，有简便的管理后台，如果有兴趣试用的，请联系我~在此，感恩不言谢！

## 流程图

![](./process.png)
