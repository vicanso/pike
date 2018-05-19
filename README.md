# pike

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
