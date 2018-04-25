# pike


## 中间件

中间件设置的Cotenxt参数有：

- `status`: 该请求对应的状态（必须）
- `identity`: 该请求对应的id（如果是Pass的请求则无此参数）
- `director`: 该请求对应的director
- `response`: 该请求的响应数据（包括HTTP状态码，响应头，响应内容）


### Identifier

- 判断该请求是否为Pass(非GET与HEAD请求)，如果是则跳至下一中间件
- 生成请求的唯一ID (method + host + requestURI)
- 获取该ID对应的请求状态（fetching, hitforpas cacheable）
- 设置ID与Status至Context中


## DirectorPicker

- 根据Host与Request从配置的director列表中选择符合的director
- 设置符合的director至Context中

## CacheFetcher

- 如果该请求对应的状态不是`cacheable`，则跳至下一中间件
- 从缓存数据库中读取该请求对应的响应数据
- 设置响应数据至Context中

## ProxyWithConfig

- 如果该请求已经从缓存中获取数据，则跳至下一中间件
- 根据director配置的backend选择算法，选择符合的可用backend
- 将当前请求转发至backend，获取数据
- 生成响应数据，并设置至Context中

## Dispatcher

- 从Context中获取Response
- 将Response.Header中的数据设置至响应头
- 判断该请求是否非(cacheable与pass)，根据TTL生成hitForPass或者Cacheable状态写入缓存数据库
- 如果该请求状态为cacheable，设置HTTP Response Header:Age
- 设置HTTP Response Header:X-Status 
- 根据请求头与响应头，判断该请求是否304，如果是则直接返回NotModified
- 设置HTTP状态码，根据AcceptEncoding生成响应数据并返回
