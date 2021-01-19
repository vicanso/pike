# pike

[![Build Status](https://github.com/vicanso/pike/workflows/Test/badge.svg)](https://github.com/vicanso/pike/actions)

与varnish类似的HTTP缓存服务器，主要的特性如下：

- 提供WEB的管理配置界面，简单易上手
- 支持br与gzip两种压缩方式，根据客户端自动选择。对于可缓存与不可缓存请求使用不同的压缩配置，更佳的时间与空间的平衡
- 仅基于`Cache-Control`生成缓存有效期，接口缓存完全由接口开发者决定，准确而高效(开发比运维更清楚接口是否可缓存，可缓存时长)
- 配置支持文件与etcd两种形式存储，无中断的配置实时更新
- 支持H2C的转发，提升与后端服务的调用性能（如果是内网转发，不需要启用）
- 与upstream的调用支持`gzip`，`brotli`，`lz4`，`snappy`以及`zstd`压缩，可根据与upstream的网络线路选择合适的压缩方式
- 支持upstream检测失败时回调配置，可及时获取异常upstream


<p align="center">
<img src="./docs/images/home.png"/>
</p>

## 启动方式

启动参数主要如下：

- `config` 配置保存地址，可以指定为etcd或者本地文件，如：`etcd://user:pass@127.0.0.1:2379/pike`，本地文件：`/opt/pike/config.yml`
- `admin` 配置管理后台的访问地址，如：`--admin=:9013`
- `log` 日志文件目录，支持单文件与lumberjack形式，如`/var/pike.log`或`lumberjack:///tmp/pike.log?maxSize=100&maxAge=1&compress=true`，lumberjack会根据文件内容大小与时间将文件分割

### 使用文件保存配置

```bash
# linux etcd，管理后台使用9013端口访问
./pike --config=etcd://127.0.0.1:2379/pike --admin=:9013
# linux file，配置文件保存在/opt/pike.yml，管理后台使用9013端口访问
./pike --config=/opt/pike.yml --admin=:9013

# docker
docker run -it --rm \
    -p 9013:9013 \
    vicanso/pike:4.0.0-alpha --config=etcd://172.16.183.177:2379/pike --admin=:9013
```

## build

编译需要依赖`packr2`，需要先执行脚本安装：

```bash
go get -u github.com/gobuffalo/packr/v2/packr2 
```

后执行`make build`则可编译当前系统版本程序

## TODO

- 缓存查询（如果缓存量较大，有可能导致查询性能较差，暂时未支持）