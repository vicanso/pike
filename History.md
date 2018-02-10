# 0.1.1

* proxy响应数据增加对`deflate`压缩的支持
* 调整health的检测，更快的检测可用backend
* 使用`packr`来处理静态文件
* 调整`Server-Timing`的生成处理，增加`proxy`与`gzip`的耗时记录
