package app

import (
	"fmt"

	"github.com/bonoboris/satisfied/log"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// NewPath represents a new path creation state (corresponding to [ModeNewPath])
var newPath NewPath

// NewPath represents a new path creation state (corresponding to [ModeNewPath])
type NewPath struct {
	path           Path
	firstEndPlaced bool
	reverse        bool
	isValid        bool
}

func (np NewPath) traceState(key, val string) {
	if key != "" && val != "" {
		log.Trace("newPath", key, val, "path", np.path, "firstEndPlaced", np.firstEndPlaced, "reverse", np.reverse, "isValid", np.isValid)
	} else {
		log.Trace("newPath", "path", np.path, "firstEndPlaced", np.firstEndPlaced, "reverse", np.reverse, "isValid", np.isValid)
	}
}

// Reset resets the [NewPath] state
func (np *NewPath) Reset() {
	np.traceState("before", "Reset")
	log.Debug("newPath.reset")
	np.path = Path{DefIdx: -1}
	np.firstEndPlaced = false
	np.isValid = false
	np.reverse = false
	np.traceState("after", "Reset")
}

// GetAction processes inputs in [ModeNewPath], and returns an action to be performed
//
// See: [GetActionFunc]
func (np *NewPath) GetAction() Action {
	app.Mode.Assert(ModeNewPath)
	// TODO: Implement arrow keys nudging ?

	switch keyboard.Pressed {
	case rl.KeyR:
		return np.doReverse()
	case rl.KeyEscape:
		// escape cancels first end placement then switches to normal mode
		if np.firstEndPlaced {
			return np.doInit(np.path.DefIdx)
		} else {
			return app.doSwitchMode(ModeNormal, ResetAll())
		}
	}

	if !mouse.InScene {
		return nil
	}

	if mouse.Left.Released {
		if np.firstEndPlaced {
			return np.doPlace()
		}
		return np.doPlaceStart()
	}
	if !mouse.Left.Down {
		return np.doMoveTo(mouse.SnappedPos)
	}
	return nil
}

func (np *NewPath) doInit(defIdx int) Action {
	np.traceState("before", "doInit")
	log.Debug("newPath.doInit", "defIdx", defIdx)
	np.path = Path{DefIdx: defIdx}
	np.firstEndPlaced = false
	np.isValid = true
	resets := ResetAll().WithNewPath(false).WithGui(false)
	np.traceState("after", "doInit")
	return app.doSwitchMode(ModeNewPath, resets)
}

func (np *NewPath) doReverse() Action {
	np.traceState("before", "doReverse")
	log.Debug("newPath.doReverse")
	app.Mode.Assert(ModeNewPath)
	np.reverse = !np.reverse
	np.path.Start, np.path.End = np.path.End, np.path.Start
	np.traceState("after", "doReverse")
	return nil
}

func (np *NewPath) doMoveTo(pos rl.Vector2) Action {
	np.traceState("before", "doMoveTo")
	log.Trace("newPath.doMoveTo", "pos", pos) // moving by mouse -> tracing
	app.Mode.Assert(ModeNewPath)
	if !np.firstEndPlaced {
		np.path.Start = pos
		np.path.End = pos
	} else {
		if np.reverse {
			np.path.Start = pos
		} else {
			np.path.End = pos
		}
		np.isValid = scene.IsPathValid(np.path)
	}
	np.traceState("after", "doMoveTo")
	return nil
}

func (np *NewPath) doPlaceStart() Action {
	np.traceState("before", "doPlaceStart")
	log.Debug("newPath.doPlaceStart")
	app.Mode.Assert(ModeNewPath)
	np.firstEndPlaced = true
	np.traceState("after", "doPlaceStart")
	return nil
}

func (np *NewPath) doPlace() Action {
	np.traceState("before", "doPlace")
	log.Debug("newPath.doPlace")
	app.Mode.Assert(ModeNewPath)
	assert(np.firstEndPlaced, "path start not placed")
	if np.isValid = scene.IsPathValid(np.path); np.isValid {
		scene.AddPath(np.path)
	}
	np.path.Start = mouse.SnappedPos
	np.path.End = mouse.SnappedPos
	np.firstEndPlaced = false
	np.isValid = true
	np.traceState("after", "doPlace")
	return nil
}

// Dispatch performs an [NewPath] action, updating its state, and returns an new action to be performed
//
// See: [ActionHandler]
func (np *NewPath) Dispatch(action Action) Action {
	switch action := action.(type) {
	case NewPathActionInit:
		return np.doInit(action.DefIdx)
	case NewPathActionMoveTo:
		return np.doMoveTo(action.Pos)
	case NewPathActionReverse:
		return np.doReverse()
	case NewPathActionPlaceStart:
		return np.doPlaceStart()
	case NewPathActionPlace:
		return np.doPlace()

	default:
		panic(fmt.Sprintf("NewPath.Dispatch: cannot handle: %T", action))
	}
}

func (np NewPath) Draw() {
	if np.isValid {
		np.path.Draw(DrawNew)
	} else {
		np.path.Draw(DrawInvalid)
	}
}
