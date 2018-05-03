package controller

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/vicanso/pike/performance"
	"github.com/vicanso/pike/proxy"

	"github.com/labstack/echo"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/vars"
)

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
	buf, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}
	m := make(map[string]string)
	err = json.Unmarshal(buf, &m)
	if err != nil {
		return err
	}
	client := c.Get(vars.CacheClient).(*cache.Client)
	key := []byte(m["key"])
	return client.Remove(key)
}
