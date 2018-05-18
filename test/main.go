package main

import (
	"net/http"
	"sync/atomic"

	"github.com/labstack/echo"
)

var postCount, putCount, patchCount, deleteCount int32

var noCacheCount, noStoreCount, privateCacheCount, maxAgeZeroCount, cacheableCount int32

func main() {
	e := echo.New()

	e.GET("/ping", func(c echo.Context) error {
		return c.String(200, "pong")
	})

	// 注册用户
	e.POST("/users", func(c echo.Context) error {
		// 注册处理
		data := make(map[string]interface{})
		data["account"] = "vicanso"
		data["count"] = atomic.AddInt32(&postCount, 1)
		return c.JSON(http.StatusOK, data)
	})
	// 账号信息更新（全数据替换）
	e.PUT("/users/:uid", func(c echo.Context) error {
		// 数据替换处理
		data := make(map[string]interface{})
		data["count"] = atomic.AddInt32(&putCount, 1)
		return c.JSON(http.StatusOK, data)
	})
	// 账户信息更新（部分更新）
	e.PATCH("/users/:uid", func(c echo.Context) error {
		// 数据更新
		data := make(map[string]interface{})
		data["count"] = atomic.AddInt32(&patchCount, 1)
		return c.JSON(http.StatusOK, data)
	})
	// 账户删除
	e.DELETE("/users/:uid", func(c echo.Context) error {
		// 删除
		data := make(map[string]interface{})
		data["count"] = atomic.AddInt32(&deleteCount, 1)
		return c.JSON(http.StatusOK, data)
	})

	// GET请求，不可缓存
	e.GET("/no-cache", func(c echo.Context) error {
		header := c.Response().Header()
		header.Set("Cache-Control", "no-cache")
		data := make(map[string]interface{})
		data["count"] = atomic.AddInt32(&noCacheCount, 1)
		return c.JSON(http.StatusOK, data)
	})

	// GET请求，不可存储
	e.GET("/no-store", func(c echo.Context) error {
		header := c.Response().Header()
		header.Set("Cache-Control", "no-store")
		data := make(map[string]interface{})
		data["count"] = atomic.AddInt32(&noStoreCount, 1)
		return c.JSON(http.StatusOK, data)
	})

	// GET请求，私有缓存
	e.GET("/private-cache", func(c echo.Context) error {
		header := c.Response().Header()
		header.Set("Cache-Control", "private, max-age=10")
		data := make(map[string]interface{})
		data["count"] = atomic.AddInt32(&privateCacheCount, 1)
		return c.JSON(http.StatusOK, data)
	})

	// GET请求，max-age:0
	e.GET("/max-age-zero", func(c echo.Context) error {
		header := c.Response().Header()
		header.Set("Cache-Control", "max-age=0")
		data := make(map[string]interface{})
		data["count"] = atomic.AddInt32(&maxAgeZeroCount, 1)
		return c.JSON(http.StatusOK, data)
	})

	// GET请求，可缓存
	e.GET("/cacheable", func(c echo.Context) error {
		header := c.Response().Header()
		header.Set("Cache-Control", "max-age=10")
		data := make(map[string]interface{})
		data["count"] = atomic.AddInt32(&cacheableCount, 1)
		return c.JSON(http.StatusOK, data)
	})

	e.Logger.Fatal(e.Start(":5018"))
}
