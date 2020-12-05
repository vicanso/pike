---
description: 告警
---

当程序出现异常时，如更新配置失败、upstream节点异常等，若启动指定了告警的回调地址，程序则会以POST的形式调用告警地址，内容如下： 

```json
{
    "application": "pike",
    "category": "类别",
    "message": "告警消息"
}
```

category有如下的类型：

- `upstream` 当upstream下的某个server检测失败时，则触发告警
- `config` 当更新config失败时，则触发告警
- `admin` 当admin管理后台启动失败时，则触发告警

建议生产使用时，通过告警回调发送短信、邮件等方式及时获取告警内容
