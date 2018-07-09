package pike

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vicanso/pike/cache"
)

func TestContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://aslant.site:5000/users/login?cache-control=no-cache", nil)
	c := NewContext(req)
	c.Status = 1
	c.Identity = []byte("ABCD")
	c.Director = &Director{}
	c.Resp = &cache.Response{}
	c.Fresh = true
	t.Run("reset", func(t *testing.T) {
		time.Sleep(10 * time.Millisecond)
		c.Reset()
		if c.Status != 0 {
			t.Fatalf("reset status fail")
		}
		if c.Identity != nil {
			t.Fatalf("reset identity fail")
		}
		if c.Director != nil {
			t.Fatalf("reset director fail")
		}
		if c.Resp != nil {
			t.Fatalf("reset resp fail")
		}
		if c.Fresh {
			t.Fatalf("reset fresh fail")
		}
		if time.Now().UnixNano()-c.CreatedAt.UnixNano() > int64(time.Millisecond) {
			t.Fatalf("reset created at fail")
		}
	})
	t.Run("get real ip", func(t *testing.T) {
		req.Header.Set(HeaderXRealIP, "1.1.1.1")
		if c.RealIP() != "1.1.1.1" {
			t.Fatalf("get the real ip from x-real-ip fail")
		}

		req.Header.Set(HeaderXForwardedFor, "1.1.1.2, 192.168.1.1")
		if c.RealIP() != "1.1.1.2" {
			t.Fatalf("get the real ip from x-forwarded-for fail")
		}
	})

	t.Run("json", func(t *testing.T) {
		data := make(map[string]string)
		data["a"] = "1"
		err := c.JSON(data, http.StatusOK)
		if err != nil {
			t.Fatalf("json fail, %v", err)
		}
		if string(c.Response.Bytes()) != `{"a":"1"}` {
			t.Fatalf("json response fail")
		}
	})

	t.Run("error", func(t *testing.T) {

		r := &http.Request{}
		w := httptest.NewRecorder()
		c := NewContext(r)
		c.ResponseWriter = w
		c.Error(errors.New("ABCD"))
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status code should be internal error")
		}
		if string(w.Body.Bytes()) != "ABCD" {
			t.Fatalf("response body should be ABCD")
		}
	})
}
