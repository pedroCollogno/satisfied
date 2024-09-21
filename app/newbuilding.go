package app

import (
	"fmt"

	"github.com/bonoboris/satisfied/log"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// NewBuilding represents a new building creation state (corresponding to [ModeNewBuilding])
var newBuilding NewBuilding

// NewBuilding represents a new building creation state (corresponding to [ModeNewBuilding])
type NewBuilding struct {
	building Building
	isValid  bool
}

func (nb NewBuilding) traceState(key, val string) {
	if key != "" && val != "" {
		log.Trace("newBuilding", key, val, "building", nb.building, "isValid", nb.isValid)
	} else {
		log.Trace("newBuilding", "building", nb.building, "isValid", nb.isValid)
	}
}

// Reset resets the [NewBuilding] state
func (nb *NewBuilding) Reset() {
	nb.traceState("before", "Reset")
	log.Debug("newBuilding.reset")
	nb.building = Building{DefIdx: -1}
	nb.isValid = false
	nb.traceState("after", "Reset")
}

// GetAction processes inputs in [ModeNewBuilding], and returns an action to be performed.
//
// See: [GetActionFunc]
func (nb *NewBuilding) GetAction() (action Action) {
	app.Mode.Assert(ModeNewBuilding)

	switch keyboard.Pressed {
	case rl.KeyEscape:
		return app.doSwitchMode(ModeNormal, ResetAll())
	case rl.KeyR:
		return nb.doRotate()
	}

	if !mouse.InScene {
		return nil
	}
	if mouse.Left.Released {
		return nb.doPlace()
	}
	if !mouse.Left.Down {
		return nb.doMoveTo(mouse.SnappedPos)
	}
	return nil
}

func (nb *NewBuilding) doInit(defIdx int) Action {
	nb.traceState("before", "doInit")
	log.Debug("newBuilding.doInit", "defIdx", defIdx)
	nb.building = Building{DefIdx: defIdx}
	nb.isValid = true
	resets := ResetAll().WithNewBuilding(false).WithGui(false)
	nb.traceState("after", "doInit")
	return app.doSwitchMode(ModeNewBuilding, resets)
}

func (nb *NewBuilding) doMoveTo(pos rl.Vector2) Action {
	nb.traceState("before", "doMoveTo")
	log.Trace("newBuilding.doMoveTo", "pos", pos) // moving by mouse -> tracing
	app.Mode.Assert(ModeNewBuilding)
	nb.building.Pos = pos
	nb.isValid = scene.IsBuildingValid(nb.building, -1)
	nb.traceState("after", "doMoveTo")
	return nil
}

func (nb *NewBuilding) doRotate() Action {
	nb.traceState("before", "doRotate")
	log.Debug("newBuilding.doRotate")
	app.Mode.Assert(ModeNewBuilding)
	nb.isValid = scene.IsBuildingValid(nb.building, -1)
	nb.building.Rot += 90
	nb.traceState("after", "doRotate")
	return nil
}

func (nb *NewBuilding) doPlace() Action {
	nb.traceState("before", "doPlace")
	log.Debug("newBuilding.doPlace")
	app.Mode.Assert(ModeNewBuilding)
	nb.isValid = scene.IsBuildingValid(nb.building, -1)
	if nb.isValid {
		scene.AddBuilding(nb.building)
	}
	nb.building.Pos = mouse.SnappedPos
	nb.isValid = true
	nb.traceState("after", "doPlace")
	return nil
}

// Dispatch performs an [NewBuilding] action, updating its state, and returns an new action to be performed
//
// See: [ActionHandler]
func (np *NewBuilding) Dispatch(action Action) Action {
	switch action := action.(type) {
	case NewBuildingActionInit:
		return np.doInit(action.DefIdx)
	case NewBuildingActionMoveTo:
		return np.doMoveTo(action.Pos)
	case NewBuildingActionRotate:
		return np.doRotate()
	case NewBuildingActionPlace:
		return np.doPlace()

	default:
		panic(fmt.Sprintf("NewBuilding.Dispatch: cannot handle: %T", action))
	}
}

func (np NewBuilding) Draw() {
	if np.isValid {
		np.building.Draw(DrawNew)
	} else {
		np.building.Draw(DrawInvalid)
	}
}
