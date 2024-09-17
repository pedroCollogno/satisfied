package app

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// NewBuilding represents a new building creation state (corresponding to [ModeNewBuilding])
var newBuilding NewBuilding

// NewBuilding represents a new building creation state (corresponding to [ModeNewBuilding])
type NewBuilding struct {
	building Building
	isValid  bool
}

// Reset resets the [NewBuilding] state
func (nb *NewBuilding) Reset() {
	nb.building = Building{DefIdx: -1}
	nb.isValid = false
}

// GetAction processes inputs in [ModeNewBuilding], and returns an action to be performed.
//
// See: [GetActionFunc]
func (nb *NewBuilding) GetAction() (action Action) {
	appMode.Assert(ModeNewBuilding)

	if keyboard.Pressed == rl.KeyR {
		return nb.doRotate()
	}
	if mouse.Left.Released {
		return nb.doPlace()
	}
	if mouse.InScene && !mouse.Left.Down {
		return nb.doMoveTo(mouse.SnappedPos)
	}
	return nil
}

func (nb *NewBuilding) doInit(defIdx int) Action {
	nb.building = Building{DefIdx: defIdx}
	nb.isValid = true
	resets := ResetAll().WithNewBuilding(false).WithGui(false)
	return appMode.doSwitchMode(ModeNewBuilding, resets)
}

func (nb *NewBuilding) doMoveTo(pos rl.Vector2) Action {
	appMode.Assert(ModeNewBuilding)
	nb.building.Pos = pos
	nb.isValid = scene.IsBuildingValid(nb.building, -1)
	return nil
}

func (nb *NewBuilding) doRotate() Action {
	appMode.Assert(ModeNewBuilding)
	nb.isValid = scene.IsBuildingValid(nb.building, -1)
	nb.building.Rot += 90
	return nil
}

func (nb *NewBuilding) doPlace() Action {
	rl.TraceLog(rl.LogWarning, fmt.Sprintf("NewBuilding.doPlace: building=%v", nb.building))
	appMode.Assert(ModeNewBuilding)
	nb.isValid = scene.IsBuildingValid(nb.building, -1)
	if nb.isValid {
		scene.AddBuilding(nb.building)
	}
	nb.building.Pos = mouse.SnappedPos
	nb.isValid = true
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
