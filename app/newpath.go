package app

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// NewPath represents a new path creation state (corresponding to [ModeNewPath])
var newPath NewPath

// NewPath represents a new path creation state (corresponding to [ModeNewPath])
type NewPath struct {
	path        Path
	startPlaced bool
	reverse     bool
	isValid     bool
}

// Reset resets the [NewPath] state
func (np *NewPath) Reset() {
	np.path = Path{DefIdx: -1}
	np.startPlaced = false
	np.isValid = false
	np.reverse = false
}

// GetAction processes inputs in [ModeNewPath], and returns an action to be performed
//
// See: [GetActionFunc]
func (np *NewPath) GetAction() Action {
	appMode.Assert(ModeNewPath)
	// TODO: Implement arrow keys nudging ?

	switch keyboard.Pressed {
	case rl.KeyR:
		return np.doReverse()
	}

	if !mouse.InScene {
		return nil
	}

	if mouse.Left.Released {
		if np.startPlaced {
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
	np.path = Path{DefIdx: defIdx}
	np.startPlaced = false
	np.isValid = true
	resets := ResetAll().WithNewPath(false).WithGui(false)
	return appMode.doSwitchMode(ModeNewPath, resets)
}

func (np *NewPath) doReverse() Action {
	appMode.Assert(ModeNewPath)
	np.reverse = !np.reverse
	np.path.Start, np.path.End = np.path.End, np.path.Start
	return nil
}

func (np *NewPath) doMoveTo(pos rl.Vector2) Action {
	appMode.Assert(ModeNewPath)
	if !np.startPlaced {
		np.path.Start = pos
		np.path.End = pos
	} else {
		if np.reverse {
			np.path.End = pos
		} else {
			np.path.Start = pos
		}
		np.isValid = scene.IsPathValid(np.path)
	}
	return nil
}

func (np *NewPath) doPlaceStart() Action {
	appMode.Assert(ModeNewPath)
	np.startPlaced = true
	return nil
}

func (np *NewPath) doPlace() Action {
	appMode.Assert(ModeNewPath)
	assert(np.startPlaced, "path start not placed")
	if np.isValid = scene.IsPathValid(np.path); np.isValid {
		scene.AddPath(np.path)
	}
	np.path.Start = mouse.SnappedPos
	np.path.End = mouse.SnappedPos
	np.startPlaced = false
	np.isValid = true
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
