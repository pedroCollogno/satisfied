package app

import rl "github.com/gen2brain/raylib-go/raylib"

var dims Dims

type Dims struct {
	// Screen size
	Screen rl.Vector2
	// Scene area dimensions
	Scene rl.Rectangle
}

// Update screen and scene dimensions state
//
// Do not depends on any other state
func (d *Dims) Update() {
	width := rl.GetScreenWidth()
	height := rl.GetScreenHeight()
	d.Screen = vec2(float32(width), float32(height))
	d.Scene = rl.NewRectangle(
		SidebarWidth,
		TopbarHeight,
		d.Screen.X-SidebarWidth,
		d.Screen.Y-TopbarHeight-StatusBarHeight,
	)
}
