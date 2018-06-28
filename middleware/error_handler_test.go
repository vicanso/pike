package custommiddleware

import (
	"net/http/httptest"
	"testing"

	"github.com/vicanso/pike/vars"

	"github.com/labstack/echo"
	"github.com/vicanso/pike/cache"
)

func TestErrorHandler(t *testing.T) {
	client := &cache.Client{
		Path: "/tmp/test.cache",
	}
	err := client.Init()
	if err != nil {
		t.Fatalf("cache init fail, %v", err)
	}
	defer client.Close()
	e := echo.New()
	fn := CreateErrorHandler(e, client)
	req := httptest.NewRequest(echo.POST, "/users/me", nil)
	resp := &httptest.ResponseRecorder{}
	c := e.NewContext(req, resp)
	pc := NewContext(c)
	pc.identity = []byte("ABCD")
	pc.status = cache.Fetching
	fn(vars.ErrDirectorNotFound, pc)
}
