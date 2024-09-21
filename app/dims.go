package app

import (
	"github.com/bonoboris/satisfied/log"
	rl "github.com/gen2brain/raylib-go/raylib"
)

var dims Dims

type Dims struct {
	pScreen rl.Vector2
	// Screen size
	Screen rl.Vector2
	// Scene area dimensions
	Scene rl.Rectangle
	// Scene area dimensions (world coordinates)
	World rl.Rectangle
	// Scene area dimensions (world coordinates) expanded by 1 world unit on each side
	ExWorld rl.Rectangle
}

func (d Dims) traceState() {
	log.Trace("dims.screen", "W", d.Screen.X, "H", d.Screen.Y)
	log.Trace("dims.scene", "X", d.Scene.X, "Y", d.Scene.Y, "W", d.Scene.Width, "H", d.Scene.Height)
	log.Trace("dims.world", "X", d.World.X, "Y", d.World.Y, "W", d.World.Width, "H", d.World.Height)
}

// Update screen and scene dimensions state
//
// Depends on [Camera]
func (d *Dims) Update() {
	d.pScreen = d.Screen
	width := rl.GetScreenWidth()
	height := rl.GetScreenHeight()
	d.Screen = vec2(float32(width), float32(height))
	d.Scene = rl.NewRectangle(
		SidebarWidth,
		TopbarHeight,
		d.Screen.X-SidebarWidth,
		d.Screen.Y-TopbarHeight-StatusBarHeight,
	)
	d.World = rl.NewRectangleV(camera.WorldPos(d.Scene.TopLeft()), d.Scene.Size().Scale(1/camera.Zoom()))
	d.ExWorld = rl.NewRectangleV(d.World.TopLeft().SubtractValue(1), d.World.Size().AddValue(2))
	if d.Screen != d.pScreen {
		log.Debug("windows resized", "W", d.Screen.X, "H", d.Screen.Y)
		d.traceState()
	}
}
