package custommiddleware

import (
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/util"
	"github.com/vicanso/pike/vars"
)

type (
	// DispatcherConfig dipatcher的配置
	DispatcherConfig struct {
		Skipper middleware.Skipper
		// 压缩数据类型
		CompressTypes []string
		// 最小压缩
		CompressMinLength int
		// CompressLevel 数据压缩级别
		CompressLevel int
	}
)

var (
	defaultCompressTypes = []string{
		"text",
		"javascript",
		"json",
	}
)

func save(client *cache.Client, identity []byte, resp *cache.Response, compressible bool) {
	doSave := func() {
		client.SaveResponse(identity, resp)
		client.Cacheable(identity, resp.TTL)
	}
	level := resp.CompressLevel
	compressMinLength := resp.CompressMinLength
	if compressMinLength == 0 {
		compressMinLength = vars.CompressMinLength
	}
	if resp.StatusCode == http.StatusNoContent || !compressible {
		doSave()
		return
	}
	body := resp.Body
	// 如果body为空，表示从backend取回来的数据已压缩
	if len(body) == 0 {
		// 如果body为空，但是gzipBody不为空，表示从backend取回来的数据已压缩
		if len(resp.GzipBody) != 0 {
			// 解压gzip数据，用于生成br
			unzipBody, err := util.Gunzip(resp.GzipBody)
			if err != nil {
				doSave()
				return
			}
			body = unzipBody
		} else if len(resp.BrBody) != 0 {
			// 如果gzip数据为0，br数据不为0，则表示backend返回的数据为br
			// 解压br数据，用于生成br
			unzipBody, err := util.BrotliDecode(resp.BrBody)
			if err != nil {
				doSave()
				return
			}
			body = unzipBody
		}
	}
	bodyLength := len(body)
	// 204没有内容的情况已处理，不应该出现 body为空的现象（因为程序不支持304的处理，所以在proxy时已删除相应头，也不会出现304）
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
		resp.BrBody, _ = util.BrotliEncode(body, level)
	}
	doSave()
	return
}

func shouldCompress(compressTypes []string, contentType string) (compressible bool) {
	for _, v := range compressTypes {
		reg := regexp.MustCompile(v)
		if reg.MatchString(contentType) {
			compressible = true
			return
		}
	}
	return
}

// Dispatcher 对响应数据做缓存，复制等处理
func Dispatcher(config DispatcherConfig, client *cache.Client) echo.MiddlewareFunc {
	compressTypes := config.CompressTypes
	if len(compressTypes) == 0 {
		compressTypes = defaultCompressTypes
	}
	compressMinLength := config.CompressMinLength
	compressLevel := config.CompressLevel
	// Defaults
	if config.Skipper == nil {
		config.Skipper = middleware.DefaultSkipper
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Logger().Debug("dispatcher middleware")
			if config.Skipper(c) {
				return next(c)
			}
			pc := c.(*Context)
			if pc.Debug {
				c.Logger().Info("dispatcher middleware")
			}
			serverTiming := pc.serverTiming
			done := serverTiming.Start(ServerTimingDispatcher)
			status := pc.status
			cr := pc.resp
			cr.CompressMinLength = compressMinLength
			cr.CompressLevel = compressLevel

			resp := pc.Response()
			respHeader := resp.Header()
			reqHeader := pc.Request().Header

			setSeverTiming := func() {
				done()
				timingStr := serverTiming.String()
				if len(timingStr) != 0 {
					respHeader.Add(vars.ServerTiming, timingStr)
				}
			}

			compressible := shouldCompress(compressTypes, respHeader.Get(echo.HeaderContentType))

			if status == cache.Cacheable {
				// 如果数据是读取缓存，有需要设置Age
				age := uint32(time.Now().Unix()) - cr.CreatedAt
				respHeader.Set(vars.Age, strconv.Itoa(int(age)))
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
			respHeader.Set(vars.XStatus, xStatus)
			statusCode := int(cr.StatusCode)

			// pass的都是不可能缓存
			// 可缓存的处理继续后续缓存流程
			if status != cache.Cacheable && status != cache.Pass {
				identity := pc.identity
				go func() {
					if cr.TTL == 0 {
						if status != cache.HitForPass {
							client.HitForPass(identity, vars.HitForPassTTL)
						}
					} else {
						save(client, identity, cr, compressible)
					}
				}()
			}

			// 304 的处理
			if pc.fresh {
				resp.WriteHeader(http.StatusNotModified)
				setSeverTiming()
				return nil
			}

			acceptEncoding := ""
			// 如果数据不应该被压缩，则直接认为客户端不接受压缩数据
			if compressible {
				acceptEncoding = reqHeader.Get(echo.HeaderAcceptEncoding)
			}
			body, enconding := cr.GetBody(acceptEncoding)
			if enconding != "" {
				respHeader.Set(echo.HeaderContentEncoding, enconding)
			}

			setSeverTiming()
			respHeader.Set(echo.HeaderContentLength, strconv.Itoa(len(body)))
			resp.WriteHeader(statusCode)
			_, err := resp.Write(body)
			return err
		}
	}
}
