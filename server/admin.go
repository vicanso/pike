package server

import (
	"github.com/vicanso/cod"
	recover "github.com/vicanso/cod-recover"
	responder "github.com/vicanso/cod-responder"
	"github.com/vicanso/pike/performance"
)

// NewAdminServer create an admin server
func NewAdminServer(prefix string) *cod.Cod {
	d := cod.New()
	d.Use(recover.New())
	d.Use(responder.NewDefault())

	g := cod.NewGroup(prefix)
	g.GET("/stats", func(c *cod.Context) error {
		c.Body = performance.GetStats()
		// c.BodyBuffer = bytes.NewBufferString("abcd")
		return nil
	})
	d.AddGroup(g)
	return d
}
