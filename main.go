package main

import (
	"embed"

	"github.com/bonoboris/satisfied/app"
)

// Data are embedded in the executable

//go:embed assets
var assets embed.FS

func main() {
	app.Init(assets)
	defer app.Close()

	app.LoadSomeScene()

	for !app.ShouldExit() {
		app.Step()
	}
}
