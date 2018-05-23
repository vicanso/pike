# pike [![Build Status](https://img.shields.io/travis/vicanso/pike.svg?label=linux+build)](https://travis-ci.org/vicanso/pike)


HTTP缓存服务，提供高效简单的HTTP缓存服务。

一直以来都使用`varnish`来做HTTP缓存，喜欢它的性能高效与vcl配置的强大。在规范化缓存的配置之后，强大的vcl对于我也没有太多的作用了，此时我更希望易上手，更简洁的配置，`Pike`则由此诞生。

## 测试命令

go test -race -coverprofile=test.out ./... && go tool cover --html=test.out

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

中间件设置的Cotenxt参数有：

- `status`: 该请求对应的状态（必须）
- `identity`: 该请求对应的id（如果是Pass的请求则无此参数）
- `director`: 该请求对应的director
- `response`: 该请求的响应数据（包括HTTP状态码，响应头，响应内容）
- `timing`: 记录处理时长，生成Server-Timing
- `fresh`: 根据HTTP请求头与响应头判断数据是否为fresh

### Initialization

- 设置公共响应头
- 处理请求数+1，当前处理并发数+1
- 如果当前处理并发数大于最大值（默认为256 * 1000），则返回出错
- 将请求交至下一中间件（在所有中间件处理完成时，当前处理并发数-1）

### Identifier

- 生成Timing并添加至Context中
- 判断该请求是否为Pass(非GET与HEAD请求)，如果是则跳至下一中间件
- 生成请求的唯一ID (method + host + requestURI)
- 获取该ID对应的请求状态（fetching, hitforpas cacheable）
- 设置ID与Status至Context中


## DirectorPicker

- 根据Host与Request从配置的director列表中选择符合的director
- 设置符合的director至Context中

## CacheFetcher

- 如果该请求对应的状态不是`cacheable`，则跳至下一中间件
- 从缓存数据库中读取该请求对应的响应数据并记录耗时
- 设置响应数据至Context中

## Proxy

- 如果该请求已经从缓存中获取数据，则跳至下一中间件
- 根据director配置的backend选择算法，选择符合的可用backend
- 将当前请求转发至backend，获取数据并记录耗时
- 生成响应数据，并设置至Context中

## HeaderSetter

- 从响应数据中获取响应数据，设置至Responser.Header中

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

### Ping测试

Pike的health check，无其它处理逻辑，返回200

```bash
wrk -H 'Accept-Encoding: gzip, deflate' -t10 -c200 \
-d1m 'http://127.0.0.1:3015/ping' --latency
```

```bash
10 threads and 200 connections
Thread Stats   Avg      Stdev     Max   +/- Stdev
  Latency     3.54ms    2.02ms  32.63ms   74.82%
  Req/Sec     5.78k   630.05    33.10k    78.33%
Latency Distribution
    50%    3.20ms
    75%    4.43ms
    90%    6.00ms
    99%   10.38ms
3450900 requests in 1.00m, 394.92MB read
Requests/sec:  57421.72
Transfer/sec:      6.57MB
```

### 获取可缓存请求

对于可缓存请求的压测，主要三种情况，客户端不支持压缩、支持gzip压缩、支持br压缩（在我自己的HP gen8 做的压力测试）

#### 客户端不支持压缩

由于默认缓存的数据只有gzip与br两份数据（现在的客户端都支持两种压缩之一），因此如果客户端不支持，需要将gzip解压返回，性能有所损耗，平均每个请求的数据量为：75KB。

```bash
wrk -t10 -c200 -d1m 'http://127.0.0.1:3015/css/app.f81943d4.css' --latency
```

```bash
10 threads and 200 connections
Thread Stats   Avg      Stdev     Max   +/- Stdev
  Latency    66.49ms   69.82ms   1.03s    97.56%
  Req/Sec   337.83     46.65   666.00     74.90%
Latency Distribution
    50%   62.98ms
    75%   76.56ms
    90%   96.10ms
    99%  344.35ms
199258 requests in 1.00m, 14.41GB read
Requests/sec:   3318.15
Transfer/sec:    245.70MB
```

### 客户端支持gzip压缩

缓存数据中有gzip与br数据，因此无需要重新做压缩，性能较高，平均每个请求的数据量为：23KB。

```bash
wrk -H 'Accept-Encoding: gzip, deflate' -t10 -c200 \
-d1m 'http://127.0.0.1:3015/css/app.f81943d4.css' --latency
```

```bash
10 threads and 200 connections
Thread Stats   Avg      Stdev     Max   +/- Stdev
  Latency    19.54ms    8.67ms  74.71ms   77.84%
  Req/Sec     1.03k    95.60     1.56k    72.60%
Latency Distribution
    50%   21.60ms
    75%   23.47ms
    90%   26.54ms
    99%   37.29ms
615208 requests in 1.00m, 13.59GB read
Requests/sec:  10245.77
Transfer/sec:    231.70MB
```

### 客户端支持br压缩

缓存数据中有gzip与br数据，因此无需要重新做压缩，性能较高，平均每个请求的数据量为：21KB。

```bash
wrk -H 'Accept-Encoding: br, gzip, deflate' -t10 -c200 \
-d1m 'http://127.0.0.1:3015/css/app.f81943d4.css' --latency
```

```bash
10 threads and 200 connections
Thread Stats   Avg      Stdev     Max   +/- Stdev
  Latency    19.54ms    8.65ms  69.72ms   77.90%
  Req/Sec     1.03k    93.44     1.54k    73.88%
Latency Distribution
    50%   21.59ms
    75%   23.36ms
    90%   26.47ms
    99%   37.22ms
615076 requests in 1.00m, 12.40GB read
Requests/sec:  10243.72
Transfer/sec:    211.52MB
```