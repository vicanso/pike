---
description: 如何开始使用pike
---

Pike是纯go的项目，可以使用各平台的执行文件启动或者已打包好的docker镜像(vicanso/pike)，配置信息支持保存在文件或etcd中，生产环境中建议使用etcd便于多实例部署。


## 启动参数

```bash
Usage:
  pike [flags]

Flags:
      --admin string    The address of admin web page, e.g.: :9013
      --alarm string    The alarm request url, alarm will post to the url, e.g.: http://192.168.1.2:3000/alarms
      --config string   The config of pike, support etcd or file, etcd://user:pass@192.168.1.2:2379,192.168.1.3:2379/pike or /opt/pike (default "pike.yml")
  -h, --help            help for pike
```

如上所示pike的启动参数如下：

- admin 管理后台监听地址，用于启动管理后台，建议最少其中一个实例启用管理后台，方便使用WEB管理后台编辑配置
- alarm 告警回调服务地址，当upstream的服务器检测失败或配置更新失败等时回调，用于告警通知
- config 配置地址，可以用于etcd或者file的形式，建议在生产环境中使用etcd，如果不配置则直接使用文件形式，文件为pike.yml


首次启动指定管理后台监听地址为:9013，未指定配置地址（使用pike.yml)，启动成功后可以看到在执行目录下生成了一个空白的新文件pike.yml（如果该文件已存在则不新建），之后可以打开`http://127.0.0.1:9013/#/`进行配置。

```bash
./pike --admin=:9013
```

## 压缩参数配置

- `Name` 压缩配置名称，用于区分每个压缩配置，可根据不同的应用场景配置不同的压缩参数，一般只使用一个通用配置则可
- `Gzip Level` gzip的压缩级别，如果CPU较为紧张，则可以配置为默认的压缩级别6，如果CPU较为空闲，建议直接配置为最高压缩级别9，减少网络带宽的占用
- `Br Level` brotli的压缩级别，由于br的压缩率较高，占用CPU较大，因此一般配置为6则可，具体根据CPU的使用状况可以选择更优的配置方式
- `Remark` 备注

<p align="center">
<img src="./images/add-compress.png"/>
</p>

需要注意，对于缓存数据的压缩会直接使用默认的`bestCompression`的压缩配置，该配置的压缩级别为gzip:9, br:6，如果需要覆盖默认的配置，则直接新配置名为`bestCompression`的配置则可覆盖。缓存的数据只压缩一次而可使用多次，可以选择较高的压缩级别，对于常规的压缩配置，br的压缩级别配置为6则可，如果CPU占用较多，可以选择更小的值，具体各压缩级别耗时可查看模块-压缩模块的说明。

## 缓存参数配置

- `Name` 缓存配置名称，用于区分每个缓存配置，对于请求量特别大的服务，可单独使用一个缓存，其它的服务则共用一个缓存则可
- `Size` 缓存数量大小，指定LRU缓存的最大数量，可根据服务的缓存情况以及机器内存选择较为合适的值，一般设置为51200已能满足大部分应用的需求，如果内存较少则设置为更小的值
- `HitForPass` 设置hit for pass的缓存时长，对于不可缓存的GET、HEAD请求，为了后续快速判断请求是否hit for pass，缓存中也有保存该请求的缓存状态(hitForPass)。
- `Remark` 备注

为什么会有需要hit for pass的场景？考虑一下以下场景，由于产品刚好被下架处理，因此请求产品详情信息时，该接口返回了出错（http status: 400，cache control: no-cache），因此访问该产品的接口缓存为hit for pass，而后续产品上架了，接口正常响应，缓存时长为cache-control: max-age=60，此时接口应该可缓存的。而由于hit for pass未过期，因此只能等hit for pass过期后接口才变为可缓存。

因此在设置hit for pass的时候需要考虑应用的具体出错处理逻辑，Cache-Control是否无论怎样都不会变化（有一种处理是同样的参数，无论成功失败均使用同样的Cache-Control，这样保证无论成功还是失败，接口均是缓存，避免过多请求），如果是不变的，可以将hit for pass设置为较长的有效期，否则应该选择更短的有效期。

<p align="center">
<img src="./images/add-cache.png"/>
</p>

## Upstream配置

