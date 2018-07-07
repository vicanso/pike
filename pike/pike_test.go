package pike

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPike(t *testing.T) {
	t.Run("use middleware", func(t *testing.T) {
		p := New()

		stepList := []string{}
		p.Use(func(c *Context, next Next) error {
			stepList = append(stepList, "START1")
			err := next()
			if err != nil {
				return err
			}
			stepList = append(stepList, "END1")
			return nil
		})

		p.Use(func(c *Context, next Next) error {
			stepList = append(stepList, "START2")
			err := next()
			if err != nil {
				return err
			}
			stepList = append(stepList, "END2")
			return nil
		})

		r := &http.Request{}
		w := httptest.NewRecorder()
		p.ServeHTTP(w, r)
		if strings.Join(stepList, ",") != "START1,START2,END2,END1" {
			t.Fatalf("the midllde run order is wrong")
		}
	})

	t.Run("error handler", func(t *testing.T) {
		p := New()
		catchError := false
		customError := errors.New("throw an error")
		p.ErrorHandler = func(err error, c *Context) {
			if err != customError {
				t.Fatalf("error handler fail")
			}
			catchError = true
		}
		stepList := []string{}

		p.Use(func(c *Context, next Next) error {
			stepList = append(stepList, "START1")
			err := next()
			if err != nil {
				return err
			}
			stepList = append(stepList, "END1")
			return nil
		})

		p.Use(func(c *Context, next Next) error {
			return customError
		})
		r := &http.Request{}
		w := httptest.NewRecorder()
		p.ServeHTTP(w, r)
		if !catchError {
			t.Fatalf("error handler should catch error")
		}
		if len(stepList) != 1 && stepList[0] != "START1" {
			t.Fatalf("midlleware run order is wrong")
		}
	})

}
