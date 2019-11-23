# Pike

## Config

- `cache` 缓存配置
- `compress` 压缩配置
- `location` location配置，需要依赖`upstream`
- `server` HTTP服务配置，需要依赖`cache`，`compress`，`location`
- `upstream` upstream配置