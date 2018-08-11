package controller

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"path"
	"strings"
	"sync/atomic"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/performance"

	"github.com/gobuffalo/packr"
	"github.com/vicanso/pike/pike"
	"github.com/vicanso/pike/util"
)

const (
	statsURL         = "/stats"
	directorsURL     = "/directors"
	cachesURL        = "/cacheds"
	fetchingsURL     = "/fetchings"
	cacheRemoveURL   = "/cacheds/"
	togglePingURL    = "/toggle/ping"
	pingIsDiabledURL = "/ping/is-disabled"
	adminToken       = "X-Admin-Token"
	defaultHTMLFile  = "/index.html"
)

var (
	// ErrTokenInvalid token校验失败
	ErrTokenInvalid = pike.NewHTTPError(http.StatusUnauthorized, "token is invalid")
)

type (
	// AdminConfig admin config
	AdminConfig struct {
		Prefix       string
		Token        string
		Client       *cache.Client
		Directors    pike.Directors
		DisabledPing *int32
	}
)

// serve 静态文件处理
func serve(c *pike.Context, file string) error {
	box := packr.NewBox("../admin/dist")
	buf, err := box.MustBytes(file)
	if err != nil {
		return err
	}
	ext := path.Ext(file)
	resp := c.Response
	header := resp.Header()
	contentType := ""
	doGzip := func() {
		gzip, err := util.Gzip(buf, 0)
		if err == nil {
			buf = gzip
			header.Set(pike.HeaderContentEncoding, pike.GzipEncoding)
		}
	}
	setMaxAge := func(age int) {
		header.Set(pike.HeaderCacheControl, fmt.Sprintf("public, max-age=%d", age))
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

	resp.WriteHeader(http.StatusOK)
	header.Set(pike.HeaderContentType, contentType)
	_, err = resp.Write(buf)
	return err
}

// getStats 获取系统性能统计
func getStats(c *pike.Context, client *cache.Client) error {
	stats := performance.GetStats(client)
	return c.JSON(stats, http.StatusOK)
}

// getDirectors 获取directors列表
func getDirectors(c *pike.Context, directors pike.Directors) error {
	m := make(map[string]interface{})
	m["directors"] = directors
	return c.JSON(m, http.StatusOK)
}

// getCachedList 获取缓存数据列表
func getCachedList(c *pike.Context, client *cache.Client) error {
	cachedList := client.GetCachedList()
	m := make(map[string]interface{})
	m["cacheds"] = cachedList
	return c.JSON(m, http.StatusOK)
}

// getFetchingList 获取fetching的列表
func getFetchingList(c *pike.Context, client *cache.Client) error {
	fetchingList := client.GetFetchingList()
	m := make(map[string]interface{})
	m["fetchings"] = fetchingList
	return c.JSON(m, http.StatusOK)
}

// removeCached 删除缓存
func removeCached(c *pike.Context, client *cache.Client, key string) error {
	k, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return err
	}
	err = client.Remove(k)
	if err != nil {
		return err
	}
	c.Response.WriteHeader(http.StatusNoContent)
	return nil
}

// togglePing 切换ping的状态
func togglePing(c *pike.Context, addr *int32) error {
	// 0表示非禁用，非0表示禁用
	v := atomic.LoadInt32(addr)
	newValue := int32(0)
	if v == 0 {
		newValue = 1
	}
	atomic.StoreInt32(addr, newValue)
	m := make(map[string]interface{})
	m["disabled"] = newValue != 0
	return c.JSON(m, http.StatusOK)
}

// getPingIsDisabeld 获取ping是否禁用
func getPingIsDisabeld(c *pike.Context, addr *int32) error {
	// 0表示非禁用，非0表示禁用
	v := atomic.LoadInt32(addr)
	m := make(map[string]interface{})
	m["disabled"] = v != 0
	return c.JSON(m, http.StatusOK)
}

// AdminHandler admin handler
func AdminHandler(config AdminConfig) pike.Middleware {
	prefix := config.Prefix
	client := config.Client
	directors := config.Directors
	return func(c *pike.Context, next pike.Next) error {
		req := c.Request
		uri := req.URL.Path
		if !strings.HasPrefix(uri, prefix) {
			return next()
		}
		uri = uri[len(prefix):]
		if uri == "" {
			c.Response.Committed = true
			http.Redirect(c.ResponseWriter, c.Request, prefix+"/", http.StatusTemporaryRedirect)
			return nil
		}
		if uri == "/" {
			uri = defaultHTMLFile
		}
		ext := path.Ext(uri)
		// 静态文件不校验token
		if len(ext) != 0 {
			return serve(c, uri[1:])
		}
		if req.Header.Get(adminToken) != config.Token {
			return ErrTokenInvalid
		}
		switch uri {
		case statsURL:
			return getStats(c, client)
		case directorsURL:
			return getDirectors(c, directors)
		case cachesURL:
			return getCachedList(c, client)
		case fetchingsURL:
			return getFetchingList(c, client)
		case togglePingURL:
			return togglePing(c, config.DisabledPing)
		case pingIsDiabledURL:
			return getPingIsDisabeld(c, config.DisabledPing)
		}
		if strings.HasPrefix(uri, cacheRemoveURL) {
			key := uri[len(cacheRemoveURL):]
			return removeCached(c, client, key)
		}
		return nil
	}
}
