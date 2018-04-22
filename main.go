package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/vicanso/pike/middleware"
)

func main() {
	// Echo instance
	e := echo.New()

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	})

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(customMiddleware.UpstreamPicker())

	e.Use(customMiddleware.Identifier())

	// Routes

	// Start server
	e.Logger.Fatal(e.Start(":3015"))

}
