---
description: 程序处理异常时的各出错信息
---

pike程序处理出错时，均返回category: "pike"的出错信息，可根据此分类判断是否系统错误，主要的错误如下：

- `ErrInvalidResponse` 响应数据异常时使用，主要是程序无法获取正常的响应，http状态码为`503`，出错信息为`Invalid response`
- `ErrCacheDispatcherNotFound` 无法获取配置的缓存时使用，由于配置了不存在的缓存导致，使用管理后台配置时会有相应的校验，因此一般不会触发。http状态码为`503`，出错信息为`Available cache dispatcher not found`
- `ErrLocationNotFound` 无法获取可用的Location，由于配置的Location无符合该请求时触发。http状态码为`503`，出错信息为`Available location not found`
- `ErrUpstreamNotFound` 无法获取可用的Upstream，如果配置的所有Upstream的状态检测均不通过，转发至相应的Upstream时触发。http状态码为`502`，出错信息为`Available upstream not found`
