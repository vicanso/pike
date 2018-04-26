package customMiddleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/vicanso/dash"
	"github.com/vicanso/fresh"

	"github.com/labstack/echo"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/util"
	"github.com/vicanso/pike/vars"
)

var ignoreKeys = []string{
	"Date",
}

func save(client *cache.Client, identity []byte, resp *cache.Response) {
	doSave := func() {
		client.SaveResponse(identity, resp)
		client.Cacheable(identity, resp.TTL)
	}
	level := resp.CompressLevel
	compressMinLength := resp.CompressMinLength
	if compressMinLength == 0 {
		compressMinLength = vars.CompressMinLength
	}
	if resp.StatusCode == http.StatusNoContent {
		doSave()
		return
	}
	body := resp.Body
	// 如果body为空，但是gzipBody不为空，表示从backend取回来的数据已压缩
	if len(body) == 0 && len(resp.GzipBody) != 0 {
		// 解压gzip数据，用于生成br
		unzipBody, err := util.Gunzip(resp.GzipBody)
		if err != nil {
			doSave()
			return
		}
		body = unzipBody
	}
	bodyLength := len(body)
	// 204没有内容的情况已处理，不应该出现 body为空的现象
	// 如果原始数据还是为空，则直接设置为hit for pass
	if bodyLength == 0 {
		client.HitForPass(identity, vars.HitForPassTTL)
		return
	}
	// 如果数据比最小压缩还小，不需要压缩缓存
	if bodyLength < compressMinLength {
		doSave()
		return

	}
	if len(resp.GzipBody) == 0 {
		gzipBody, _ := util.Gzip(body, level)
		// 如果gzip压缩成功，可以删除原始数据，只保留gzip
		if len(gzipBody) != 0 {
			resp.GzipBody = gzipBody
			resp.Body = nil
		}
	}
	if len(resp.BrBody) == 0 {
		resp.BrBody, _ = util.Brotli(body, level)
	}
	doSave()
	return
}

// Dispatcher 对响应数据做缓存，复制等处理
func Dispatcher(client *cache.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			iStatus := c.Get(vars.Status)
			if iStatus == nil {
				return vars.ErrRequestStatusNotSet
			}
			iResponse := c.Get(vars.Response)
			if iResponse == nil {
				return vars.ErrResponseStatusNotSet
			}

			status := iStatus.(int)
			cr := iResponse.(*cache.Response)
			resp := c.Response()
			reqHeader := c.Request().Header

			h := resp.Header()
			for k, values := range cr.Header {
				if dash.IncludesString(ignoreKeys, k) {
					continue
				}
				for _, v := range values {
					h.Add(k, v)
				}
			}

			// 如果数据不是从缓存中读取，都需要判断是否需要写入缓存
			// pass的都是不可能缓存
			if status != cache.Cacheable && status != cache.Pass {
				go func() {
					iIdentity := c.Get(vars.Identity)
					if iIdentity == nil {
						return
					}
					identity := iIdentity.([]byte)
					if cr.TTL == 0 {
						client.HitForPass(identity, vars.HitForPassTTL)
					} else {
						save(client, identity, cr)
					}
				}()
			} else if status == cache.Cacheable {
				// 如果数据是读取缓存，有需要设置Age
				age := uint32(time.Now().Unix()) - cr.CreatedAt
				h.Set(vars.Age, strconv.Itoa(int(age)))
			}

			xStatus := ""
			switch status {
			case cache.Pass:
				xStatus = vars.Pass
			case cache.Fetching:
				xStatus = vars.Fetching
			case cache.HitForPass:
				xStatus = vars.HitForPass
			default:
				xStatus = vars.Cacheable
			}
			h.Set(vars.XStatus, xStatus)
			statusCode := int(cr.StatusCode)
			method := c.Request().Method
			if method == echo.GET || method == echo.HEAD {
				if (statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices) || statusCode == http.StatusNotModified {
					ifModifiedSince := reqHeader.Get(echo.HeaderIfModifiedSince)
					ifNoneMatch := reqHeader.Get(vars.IfNoneMatch)
					cacheControl := reqHeader.Get(vars.CacheControl)
					requestHeaderData := &fresh.RequestHeader{
						IfModifiedSince: []byte(ifModifiedSince),
						IfNoneMatch:     []byte(ifNoneMatch),
						CacheControl:    []byte(cacheControl),
					}
					eTag := h.Get(vars.ETag)
					lastModified := h.Get(echo.HeaderLastModified)
					respHeaderData := &fresh.ResponseHeader{
						ETag:         []byte(eTag),
						LastModified: []byte(lastModified),
					}
					if fresh.Fresh(requestHeaderData, respHeaderData) {
						resp.WriteHeader(http.StatusNotModified)
						return nil
					}
				}
			}
			body, enconding := cr.GetBody(reqHeader.Get(echo.HeaderAcceptEncoding))
			if enconding != "" {
				h.Set(echo.HeaderContentEncoding, enconding)
			}

			h.Set(echo.HeaderContentLength, strconv.Itoa(len(body)))
			resp.WriteHeader(statusCode)
			_, err := resp.Write(body)
			return err
		}
	}
}
