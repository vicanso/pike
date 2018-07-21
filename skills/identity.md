# identity请求唯一标识

在HTTP请求中，只有`GET`与`HEAD`请求是可缓存的，因此`Pike`中默认的`identity`是`method host uri`三个参数来生成，可以满足大部分的使用场景。在现实使用中，的确存在各种不同的自定义的需求，下面来讲解一下怎么自定义`identity`的生成规则。

自定义规则支持以下的参数配置：

- `host` 请求的host
- `method` 请求的method
- `path` 请求路径的path部分，不包括querystring
- `proto` 请求的http proto，如HTTP/1.1 HTTP/1.0
- `scheme` 请求的scheme，HTTPS HTTP
- `uri` 请求的request uri，包括querystring
- `userAgent` 请求头的User-Agent
- `query` 请求的querystring
- `~key` 从请求头的cookie中取出key的值
- `>key` 从请求头中取key的值
- `?key` 从querystring中取key的值

如我们配置的`identity`的规则是`host method path proto scheme uri userAgent query ~jt >X-Token ?id`，有如下的一个请求：

```go
	c := &http.Cookie{
		Name:  "jt",
		Value: "HJxX4OOoX7",
	}
  req := httptest.NewRequest(http.MethodGet, "/users/me?cache=no-cache&id=1", nil)
  req.Host = "aslant.site"
	req.Header.Set("User-Agent", "golang-http")
	req.Header.Set("X-Token", "ABCD")
	req.AddCookie(c)
```

此请求对应的`identity`的值是`aslant.site GET /users/me HTTP/1.1 HTTP /users/me?cache=no-cache&id=1 golang-http cache=no-cache&id=1 HJxX4OOoX7 ABCD 1`，每个值以空格分隔，如果使用默认的配置，则生成的值是`GET aslant.site /users/me?cache=no-cache&id=1`。

虽然`Pike`支持自定义的`identity`配置，但是不建议使用自定义的配置，尽可能让程序去兼容标准的处理方式，让`Pike`的配置更简化、标准，便于各种不同的后端应用接入。
