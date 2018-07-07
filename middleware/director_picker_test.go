package middleware

import (
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/pike"
)

func TestDirectorPicker(t *testing.T) {
	directors := make(pike.Directors, 0)
	aslant := "aslant"
	d := &pike.Director{
		Name: aslant,
		Hosts: []string{
			"(www.)?aslant.site",
		},
	}
	directors = append(directors, d)

	tiny := "tiny"
	d = &pike.Director{
		Name: tiny,
		Prefixs: []string{
			"/api",
		},
	}
	directors = append(directors, d)
	for _, d := range directors {
		d.RefreshPriority()
	}
	sort.Sort(directors)
	config := DirectorPickerConfig{}
	t.Run("get director match host", func(t *testing.T) {
		fn := DirectorPicker(config, directors)
		r := httptest.NewRequest(http.MethodGet, "http://aslant.site/api/users/me", nil)
		c := pike.NewContext(r)
		err := fn(c, func() error {
			return nil
		})
		if err != nil {
			t.Fatalf("director picker middleware fail, %v", err)
		}
		if c.Director == nil || c.Director.Name != aslant {
			t.Fatalf("director picker fail")
		}
	})

	t.Run("get director match url prefix", func(t *testing.T) {
		fn := DirectorPicker(config, directors)
		r := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
		c := pike.NewContext(r)
		err := fn(c, func() error {
			return nil
		})
		if err != nil {
			t.Fatalf("director picker middleware fail, %v", err)
		}
		if c.Director == nil || c.Director.Name != tiny {
			t.Fatalf("director picker fail")
		}
	})

	t.Run("no director match", func(t *testing.T) {
		fn := DirectorPicker(config, directors)
		r := httptest.NewRequest(http.MethodGet, "/test", nil)
		c := pike.NewContext(r)
		c.Request = r
		err := fn(c, func() error {
			return nil
		})
		if err != ErrDirectorNotFound {
			t.Fatalf("no director match should return error")
		}
	})

	t.Run("cache response pass director picker", func(t *testing.T) {
		fn := DirectorPicker(config, directors)
		r := httptest.NewRequest(http.MethodGet, "/test", nil)
		c := pike.NewContext(r)
		c.Request = r
		c.Status = cache.Cacheable
		err := fn(c, func() error {
			return nil
		})
		if err != nil {
			t.Fatalf("pass diretcor picker fail, %v", err)
		}
		if c.Director != nil {
			t.Fatalf("cache response should pass director picker")
		}
	})

}
