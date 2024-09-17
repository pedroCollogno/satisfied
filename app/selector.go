// selector - Handle adding / removing to selection and hovered object

package app

import (
	"fmt"

	"github.com/bonoboris/satisfied/colors"
	rl "github.com/gen2brain/raylib-go/raylib"
)

var selector Selector

type Selector struct {
	// True when drawing the selector rectangle
	selecting bool
	// Selector rectangle corners
	start, end rl.Vector2
}

func (s *Selector) Reset() {
	s.selecting = false
	s.start = rl.Vector2{}
	s.end = rl.Vector2{}
}

// GetAction processes inputs in [ModeNormal] and returns an action to be performed.
//
// See: [GetActionFunc]
func (s *Selector) GetAction() Action {
	appMode.Assert(ModeNormal)

	if mouse.Left.Pressed && mouse.InScene {
		if scene.Hovered.IsEmpty() {
			return s.doInit(mouse.Pos)
		} else {
			return selection.doInitSingle(scene.Hovered, SelectionDrag, mouse.Pos)
		}
	}
	if s.selecting {
		if mouse.Left.Released {
			return s.doSelect()
		} else if mouse.Left.Down {
			return s.doMoveTo(mouse.Pos)
		}
	}
	return nil
}

func (s *Selector) doInit(pos rl.Vector2) Action {
	s.selecting = true
	s.start = pos
	s.end = pos
	return appMode.doSwitchMode(ModeNormal, ResetAll().WithSelector(false))
}

func (s *Selector) doMoveTo(pos rl.Vector2) Action {
	appMode.Assert(ModeNormal)
	assert(s.selecting, "Selector.doMoveTo: selector not active")
	s.end = pos
	return nil
}

func (s *Selector) doSelect() Action {
	appMode.Assert(ModeNormal)
	assert(s.selecting, "Selector.doMoveTo: selector not active")
	rect := rl.NewRectangleCorners(s.start, s.end)
	return selection.doInitRectangle(rect, SelectionNormal, rl.Vector2{})
}

// Dispatch performs [Selector] action, updating its state, and returns an new action to be performed
//
// See: [ActionHandler]
func (s *Selector) Dispatch(action Action) Action {
	switch action := action.(type) {
	case SelectorActionInit:
		return s.doInit(action.Pos)
	case SelectorActionMoveTo:
		return s.doMoveTo(action.Pos)
	case SelectorActionSelect:
		return s.doSelect()

	default:
		panic(fmt.Sprintf("Selector.Dispatch: cannot handle: %T", action))
	}
}

// Draw selector rectangle
func (s Selector) Draw() {
	if s.selecting {
		rect := rl.NewRectangleCorners(s.start, s.end)
		rl.DrawRectangleLinesEx(rect, 3/camera.Zoom(), colors.WithAlpha(colors.Blue500, 0.5))
	}
}