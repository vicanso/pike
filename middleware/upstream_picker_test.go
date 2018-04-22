package customMiddleware

import (
	"fmt"
	"testing"

	"github.com/vicanso/pike/config"
)

func TestUpstreamPicker(t *testing.T) {
	directors := make([]*config.Director, 0, 5)
	directors = append(directors, &config.Director{
		Name: "aslant",
		Type: "first",
		Pass: []string{
			"cache-control=no-cache",
		},
		Ping: "/ping",
		Prefix: []string{
			"/api",
		},
		Host: []string{
			"aslant.site",
		},
		Backends: []string{
			"http://127.0.0.1:5018",
			"http://127.0.0.1:5019",
		},
	})
	directors = append(directors, &config.Director{
		Name: "tiny",
		Pass: []string{
			"cache-control=no-cache",
		},
		Ping: "/ping",
		Host: []string{
			"tiny.aslant.site",
		},
		Backends: []string{
			"http://127.0.0.1:6018",
			"http://127.0.0.1:6019",
		},
	})

	fn := UpstreamPicker(directors)
	fmt.Println(fn)

}
