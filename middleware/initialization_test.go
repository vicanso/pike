package custommiddleware

import (
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
)

func TestInitialization(t *testing.T) {
	conf := InitializationConfig{
		Header: []string{
			"X-Token:ABCD",
		},
		Concurrency: 10,
	}
	fn := Initialization(conf)(func(c echo.Context) error {
		pc := c.(*Context)
		if pc.Response().Header().Get("X-Token") != "ABCD" {
			t.Fatalf("set header in init function fail")
		}
		return nil
	})
	resp := &httptest.ResponseRecorder{}
	e := echo.New()
	c := e.NewContext(nil, resp)
	pc := NewContext(c)
	err := fn(pc)
	if err != nil {
		t.Fatalf("initialization fail")
	}
}
