// mouse - mouse input state

package app

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

var mouse = Mouse{}

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

// Update mouse state
//
// Depends on [Dims] and [Camera]
func (m *Mouse) Update() {
	m.Wheel = rl.GetMouseWheelMove()

	newScreenPos := rl.GetMousePosition()
	newInScene := dims.Scene.CheckCollisionPoint(newScreenPos)

	// Disable cursor in scene area
	if newInScene && !m.InScene {
		rl.SetMouseCursor(rl.MouseCursorCrosshair)
		// rl.DisableCursor()
		// toggling cursor seem to reset its position
		rl.SetMousePosition(int(newScreenPos.X), int(newScreenPos.Y))
	} else if !newInScene && m.InScene {
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
	// True on the first frame the button is pressed
	Pressed bool
	// True on the first frame the button is released
	Released bool
	// True if the button is currently down
	Down bool
	// Mouse position the last time the button was up (in world coordinates)
	LastUpPos rl.Vector2
	// Mouse position the last time the button was up (in screen coordinates)
	LastUpScreenPos rl.Vector2
	// Snapped mouse position when the last time the button was up (in world coordinates)
	LastUpSnappedPos rl.Vector2
}

func (mb *mouseButton) Update(newDown bool) {
	if !newDown && !mb.Down {
		// Update position only on 2 consecutive frames in up state
		// so that we can get LastUp values on Released.
		mb.LastUpPos = mouse.Pos
		mb.LastUpScreenPos = mouse.ScreenPos
		mb.LastUpSnappedPos = mouse.SnappedPos
	}
	mb.Pressed = newDown && !mb.Down
	mb.Released = !newDown && mb.Down
	mb.Down = newDown
}
