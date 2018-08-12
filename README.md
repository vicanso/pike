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
  && mv ./dist/fonts ./dist/css/ \
  && cd .. \
  && packr -z
```

## 相关文档

- [为什么使用Pike](https://github.com/vicanso/pike/wiki/%E4%B8%BA%E4%BB%80%E4%B9%88%E4%BD%BF%E7%94%A8Pike)
- [Pike配置详解](https://github.com/vicanso/pike/wiki/Pike%E9%85%8D%E7%BD%AE%E8%AF%A6%E8%A7%A3)
