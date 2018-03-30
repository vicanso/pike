# 0.2.0

* 增加智能选择`gzip`与`brotli`压缩算法

# 0.1.3

* 未配置`ping`的backend判断是否可用调整为判断配置端口是否监听
* 增加docker health check
* 支持压缩级别`compressLevel`的配置
* 增加`jpeg`压缩质量配置

# 0.1.2

* 日志格式化字符串时，正则调整为非贪婪匹配，修复紧贴的placeholder无法解析
* 支持最小压缩数据长度配置
* HTTP访问日志增加`payload-size`的支持

# 0.1.1

* proxy响应数据增加对`deflate`压缩的支持
* 调整health的检测，更快的检测可用backend
* 使用`packr`来处理静态文件
* 调整`Server-Timing`的生成处理，增加`proxy`与`gzip`的耗时记录
