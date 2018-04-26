package customMiddleware

import (
	"fmt"
	"net/http"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/vars"

	"github.com/labstack/echo"
)

// echo 默认的http error处理，增加对feching状态的设置

// CreateErrorHandler  创建异常处理函数
func CreateErrorHandler(e *echo.Echo, client *cache.Client) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		iStatus := c.Get(vars.Status)
		// 如果status状态为fetching，则设置此请求为hit for pass
		if iStatus != nil && iStatus.(int) == cache.Fetching {
			iIdentity := c.Get(vars.Identity)
			if iIdentity != nil {
				client.HitForPass(iIdentity.([]byte), vars.HitForPassTTL)
			}
		}
		var (
			code = http.StatusInternalServerError
			msg  interface{}
		)

		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			msg = he.Message
			if he.Internal != nil {
				msg = fmt.Sprintf("%v, %v", err, he.Internal)
			}
		} else if e.Debug {
			msg = err.Error()
		} else {
			msg = http.StatusText(code)
		}
		if _, ok := msg.(string); ok {
			msg = map[string]interface{}{"message": msg}
		}

		e.Logger.Error(err)
		// Send response
		if !c.Response().Committed {
			if c.Request().Method == echo.HEAD { // Issue #608
				err = c.NoContent(code)
			} else {
				err = c.JSON(code, msg)
			}
			if err != nil {
				e.Logger.Error(err)
			}
		}
	}
}
