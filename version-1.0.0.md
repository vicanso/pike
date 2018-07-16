# pike 1.0.0

Pike：HTTP缓存服务，提供高效简单的HTTP缓存服务，类似于varnish但配置更简单。

Pike由最开始基于`fasthttp`，`fasthttp`的性能的确很高效，但该项目在2017年底之后就没有`commit`，提的`issue`也没有反馈，BUG只能自己修复，因此后续切换至`echo`。`echo`的大部分增强的功能都基本没使用到，自带的`http`已足够满足现有的业务场景，最终选择了直接使用自带的`http`，版本1.0.0也正式发布。

简洁的配置：

```yaml
# HTTP response header 中的 Pike 
name: Pike
# 程序监听的端口，默认为 :3015
listen: :3015
# 数据缓存的db文件（必须指定）
db: /tmp/pike.cache
directors:
  -
    # 名称
    name: tiny 
    # backend的健康检测，如果不配置，则默认判断该端口是否被监听
    ping: /ping
    # prefix与host是AND的关系
    # 判断请求url的是否包含该前缀，如果是，则是此director
    prefixs:
      - /api
    # backend列表
    backends:
      - http://127.0.0.1:5018
  -
    name: npmtrend
    ping: /ping
    hosts:
      - npmtrend.com
    backends:
      - http://127.0.0.1:5020
      - http://127.0.0.1:5021
```

## 特性

- 基于yaml的配置，简洁易懂
- WEB管理后台，提供节点监控、系统性能、缓存清理功能
- 标准化的基于HTTP头Cache-Control缓存控制
- 压缩保存的响应数据，避免每次响应时重新压缩（如果客户端不支持压缩则解压）
- 自定义日志格式，支持二十多种placeholder，如：cookie，请求头字段，响应头字段，响应时间等
- 访问日志支持以文件（按天分隔）或者UDP的形式输出
- 支持自定义HTTP请求、响应头配置
- 支持自定义最小压缩长度，对于内网之间的访问，避免压缩、解压的时间损耗
- 支持自定义文本压缩级别与指定压缩数据类型
- 根据客户端智能选择响应数据压缩方式：`gzip`或者`brotli`

## 性能

测试机器：8核 8GB内存，测试环境有限，wrk与测试程序均在同一机器上运行

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

## 安装

因为`pike`支持`br`的压缩处理，此功能需要依赖于动态库，建议直接使用打包好的docker镜像：vicanso/pike:1.0.0，相应编译好的动态库(ubuntu)也可以在github中的release中下载。


```shell
docker run -d --restart=always \
  -p 3015:3015 \
  -v /data/pike/config.yml:/etc/pike/config.yml \
  --name=pike \
  vicanso/pike
```

## 结语

`Pike`在性能已超过10k/rps，对于大部分的网站已经能满足性能上的需求，如果对于性能有更高要求的可以不使用docker的形式执行，或者直接使用`varnish`。`Pike`性能虽然比不上`varnish`，但它的配置更简单，而且也有直观的管理后台功能，如果有兴趣试用，可以在`github`上向我反馈。在此，感恩不言谢！

注：管理后台体验[http://xs.aslant.site:3000/pike/index.html#/](http://xs.aslant.site:3000/pike/index.html)，token是`abcd`