// mouse - mouse input state

package app

import (
	"github.com/bonoboris/satisfied/log"
	rl "github.com/gen2brain/raylib-go/raylib"
)

var mouse = Mouse{Left: mouseButton{name: "mouse.left"}, Middle: mouseButton{name: "mouse.middle"}, Right: mouseButton{name: "mouse.right"}}

type Mouse struct {
	// Mouse position (in world coordinates)
	Pos rl.Vector2

	// Mouse position (in screen coordinates)
	ScreenPos rl.Vector2

	// Snapped mouse position (in world coordinates)
	SnappedPos rl.Vector2

	// Mouse movement delta since last frame (in screen coordinates)
	ScreenDelta rl.Vector2

	// True when mouse is inside the scene area
	InScene bool

	// Wheel movement
	Wheel float32

	// Left mouse button
	Left mouseButton

	// Middle mouse button
	Middle mouseButton

	// Right mouse button
	Right mouseButton
}

func (m *Mouse) traceState() {
	log.Trace("mouse", "pos", m.Pos, "screenPos", m.ScreenPos, "snapPos", m.SnappedPos, "screenDelta", m.ScreenDelta, "inScene", m.InScene, "wheel", m.Wheel)
	m.Left.traceState()
	m.Middle.traceState()
	m.Right.traceState()
}

// Update mouse state
//
// Depends on [Dims] and [Camera]
func (m *Mouse) Update() {
	m.Wheel = rl.GetMouseWheelMove()
	if m.Wheel != 0 {
		log.Debug("mouse wheel", "wheel", m.Wheel)
		defer mouse.traceState()
	}

	newScreenPos := rl.GetMousePosition()
	newInScene := dims.Scene.CheckCollisionPoint(newScreenPos)

	// Disable cursor in scene area
	if newInScene && !m.InScene {
		log.Debug("mouse enter scene")
		defer mouse.traceState()
		rl.SetMouseCursor(rl.MouseCursorCrosshair)
		// rl.DisableCursor()
		// toggling cursor seem to reset its position
		rl.SetMousePosition(int(newScreenPos.X), int(newScreenPos.Y))
	} else if !newInScene && m.InScene {
		log.Debug("mouse leave scene")
		defer mouse.traceState()
		rl.SetMouseCursor(rl.MouseCursorDefault)
		// rl.EnableCursor()
		// toggling cursor seem to reset its position
		rl.SetMousePosition(int(newScreenPos.X), int(newScreenPos.Y))
	}

	m.ScreenDelta = newScreenPos.Subtract(m.ScreenPos)
	m.ScreenPos = newScreenPos
	m.Pos = camera.WorldPos(newScreenPos)
	m.SnappedPos = grid.Snap(m.Pos)
	m.InScene = newInScene
	m.Left.Update(rl.IsMouseButtonDown(rl.MouseLeftButton))
	m.Middle.Update(rl.IsMouseButtonDown(rl.MouseMiddleButton))
	m.Right.Update(rl.IsMouseButtonDown(rl.MouseRightButton))
}

// button represents a mouse button state
type mouseButton struct {
	// button name (for logging)
	name string
	// True on the first frame the button is pressed
	Pressed bool
	// True on the first frame the button is released
	Released bool
	// True if the button is currently down
	Down bool
}

func (mb *mouseButton) traceState() {
	log.Trace(mb.name, "pressed", mb.Pressed, "released", mb.Released, "down", mb.Down)
}

func (mb *mouseButton) Update(newDown bool) {
	mb.Pressed = newDown && !mb.Down
	mb.Released = !newDown && mb.Down
	mb.Down = newDown
	if mb.Pressed {
		log.Debug("mouse button pressed", "name", mb.name)
		mouse.traceState()
	}
	if mb.Released {
		log.Debug("mouse button released", "name", mb.name)
		mouse.traceState()
	}
}
