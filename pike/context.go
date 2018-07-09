package pike

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/vicanso/pike/cache"
)

type (
	// Context context
	Context struct {
		Request        *http.Request
		Response       *Response
		ResponseWriter http.ResponseWriter
		ServerTiming   *ServerTiming
		// Status 该请求的状态 fetching pass等
		Status int
		// Identity 该请求的标记
		Identity []byte
		// Director 该请求对应的director
		Director *Director
		// Resp 该请求的响应数据
		Resp *cache.Response
		// Fresh 是否fresh
		Fresh bool
		// CreatedAt 创建时间
		CreatedAt time.Time
	}
)

// NewContext 创新新的Context并重置相应的属性
func NewContext(req *http.Request) (c *Context) {
	c = contextPool.Get().(*Context)
	if c.ServerTiming == nil {
		c.ServerTiming = NewServerTiming()
	} else {
		c.ServerTiming.Reset()
	}
	if c.Response == nil {
		c.Response = NewResponse()
	} else {
		c.Response.Reset()
	}
	c.Request = req
	c.Reset()
	return
}

// Reset 重置context
func (c *Context) Reset() {
	c.Status = 0
	c.Identity = nil
	c.Director = nil
	c.Resp = nil
	c.Fresh = false
	c.CreatedAt = time.Now()
}

// RealIP 客户端真实IP
func (c *Context) RealIP() string {
	ra := c.Request.RemoteAddr
	if ip := c.Request.Header.Get(HeaderXForwardedFor); ip != "" {
		ra = strings.Split(ip, ", ")[0]
	} else if ip := c.Request.Header.Get(HeaderXRealIP); ip != "" {
		ra = ip
	} else {
		ra, _, _ = net.SplitHostPort(ra)
	}
	return ra
}

// JSON 返回json
func (c *Context) JSON(data interface{}, status int) error {
	buf, err := json.Marshal(data)
	if err != nil {
		return err
	}
	resp := c.Response
	header := resp.Header()
	header.Set(HeaderContentType, JSONContent)
	header.Set(HeaderContentLength, strconv.Itoa(len(buf)))
	resp.WriteHeader(status)
	_, err = resp.Write(buf)
	return err
}

// Error 出错处理
func (c *Context) Error(err error) {
	resp := c.ResponseWriter
	resp.WriteHeader(http.StatusInternalServerError)
	resp.Write([]byte(err.Error()))
}
