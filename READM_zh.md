# Pike

高效简单的HTTP缓存服务，与varnish类似。

- 标准化的缓存TTL，根据HTTP响应头中的`Cache-Control`来生成
- 简单易用的Web UI
- 支持br与gzip压缩，根据客户端动态选择压缩方式

## 流程图

<p align="center">
<img src="./docs/flow.jpg"/>
</p>

## 相关脚本

### 开发

本项目使用go与nodejs开发，安装完成后执行以下命令则可运行项目。

```bash
# 使用etcd存储配置
go run main.go --config etcd://127.0.0.1:2379/pike --init

# 使用文件存储配置
go run main.go --config /tmp --init
```

```bash
cd web && yarn start
```

然后通过浏览器打开`http://127.0.0.1:3015/`则可。

### 编译

编译时需要使用packr2来打包静态文件，因此需要先安装packr2：

```bash
go get -u github.com/gobuffalo/packr/v2/packr2 
```

```bash
make build-web && make build 
```