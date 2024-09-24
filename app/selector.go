// selector - Handle adding / removing to selection and hovered object

package app

import (
	"fmt"

	"github.com/bonoboris/satisfied/colors"
	"github.com/bonoboris/satisfied/log"
	rl "github.com/gen2brain/raylib-go/raylib"
)

var selector Selector

type Selector struct {
	// True when drawing the selector rectangle
	selecting bool
	// Selector rectangle corners
	start, end rl.Vector2
	// objects in selector rectangle
	ObjectSelection
}

func (s Selector) traceState(key, val string) {
	if log.WillTrace() {
		if key != "" && val != "" {
			log.Trace("selector", key, val, "selecting", s.selecting, "start", s.start, "end", s.end)
			log.Trace("selector", "buildingIdxs", s.BuildingIdxs)
			log.Trace("selector", "pathIdxs", s.PathIdxs)
			log.Trace("selector", "textboxIdxs", s.TextBoxIdxs)
		} else {
			log.Trace("selector", "selecting", s.selecting, "start", s.start, "end", s.end)
			log.Trace("selector", "buildingIdxs", s.BuildingIdxs)
			log.Trace("selector", "pathIdxs", s.PathIdxs)
			log.Trace("selector", "textboxIdxs", s.TextBoxIdxs)
		}
	}
}

func (s *Selector) Reset() {
	s.traceState("before", "Reset")
	log.Debug("selector.reset")
	s.selecting = false
	s.start = rl.Vector2{}
	s.end = rl.Vector2{}
	s.ObjectSelection.reset()
	s.traceState("after", "Reset")
}

// GetAction processes inputs in [ModeNormal] and returns an action to be performed.
//
// See: [GetActionFunc]
func (s *Selector) GetAction() Action {
	app.Mode.Assert(ModeNormal)

	if mouse.Left.Pressed && mouse.InScene {
		if scene.Hovered.IsEmpty() {
			return s.doInit(mouse.Pos)
		} else {
			return selection.doInitSingleDrag(scene.Hovered, mouse.Pos)
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
	s.traceState("before", "doInit")
	log.Debug("selector.doInit", "pos", pos)
	s.ObjectSelection.reset()
	s.selecting = true
	s.start = pos
	s.end = pos
	s.traceState("after", "doInit")
	return app.doSwitchMode(ModeNormal, ResetAll().WithSelector(false))
}

func (s *Selector) doMoveTo(pos rl.Vector2) Action {
	s.traceState("before", "doMoveTo")
	log.Trace("selector.doMoveTo", "pos", pos) // moving by mouse -> tracing
	app.Mode.Assert(ModeNormal)
	assert(s.selecting, "Selector.doMoveTo: selector not active")
	s.end = pos
	rect := rl.NewRectangleCorners(s.start, s.end)
	s.ObjectSelection.reset()
	scene.SelectFromRect(&s.ObjectSelection, rect)
	s.traceState("after", "doMoveTo")
	return nil
}

func (s *Selector) doSelect() Action {
	s.traceState("before", "doSelect")
	log.Debug("selector.doSelect")
	app.Mode.Assert(ModeNormal)
	assert(s.selecting, "Selector.doMoveTo: selector not active")
	s.traceState("after", "doSelect")
	return selection.doInitSelection(s.ObjectSelection)
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

		// TODO: draw objects in selector rectangle
	}
}
