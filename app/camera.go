// camera - camera state

package app

import (
	"fmt"

	"github.com/bonoboris/satisfied/math32"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	// Zoom default level (px/wu) wu=world unit
	zoomDefault = 10.
	// Factor to zoom by
	zoomFactor = 1.5
	// Min zoom level = zoomDefault * zoomFactor^-5 (~ x0.13 default)
	zoomMin = zoomDefault * 32. / 243.
	// Max zoom level = zoomDefault * zoomFactor^5 (~ x7.5 default)
	zoomMax = zoomDefault * 243. / 32.
	// Ammount to move the camera by on arrow key press
	moveDelta = 100.
	// Ammount to zoom by on middle mouse button drag
	zoomPerPx = 1.0 / 100.
)

var camera = Camera{}

// Holds camera state
type Camera struct {
	// inner state
	camera rl.Camera2D
}

// Zoom returns the current zoom level
func (c *Camera) Zoom() float32 { return c.camera.Zoom }

// WorldPos converts screen position to world position
func (c *Camera) WorldPos(screenPos rl.Vector2) rl.Vector2 {
	return rl.GetScreenToWorld2D(screenPos, c.camera)
}

// ScreenPos converts world position to screen position
func (c *Camera) ScreenPos(worldPos rl.Vector2) rl.Vector2 {
	return rl.GetWorldToScreen2D(worldPos, c.camera)
}

// BeginMode2D enters raylib 2D mode
func (c *Camera) BeginMode2D() { rl.BeginMode2D(c.camera) }

// EndMode2D exits raylib 2D mode
func (c *Camera) EndMode2D() { rl.EndMode2D() }

// Update processes inputs, and directly execute the [TargetCamera] action to be performed.
//
// [TargetCamera] actions does not have follow up actions.
func (c *Camera) Update() {
	// arrow keys
	if !appMode.CameraArrowsKeyEnabled() {
		switch keyboard.Pressed {
		case rl.KeyRight:
			c.doPan(vec2(-moveDelta, 0))
		case rl.KeyLeft:
			c.doPan(vec2(+moveDelta, 0))
		case rl.KeyDown:
			c.doPan(vec2(0, -moveDelta))
		case rl.KeyUp:
			c.doPan(vec2(0, +moveDelta))
		}
	}

	// +, -, =
	switch keyboard.Pressed {
	case rl.KeyKpSubtract:
		camera.doZoom(-1, dims.Scene.Center())
	case rl.KeyKpAdd:
		camera.doZoom(+1, dims.Scene.Center())
	case rl.KeyMinus:
		if !keyboard.Shift {
			camera.doZoom(-1, dims.Scene.Center())
		}
	case rl.KeyEqual:
		if !keyboard.Shift {
			c.doReset()
		} else {
			camera.doZoom(+1, dims.Scene.Center())
		}
	}

	// mouse inputs
	if mouse.InScene {
		if mouse.Right.Down {
			// paning
			c.doPan(mouse.ScreenDelta)
		} else if mouse.Middle.Down {
			// zooming by middle mouse button drag
			if math32.Abs(mouse.ScreenDelta.X) > math32.Abs(mouse.ScreenDelta.Y) {
				camera.doZoom(mouse.ScreenDelta.X*zoomPerPx, mouse.Middle.LastUpScreenPos)
			} else {
				camera.doZoom(mouse.ScreenDelta.Y*zoomPerPx, mouse.Middle.LastUpScreenPos)
			}
		} else if mouse.Wheel != 0 {
			// zooming by mouse wheel
			camera.doZoom(mouse.Wheel, mouse.ScreenPos)
		}
	}
}

// doReset resets camera state (default zoom, target (0,0) and offset middle of the scene)
func (c *Camera) doReset() Action {
	c.camera.Zoom = zoomDefault
	c.camera.Target = vec2(0, 0)
	c.camera.Offset = dims.Scene.Center()
	return nil
}

// doZoom zooms the camera by a given amount at a given position
func (c *Camera) doZoom(by float32, at rl.Vector2) Action {
	// Set target at world position
	c.camera.Target = c.WorldPos(at)
	// Set offset at screen position
	c.camera.Offset = at
	// Change zoom
	newZoom := c.camera.Zoom * math32.Pow(zoomFactor, by)
	c.camera.Zoom = min(max(newZoom, zoomMin), zoomMax)
	return nil
}

// doPan pans the camera by a given amount
func (c *Camera) doPan(by rl.Vector2) Action {
	// Set target at world position
	c.camera.Offset = c.camera.Offset.Add(by)
	return nil
}

// Dispatch performs a [Camera] action, updating its state, and returns an new action to be performed
//
// Note: All camera actions returns nil (no follow up)
//
// See: [ActionHandler]
func (c *Camera) Dispatch(action Action) Action {
	switch action := action.(type) {
	case CameraActionReset:
		return c.doReset()
	case CameraActionZoom:
		return c.doZoom(action.By, action.At)
	case CameraActionPan:
		return c.doPan(action.By)

	default:
		panic(fmt.Sprintf("Camera.Dispatch: cannot handle: %T", action))
	}
}
