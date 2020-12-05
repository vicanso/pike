# pike


与varnish类似的HTTP缓存服务器，主要的特性如下：

- 提供WEB的管理配置界面
- 支持br与gzip两种压缩方式，根据客户端自动选择
- 无中断的配置实时更新
- 支持H2C的转发，提升与后端服务的调用性能


## build

编译需要依赖`packr2`，需要先执行脚本安装：

```bash
go get -u github.com/gobuffalo/packr/v2/packr2 
```

后执行`make build`则可编译当前系统版本程序

## TODO

- 缓存查询（如果缓存量较大，有可能导致查询性能较差，暂时未支持）