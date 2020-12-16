package main

import (
	"github.com/maxence-charriere/go-app/v7/pkg/app"
)

func main() {
	app.RouteWithRegexp("/server/[0-9]+", &web{})
	app.Run()
}
