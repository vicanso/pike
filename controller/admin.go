package controller

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"path"
	"strconv"

	"github.com/gobuffalo/packr"
	"github.com/vicanso/pike/performance"
	"github.com/vicanso/pike/proxy"
	"github.com/vicanso/pike/util"

	"github.com/labstack/echo"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/vars"
)

// Serve 静态文件处理
func Serve(c echo.Context) error {
	box := packr.NewBox("../admin/dist")
	file, _ := c.Get(vars.StaticFile).(string)
	buf, err := box.MustBytes(file)
	if err != nil {
		return err
	}
	ext := path.Ext(file)
	header := c.Response().Header()
	contentType := ""
	doGzip := func() {
		gzip, err := util.Gzip(buf, 0)
		if err == nil {
			buf = gzip
			header.Set(echo.HeaderContentEncoding, vars.GzipEncoding)
		}
	}
	setMaxAge := func(age int) {
		header.Set(vars.CacheControl, fmt.Sprintf("public, max-age=%d", age))
	}
	oneYear := 365 * 24 * 3600
	switch ext {
	case ".js":
		contentType = "application/javascript; charset=UTF-8"
		doGzip()
		setMaxAge(oneYear)
	case ".css":
		contentType = "text/css; charset=UTF-8"
		doGzip()
		setMaxAge(oneYear)
	case ".ttf":
		contentType = "application/octet-stream"
	}
	header.Set(echo.HeaderContentLength, strconv.Itoa(len(buf)))
	return c.Blob(http.StatusOK, contentType, buf)
}

// GetStats 获取系统性能统计
func GetStats(c echo.Context) error {
	client := c.Get(vars.CacheClient).(*cache.Client)
	stats := performance.GetStats(client)
	return c.JSON(http.StatusOK, stats)
}

// GetDirectors 获取directors列表
func GetDirectors(c echo.Context) error {
	directors := c.Get(vars.Directors).(proxy.Directors)
	m := make(map[string]interface{})
	m["directors"] = directors
	return c.JSON(http.StatusOK, m)
}

// GetCachedList 获取缓存数据列表
func GetCachedList(c echo.Context) error {
	client := c.Get(vars.CacheClient).(*cache.Client)
	cachedList := client.GetCachedList()
	m := make(map[string]interface{})
	m["cacheds"] = cachedList
	return c.JSON(http.StatusOK, m)
}

// RemoveCached 删除缓存
func RemoveCached(c echo.Context) error {
	key, err := base64.StdEncoding.DecodeString(c.Param("key"))
	if err != nil {
		return err
	}
	client := c.Get(vars.CacheClient).(*cache.Client)
	return client.Remove(key)
}
