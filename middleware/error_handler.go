package middleware

import (
	"fmt"
	"net/http"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/pike"
	"github.com/vicanso/pike/util"
)

// echo 默认的http error处理，增加对feching状态的设置

// CreateErrorHandler  创建异常处理函数
func CreateErrorHandler(client *cache.Client) pike.ErrorHandler {
	return func(err error, c *pike.Context) {
		// 如果出错的请求，都设置为hit for pass
		key := util.GetIdentity(c.Request)
		client.HitForPass(key, HitForPassTTL)
		if c.Response.Committed {
			return
		}

		var (
			code = http.StatusInternalServerError
			msg  interface{}
		)

		if he, ok := err.(*pike.HTTPError); ok {
			code = he.Code
			msg = he.Message
			if he.Internal != nil {
				msg = fmt.Sprintf("%v, %v", err, he.Internal)
			}
		} else {
			msg = http.StatusText(code)
		}

		c.ResponseWriter.WriteHeader(code)
		c.ResponseWriter.Write([]byte(msg.(string)))

		// if _, ok := msg.(string); ok {
		// 	msg = map[string]interface{}{"message": msg}
		// }

		// // Send response
		// if !c.Response().Committed {
		// 	if c.Request().Method == echo.HEAD { // Issue #608
		// 		err = c.NoContent(code)
		// 	} else {
		// 		err = c.JSON(code, msg)
		// 	}
		// 	if err != nil {
		// 		e.Logger.Error(err)
		// 	}
		// }
	}
}
