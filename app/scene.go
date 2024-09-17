package app

import (
	"fmt"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Scene holds the scene objects (buildings and paths)
var scene Scene

// Scene holds the scene objects (buildings and paths)
type Scene struct {
	// History of scene operations (undo / redo)
	history []sceneOp
	// Current history position:
	//   - history[:historyPos] all have been done
	//   - history[historyPos:] all have been undone (if existing)
	historyPos int

	// The scene object currently hovered by the mouse
	Hovered Object
	// Placed buildings
	Buildings []Building
	// Placed paths
	Paths []Path
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// Object
////////////////////////////////////////////////////////////////////////////////////////////////////

// Object represents a building or path in the scene
type Object struct {
	// Type of the object
	Type ObjectType
	// Index in either [Scene.Buildings] or [Scene.Paths]
	Idx int
}

// IsEmpty returns true if [Object] is [TypeInvalid]
func (o Object) IsEmpty() bool { return o.Type == TypeInvalid }

// Draw draws the object in the given state
//
// Noop if [Object.Type] is [TypeInvalid]
func (o Object) Draw(state DrawState) {
	switch o.Type {
	case TypeBuilding:
		scene.Buildings[o.Idx].Draw(state)
	case TypePath:
		scene.Paths[o.Idx].Draw(state)
	}
}

// ObjectType enumerates the different types of objects
type ObjectType int

const (
	TypeInvalid ObjectType = iota
	TypeBuilding
	TypePath
)

func (ot ObjectType) String() string {
	switch ot {
	case TypeBuilding:
		return "TypeBuilding"
	case TypePath:
		return "TypePath"
	default:
		return "TypeInvalid"
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// sceneOp (undo / redo)
////////////////////////////////////////////////////////////////////////////////////////////////////

type sceneOp interface {
	// Do the operation
	do(*Scene)
	// Redo undone operation
	redo(*Scene) Action
	// Undo done operation
	undo(*Scene) Action
}

type sceneOpAdd struct {
	paths     []Path
	buildings []Building
}

func (op *sceneOpAdd) do(s *Scene) {
	s.Paths = append(s.Paths, op.paths...)
	s.Buildings = append(s.Buildings, op.buildings...)
}

func (op *sceneOpAdd) redo(s *Scene) Action {
	op.do(s)
	ss := sceneSubset{
		buildingsIdxs: Range(len(s.Buildings)-len(op.buildings), len(s.Buildings)),
		pathsIdxs:     Range(len(s.Paths)-len(op.paths), len(s.Paths)),
	}
	ss.updateBounds()
	return selection.doInitSceneSubset(ss)
}

func (op *sceneOpAdd) undo(s *Scene) Action {
	s.Paths = s.Paths[:len(s.Paths)-len(op.paths)]
	s.Buildings = s.Buildings[:len(s.Buildings)-len(op.buildings)]
	return appMode.doSwitchMode(ModeNormal, ResetAll())
}

type sceneOpDelete struct {
	// Indices of paths to delete
	pathIdxs []int
	// Deleted paths
	paths []Path
	// Indices of buildings to delete
	buildingIdxs []int
	// Deleted buildings
	buildings []Building
}

func (op *sceneOpDelete) do(s *Scene) {
	s.Paths = SwapDeleteMany(s.Paths, op.pathIdxs)
	s.Buildings = SwapDeleteMany(s.Buildings, op.buildingIdxs)
}

func (op *sceneOpDelete) redo(s *Scene) Action {
	op.do(s)
	return appMode.doSwitchMode(ModeNormal, ResetAll())
}

func (op *sceneOpDelete) undo(s *Scene) Action {
	s.Paths = SwapInsertMany(s.Paths, op.pathIdxs, op.paths)
	s.Buildings = SwapInsertMany(s.Buildings, op.buildingIdxs, op.buildings)
	ss := sceneSubset{
		buildingsIdxs: op.buildingIdxs,
		pathsIdxs:     op.pathIdxs,
	}
	ss.updateBounds()
	return selection.doInitSceneSubset(ss)
}

type sceneOpModify struct {
	pathIdxs      []int
	newPaths      []Path
	oldPaths      []Path
	buildingsIdxs []int
	newBuildings  []Building
	oldBuildings  []Building
}

func (op *sceneOpModify) do(s *Scene) {
	for i, idx := range op.pathIdxs {
		s.Paths[idx] = op.newPaths[i]
	}
	for i, idx := range op.buildingsIdxs {
		s.Buildings[idx] = op.newBuildings[i]
	}
}

func (op *sceneOpModify) redo(s *Scene) Action {
	op.do(s)
	ss := sceneSubset{
		buildingsIdxs: op.buildingsIdxs,
		pathsIdxs:     op.pathIdxs,
	}
	ss.updateBounds()
	return selection.doInitSceneSubset(ss)
}

func (op *sceneOpModify) undo(s *Scene) Action {
	for i, idx := range op.pathIdxs {
		s.Paths[idx] = op.oldPaths[i]
	}
	for i, idx := range op.buildingsIdxs {
		s.Buildings[idx] = op.oldBuildings[i]
	}
	ss := sceneSubset{
		buildingsIdxs: op.buildingsIdxs,
		pathsIdxs:     op.pathIdxs,
	}
	ss.updateBounds()
	return selection.doInitSceneSubset(ss)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// Scene Modifiers methods
////////////////////////////////////////////////////////////////////////////////////////////////////

func (s *Scene) addSceneOp(op sceneOp) {
	rl.TraceLog(rl.LogInfo, fmt.Sprintf("Scene.Operation: %+v", op))
	s.history = s.history[:s.historyPos] // trim any undone operations
	op.redo(s)                           // actually perform the operation
	s.history = append(s.history, op)    // append the operation to the history
	s.historyPos++                       // increment history position
}

// AddPath adds the given path to the scene.
//
// No validity check is performed.
func (s *Scene) AddPath(path Path) {
	s.addSceneOp(&sceneOpAdd{
		paths:     []Path{path},
		buildings: nil,
	})
}

// AddBuilding adds the given building to the scene.
//
// No validity check is performed.
func (s *Scene) AddBuilding(building Building) {
	rl.TraceLog(rl.LogInfo, fmt.Sprintf("Scene.AddBuilding: building=%v", building))
	s.addSceneOp(&sceneOpAdd{
		paths:     nil,
		buildings: []Building{building},
	})
}

// AddObjects adds the given paths and buildings to the scene.
//
// No validity checks is performed.
func (s *Scene) AddObjects(paths []Path, buildings []Building) {
	s.addSceneOp(&sceneOpAdd{
		paths:     slices.Clone(paths),
		buildings: slices.Clone(buildings),
	})
}

// DeleteObjects deletes the given paths and buildings from the scene.
func (s *Scene) DeleteObjects(pathIdxs []int, buildingIdxs []int) {
	s.addSceneOp(&sceneOpDelete{
		pathIdxs:     slices.Clone(pathIdxs),
		paths:        CopyIdxs(s.Paths, pathIdxs),
		buildingIdxs: slices.Clip(buildingIdxs),
		buildings:    CopyIdxs(s.Buildings, buildingIdxs),
	})
}

// ModifyObjects updates the given paths and buildings in the scene.
//
// No validity checks is performed.
func (s *Scene) ModifyObjects(pathIdxs []int, newPaths []Path, buildingIdxs []int, newBuildings []Building) {
	s.addSceneOp(&sceneOpModify{
		pathIdxs:      slices.Clone(pathIdxs),
		newPaths:      slices.Clone(newPaths),
		oldPaths:      CopyIdxs(s.Paths, pathIdxs),
		buildingsIdxs: slices.Clone(buildingIdxs),
		newBuildings:  slices.Clone(newBuildings),
		oldBuildings:  CopyIdxs(s.Buildings, buildingIdxs),
	})
}

// Undo tries to undo the last operation, and returns whether it has, and the action to be performed.
func (s *Scene) Undo() (bool, Action) {
	if s.historyPos > 0 {
		s.historyPos-- // decrement history position
		op := s.history[s.historyPos]
		rl.TraceLog(rl.LogInfo, fmt.Sprintf("Scene.Undo: %#v", op))
		return true, op.undo(s)
	}
	rl.TraceLog(rl.LogInfo, "Scene.Redo: No more operations to undo")
	return false, nil
}

// Redo tries to redo the last undone operation, and returns whether it has, and the action to be performed.
func (s *Scene) Redo() (bool, Action) {
	if s.historyPos < len(s.history) {
		op := s.history[s.historyPos]
		rl.TraceLog(rl.LogInfo, fmt.Sprintf("Scene.Redo: %#v", op))
		s.historyPos++ // increment history position
		return true, op.redo(s)
	}
	rl.TraceLog(rl.LogInfo, "Scene.Redo: No more operations to redo")
	return false, nil
}

// GetObjectAt returns the object at the given position (world coordinates)
//
// If multiple objects are at the same position, returns building over path,
// and the first one in the list.
//
// If no object is found, returns an zero-valued [Object]
func (s Scene) GetObjectAt(pos rl.Vector2) Object {
	for i, b := range s.Buildings {
		if b.Bounds().CheckCollisionPoint(pos) {
			return Object{Type: TypeBuilding, Idx: i}
		}
	}
	for i, p := range s.Paths {
		if p.CheckCollisionPoint(pos) {
			return Object{Type: TypePath, Idx: i}
		}
	}
	return Object{}
}

// SceneIgnore represents groups of inputs that should be ignored in a [Scene.Update] method.
type SceneIgnore struct{ UndoRedo bool }

// Update hovered object
func (s *Scene) Update(ignore SceneIgnore) (action Action) {
	s.Hovered = s.GetObjectAt(mouse.Pos)

	if !ignore.UndoRedo && keyboard.Ctrl {
		switch keyboard.Pressed {
		case rl.KeyZ:
			_, action = s.Undo()
		case rl.KeyY:
			_, action = s.Redo()
		}
	}
	return action
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// Scene hovered object methods
////////////////////////////////////////////////////////////////////////////////////////////////////

func (s Scene) IsBuildingValid(building Building, ignore int) bool {
	bounds := building.Bounds()
	for i, b := range s.Buildings {
		if i == ignore {
			continue
		}
		if b.Bounds().CheckCollisionRec(bounds) {
			return false
		}
	}
	return true
}

func (s Scene) IsPathValid(path Path) bool {
	return !path.Start.Equals(path.End)
}

// Draw scene objects
func (s Scene) Draw() {
	if appMode == ModeSelection {
		stateIt := selection.GetStateIterator(TypePath)
		for _, b := range s.Paths {
			b.Draw(stateIt.Next())
		}
		stateIt = selection.GetStateIterator(TypeBuilding)
		for _, b := range s.Buildings {
			b.Draw(stateIt.Next())
		}
	} else {
		for _, b := range s.Paths {
			b.Draw(DrawNormal)
		}
		for _, b := range s.Buildings {
			b.Draw(DrawNormal)
		}
	}

	// draw hovered object
	if !s.Hovered.IsEmpty() {
		if appMode == ModeNormal && !selector.selecting {
			s.Hovered.Draw(DrawNormal | DrawHovered)
		} else if appMode == ModeSelection && selection.mode == SelectionNormal {
			if selection.Contains(s.Hovered) {
				s.Hovered.Draw(DrawSelected | DrawHovered)
			} else {
				s.Hovered.Draw(DrawNormal | DrawHovered)
			}
		}
	}
}
