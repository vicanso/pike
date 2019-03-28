package main

import (
	"github.com/vicanso/cod"
)

func main() {
	d := cod.New()

	d.ListenAndServe(":7001")
}