- `Name` upstream的配置名称，用于区分每个upstream配置
- `Health Check` 健康检测的url路径，对于HTTP服务尽量使用特定的url的响应来检测upstream是否可用，如果未配置，则检测地址的端口是否有监听
- `Policy` 服务器列表的选择策略，支持四种方式`roundRobin`，`random`， `first`与`leastConn`，一般选择`roundRobin`则可
- `Enable H2C` 是否启用HTTP/2 over TCP，upstream的服务支持h2c模式，则可以启用此模式，pike与upstream的服务则使用h2c方式访问
- `Accept Encoding` 设置可接受的编码，如果需要节约pike与upstream服务之间访问的网络带宽，可以添加此配置，pike支持编码：`gzip`，`br`，`lz4`，`zst`, 以及`snz`
- `Servers.Addr` 服务地址，以http(s)://ip:port的形式配置
- `Servers.Backup` 是否备用服务地址，如果设置为备用，则只要在主服务有一个可用时，均不会使用备用服务
- `Remark` 备注

<p align="center">
<img src="./images/add-upstream.png"/>
</p>

## Location配置

- `Name` location的配置名称，用于区分每个location配置
- `Upstream` 选择对应的upstream
- `Prefixes` 配置对应的前缀，可配置多个，前缀的匹配优先级高于host
- `Hosts` 配置对应的host，可配置多个
- `Rewrites` 转发请求时，需要重写的URL的规则，配置格式为`key:value`的形式，以`:`分割
- `QueryStrings` 转发请求时添加至url中querystring，配置格式为`key:value`的形式，以`:`分割
- `RespHeaders` 响应头配置，将在所有的响应中添加响应头，配置格式为`key:value`的形式，以`:`分割
- `ReqHeaders` 请求头配置，将在所有的请求中添加请求头，配置格式为`key:value`的形式，以`:`分割
- `ProxyTimeout` 请求超时配置，用于控制请求转发至upstream的服务中的超时，根据实际场景配置，如：30s，1m等等
- `Remark` 备注

<p align="center">
<img src="./images/add-location.png"/>
</p>


### Rewrite规则

重写的规则与nginx类似，支持使用正则匹配，如下面的例子：

- `/api/*:/$1` $1表示*部分，最终的处理就是转发时将/api前缀删除
- `/rest/*/user/*:/$1/$2` $1表示第一个*，$2表示第二个*，最终的处理就是转发的时候将/rest与/user替换

虽然通过正则可以实现各类的重写，但是不建议使用过于复杂的正则，尽可能少用或只用重写来处理前缀，规范url减少重写。

### ENV获取配置

`QueryStrings`，`RespHeaders`以及`ReqHeaders`均支持从ENV中获取值的处理方式，如：`DC:$DC`，$DC表示从ENV中获取DC对应的值。

## Server配置

- `Addr` 监听地址
- `Locations` 对应的location列表
- `Cache` 缓存，根据应用访问量选择合适的缓存
- `Compress` 压缩，根据带宽与CPU的考虑，选择合适的压缩
- `Compress Min Length` 最小压缩长度，此值不要设置太少，因为压缩小数据效果并不明显，而且浪费CPU。一般建议设置为1kb，如果是内网间调用，建议此值可以调更大的值
- `Compress Content Filter` 压缩数据类型筛选，指定针对哪些数据类型压缩，默认值为：`text|javascript|json|wasm|xml`，可按应用的需求自定义配置或不匹配。
- `Log Format` 请求日志格式化配置，如`{remote} {when-iso} {:proxyTarget} {method} {uri} {proto} {status} {size-human} {referer} {userAgent}`，配置规则参考[elton logger](https://github.com/vicanso/elton/blob/master/docs/middlewares.md#logger)，日志的输出对于性能会有所影响
- `Remark` 备注

<p align="center">
<img src="./images/add-server.png"/>
</p>

当Server配置完成后，可以使用`curl http://addr/ping`来检测该Server是否启动成功，`/ping`默认由pike处理而不会转发至upstream

### Location匹配

location列表中，按是否匹配prefix与host排序，其优先级是 prefix+host > prefix > host > 无配置，从列表中按顺序一个个location匹配，匹配则该请求由此location处理，如果所有均不匹配则出错。

## Admin配置

- `Account` 登录账号
- `Password` 登录密码

设置配置成功后重启pike，之后每次使用都需要登录校验，建议在首次配置则设置。

## 缓存列表

暂未支持查询当前缓存列表功能，仅可用于删除缓存

<p align="center">
<img src="./images/caches.png"/>
</p>

## 非实时生效配置

- `缓存配置` 由于缓存是多个LRU组成，因此如果调整缓存大小会导致缓存失败，而且锁的处理也比较麻烦，因此缓存更新非实时生效，只能重启应用
- `Server配置的Log` 日志的输出是在Server创建时生成，如果后续有调整，只能重启应用
- `Admin配置` admin配置非实时生效，因此在初始创建时建议配置
