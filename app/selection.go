// selection - Handle selection

package app

import (
	"fmt"
	"slices"
	"sort"

	"github.com/bonoboris/satisfied/colors"
	"github.com/bonoboris/satisfied/math32"
	"github.com/bonoboris/satisfied/matrix"
	rl "github.com/gen2brain/raylib-go/raylib"
)

var selection Selection

////////////////////////////////////////////////////////////////////////////////////////////////////
// Selection
////////////////////////////////////////////////////////////////////////////////////////////////////

type Selection struct {
	sceneSubset
	// Selection mode
	mode SelectionMode
	// transformed selection data (in drag or duplicate mode)
	transform selectionTransform
}

func (s *Selection) Reset() {
	s.buildingsIdxs = s.buildingsIdxs[:0]
	s.pathsIdxs = s.pathsIdxs[:0]
	s.mode = SelectionNormal
	s.transform.reset()
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// SelectionMode
////////////////////////////////////////////////////////////////////////////////////////////////////

// Represents a selection sub mode
type SelectionMode int

const (
	// Normal selection
	SelectionNormal SelectionMode = iota
	// Selection is dragged
	SelectionDrag
	// Selection is being duplicated
	SelectionDuplicate
)

func (m SelectionMode) String() string {
	switch m {
	case SelectionNormal:
		return "SelectionNormal"
	case SelectionDrag:
		return "SelectionDrag"
	case SelectionDuplicate:
		return "SelectionDuplicate"
	default:
		return "Invalid"
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// SceneSubset
////////////////////////////////////////////////////////////////////////////////////////////////////

// represents a subset of scene objects
type sceneSubset struct {
	// Indices of path in subset
	pathsIdxs []int
	// Indices of building in subset
	buildingsIdxs []int
	// bounds of the subset (the smallest rectangle that contains all the objects)
	bounds rl.Rectangle
}

// Returns whether scene building at idx is selected
func (s *sceneSubset) ContainsBuilding(idx int) bool {
	return SortedIntsIndex(s.buildingsIdxs, idx) >= 0
}

// Returns whether scene path at idx is selected
func (s *sceneSubset) ContainsPath(idx int) bool { return SortedIntsIndex(s.pathsIdxs, idx) >= 0 }

// Contains returns true if the given object is in the selection (false for TypeInvalid Object)
func (s *sceneSubset) Contains(obj Object) bool {
	switch obj.Type {
	case TypeBuilding:
		return s.ContainsBuilding(obj.Idx)
	case TypePath:
		return s.ContainsPath(obj.Idx)
	default:
		return false
	}
}
func (s *sceneSubset) IsEmpty() bool { return len(s.buildingsIdxs) == 0 && len(s.pathsIdxs) == 0 }

// updateBounds recomputes the outer bounds of the subset
func (s *sceneSubset) updateBounds() {
	if s.IsEmpty() {
		s.bounds = rl.NewRectangle(0, 0, 0, 0)
	}
	xmin, ymin := math32.MaxFloat32, math32.MaxFloat32
	xmax, ymax := -math32.MaxFloat32, -math32.MaxFloat32
	for _, idx := range s.buildingsIdxs {
		b := scene.Buildings[idx]
		bounds := b.Bounds()
		xmin, xmax = min(xmin, bounds.X), max(xmax, bounds.X+bounds.Width)
		ymin, ymax = min(ymin, bounds.Y), max(ymax, bounds.Y+bounds.Height)
	}
	for _, idx := range s.pathsIdxs {
		p := scene.Paths[idx]
		xmin, xmax = min(xmin, min(p.Start.X, p.End.X)), max(xmax, max(p.Start.X, p.End.X))
		ymin, ymax = min(ymin, min(p.Start.Y, p.End.Y)), max(ymax, max(p.Start.Y, p.End.Y))
	}
	s.bounds = rl.NewRectangle(xmin, ymin, xmax-xmin, ymax-ymin)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// selection transform
////////////////////////////////////////////////////////////////////////////////////////////////////

// Represent a transformation of the selection (drag or duplicate)
type selectionTransform struct {
	// Transformation state

	// transformation rotation
	rot int32
	// start position of the transformation
	startPos rl.Vector2
	// end position of the transformation
	endPos rl.Vector2

	// Transformation results / data

	// transformed [Path]
	paths []Path
	// transformed [Building]
	buildings []Building
	// invalid transformed paths mask
	invalidPaths []bool
	// invalid transformed buildings mask
	invalidBuildings []bool
	// whether the every transformed object is valid
	isValid bool
	// bounds of the transformed selection
	bounds rl.Rectangle
}

func (s *selectionTransform) reset() {
	s.rot = 0
	s.startPos = rl.Vector2{}
	s.endPos = rl.Vector2{}

	s.paths = s.paths[:0]
	s.buildings = s.buildings[:0]
	s.invalidPaths = s.invalidPaths[:0]
	s.invalidBuildings = s.invalidBuildings[:0]
	s.isValid = false
	s.bounds = rl.Rectangle{}
}

// Matrix returns the rotation matrix of the selection
func (s selectionTransform) transformMatrix(baseBounds rl.Rectangle) matrix.Matrix {
	center := baseBounds.Center()
	translate := grid.Snap(s.endPos.Subtract(s.startPos))
	return matrix.NewTranslateV(translate.Add(center)).Rotate(s.rot).TranslateV(center.Negate())
}

// recompute recomputes the transformed objects and whether they are valid
func (s *selectionTransform) recompute(ss sceneSubset, mode SelectionMode) {
	s.isValid = true
	// resize slices
	s.buildings = slices.Grow(s.buildings[:0], len(ss.buildingsIdxs))
	s.invalidBuildings = slices.Grow(s.invalidBuildings[:0], len(ss.buildingsIdxs))
	s.paths = slices.Grow(s.paths[:0], len(ss.pathsIdxs))
	s.invalidPaths = slices.Grow(s.invalidPaths[:0], len(ss.pathsIdxs))

	translate := grid.Snap(s.endPos.Subtract(s.startPos))
	// fast path for identity transform
	if translate.X == 0 && translate.Y == 0 && s.rot%360 == 0 {
		// invalid when duplicating, valid otherwise
		invalid := mode == SelectionDuplicate
		for _, idx := range ss.buildingsIdxs {
			s.buildings = append(s.buildings, scene.Buildings[idx])
			s.invalidBuildings = append(s.invalidBuildings, invalid)
		}
		for _, idx := range ss.pathsIdxs {
			s.paths = append(s.paths, scene.Paths[idx])
			s.invalidPaths = append(s.invalidPaths, invalid)
		}
		s.isValid = !invalid
		s.bounds = ss.bounds
		return
	}

	mat := s.transformMatrix(ss.bounds)
	s.bounds = mat.ApplyRecRec(ss.bounds)

	// Buildings
	for _, idx := range ss.buildingsIdxs {
		b := scene.Buildings[idx]
		s.buildings = append(s.buildings, Building{
			DefIdx: b.DefIdx,
			Pos:    mat.ApplyV(b.Pos),
			Rot:    (b.Rot + s.rot) % 360,
		})
	}

	// Paths
	for _, idx := range ss.pathsIdxs {
		p := scene.Paths[idx]
		np := Path{
			DefIdx: p.DefIdx,
			Start:  mat.ApplyV(p.Start),
			End:    mat.ApplyV(p.End),
		}
		s.paths = append(s.paths, np)
		s.invalidPaths = append(s.invalidPaths, false) // TODO proper check when path anchor done
	}

	// precompute selection building bounds and initialize invalid to false
	selectionBounds := make([]rl.Rectangle, len(ss.buildingsIdxs))
	for i := range len(ss.buildingsIdxs) {
		s.invalidBuildings = append(s.invalidBuildings, false)
		selectionBounds[i] = s.buildings[i].Bounds()
	}

	for i, sb := range scene.Buildings {
		// TODO: Use iterator instead of [sceneSubset.ContainsBuilding]
		sb := sb.Bounds()
		if mode != SelectionDuplicate && ss.ContainsBuilding(i) || !s.bounds.CheckCollisionRec(sb) {
			// skip:
			//   - scene building in selection (except when duplicating)
			//   - scene building outside transformation outer bounds
			continue
		} else {
			// check against every transformed building
			for i, bounds := range selectionBounds {
				if !s.invalidBuildings[i] && bounds.CheckCollisionRec(sb) { // no need to call CheckCollisionRec if s.invalidBuildings[i] is already true
					s.isValid = false
					s.invalidBuildings[i] = true
				}
			}
		}
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// Selection methods
////////////////////////////////////////////////////////////////////////////////////////////////////

// GetAction processes inputs in [ModeSelection] and returns an action to be performed.
//
// See: [GetActionFunc]
func (s *Selection) GetAction() Action {
	appMode.Assert(ModeSelection)
	switch s.mode {
	case SelectionNormal:
		switch keyboard.Pressed {
		case rl.KeyD:
			// Duplicate use center of current selection as start position
			return s.doBeginTransformation(SelectionDuplicate, s.bounds.Center())
		case rl.KeyDelete, rl.KeyX:
			return s.doDelete()
		case rl.KeyR:
			return s.doRotate()
		case rl.KeyLeft:
			return s.doMoveBy(vec2(-1, 0))
		case rl.KeyRight:
			return s.doMoveBy(vec2(+1, 0))
		case rl.KeyUp:
			return s.doMoveBy(vec2(0, -1))
		case rl.KeyDown:
			return s.doMoveBy(vec2(0, +1))
		}
		if mouse.Left.Pressed && mouse.InScene {
			switch {
			case scene.Hovered.IsEmpty():
				return selector.doInit(mouse.Pos)
			case selection.Contains(scene.Hovered):
				// Drag use mouse position as start position
				return s.doBeginTransformation(SelectionDrag, mouse.Pos)
			default:
				return s.doInitSingle(scene.Hovered, SelectionDrag, mouse.Pos)
			}
		}
	case SelectionDuplicate:
		// TODO: Implement arrow keys nudging ?
		if keyboard.Pressed == rl.KeyR {
			return s.doRotate()
		}
		if mouse.Left.Released {
			return s.doEndTransformation(false)
		}
		if mouse.InScene && !mouse.Left.Down {
			return s.doMoveTo(mouse.Pos)
		}
	case SelectionDrag:
		if keyboard.Pressed == rl.KeyR {
			return s.doRotate()
		}
		if mouse.Left.Released {
			return s.doEndTransformation(false)
		}
		if mouse.InScene && mouse.Left.Down {
			return s.doMoveTo(mouse.Pos)
		}
	}
	return nil
}

func (s *Selection) doInitSingle(obj Object, mode SelectionMode, dragPos rl.Vector2) Action {
	s.Reset()
	switch obj.Type {
	case TypeBuilding:
		s.buildingsIdxs = append(s.buildingsIdxs, obj.Idx)
		s.bounds = scene.Buildings[obj.Idx].Bounds()
	case TypePath:
		s.pathsIdxs = append(s.pathsIdxs, obj.Idx)
		path := scene.Paths[obj.Idx]
		s.bounds = rl.NewRectangleCorners(path.Start, path.End)
	default:
		panic("invalid object type")
	}
	s.mode = mode
	if mode == SelectionDrag {
		s.transform.startPos = dragPos
		s.transform.endPos = dragPos
	} else {
		s.transform.startPos = s.bounds.Center()
		s.transform.endPos = s.bounds.Center()
	}
	s.transform.recompute(s.sceneSubset, mode) // noop transformation ->uses fast path
	return appMode.doSwitchMode(ModeSelection, ResetAll().WithSelection(false))
}

// doInitSceneSubset initializes a new selection from a subset of the scene, in [SelectionNormal] mode
func (s *Selection) doInitSceneSubset(ss sceneSubset) Action {
	s.Reset()
	s.buildingsIdxs = ss.buildingsIdxs
	s.pathsIdxs = ss.pathsIdxs
	s.bounds = ss.bounds
	s.mode = SelectionNormal
	s.transform.startPos = s.bounds.Center()
	s.transform.endPos = s.bounds.Center()
	s.transform.recompute(s.sceneSubset, s.mode) // noop transformation ->uses fast path
	return appMode.doSwitchMode(ModeSelection, ResetAll().WithSelection(false))
}

func (s *Selection) doInitRectangle(rect rl.Rectangle, mode SelectionMode, dragPos rl.Vector2) Action {
	s.Reset()
	for i, b := range scene.Buildings {
		bounds := b.Bounds()
		if rect.CheckCollisionPoint(bounds.TopLeft()) && rect.CheckCollisionPoint(bounds.BottomRight()) {
			s.buildingsIdxs = append(s.buildingsIdxs, i)
		}
	}
	sort.Ints(s.buildingsIdxs)
	for i, p := range scene.Paths {
		if rect.CheckCollisionPoint(p.Start) && rect.CheckCollisionPoint(p.End) {
			s.pathsIdxs = append(s.pathsIdxs, i)
		}
	}
	sort.Ints(s.pathsIdxs)
	s.updateBounds()
	s.mode = mode
	if mode == SelectionDrag {
		s.transform.startPos = dragPos
		s.transform.endPos = dragPos
	} else {
		s.transform.startPos = s.bounds.Center()
		s.transform.endPos = s.bounds.Center()
	}
	s.transform.recompute(s.sceneSubset, mode) // noop transformation -> uses fast path
	return appMode.doSwitchMode(ModeSelection, ResetAll().WithSelection(false))
}

func (s *Selection) doDelete() Action {
	appMode.Assert(ModeSelection)
	assert(s.mode == SelectionNormal, "cannot delete selection in "+s.mode.String())
	if len(s.buildingsIdxs) > 0 {
		scene.Buildings = SwapDeleteMany(scene.Buildings, s.buildingsIdxs)
	}
	if len(s.pathsIdxs) > 0 {
		scene.Paths = SwapDeleteMany(scene.Paths, s.pathsIdxs)
	}
	if s.Contains(scene.Hovered) {
		// recompute hovered object in case it was deleted
		scene.Hovered = scene.GetObjectAt(mouse.Pos)
	}
	return appMode.doSwitchMode(ModeNormal, ResetAll())
}

func (s *Selection) doBeginTransformation(mode SelectionMode, pos rl.Vector2) Action {
	rl.TraceLog(rl.LogWarning, fmt.Sprintf("SelectionActionBeginTransformation: mode=%v pos=%v", mode, pos))
	appMode.Assert(ModeSelection)
	s.transform.reset()
	s.mode = mode
	s.transform.startPos = pos
	s.transform.endPos = pos
	s.transform.recompute(s.sceneSubset, mode) // noop transformation -> uses fast path
	return nil
}

func (s *Selection) doMoveBy(delta rl.Vector2) Action {
	appMode.Assert(ModeSelection)
	if s.mode == SelectionNormal {
		s.transform.startPos = s.bounds.Center()
		s.transform.endPos = s.bounds.Center().Add(delta)
		s.transform.rot = 0
		s.transform.recompute(s.sceneSubset, s.mode)
		// will apply the translation if it's valid and reset transformation in any case
		return s.doEndTransformation(false)
	} else {
		s.transform.endPos = s.transform.endPos.Add(delta)
	}
	return nil
}

func (s *Selection) doMoveTo(pos rl.Vector2) Action {
	appMode.Assert(ModeSelection)
	s.transform.endPos = pos
	s.transform.recompute(s.sceneSubset, s.mode)
	return nil
}

func (s *Selection) doRotate() Action {
	rl.TraceLog(rl.LogWarning, fmt.Sprintf("SelectionActionRotate: mode=%v", s.mode))
	appMode.Assert(ModeSelection)
	if s.mode == SelectionNormal {
		// try rotate the selection
		s.transform.startPos = s.bounds.Center()
		s.transform.endPos = s.bounds.Center()
		s.transform.rot = 90
		s.transform.recompute(s.sceneSubset, s.mode)
		// will apply the rotation if it's valid and reset transformation in any case
		return s.doEndTransformation(false)
	}

	s.transform.rot += 90
	s.transform.recompute(s.sceneSubset, s.mode)
	return nil
}

func (s *Selection) doEndTransformation(discard bool) Action {
	rl.TraceLog(rl.LogWarning,
		fmt.Sprintf("SelectionActionEndTransformation: mode=%v discard=%v isValid=%v rot=%v translate=%v",
			s.mode, discard, s.transform.isValid, s.transform.rot, s.transform.endPos.Subtract(s.transform.startPos)),
	)
	appMode.Assert(ModeSelection)
	if !discard && s.transform.isValid {
		switch s.mode {
		case SelectionNormal, SelectionDrag:
			scene.ModifyObjects(s.pathsIdxs, s.transform.paths, s.buildingsIdxs, s.transform.buildings)
			s.bounds = s.transform.bounds
		case SelectionDuplicate:
			scene.AddObjects(s.transform.paths, s.transform.buildings)
		}
	}
	s.transform.reset()
	s.mode = SelectionNormal
	return nil
}

// Dispatch performs a [Selection] action, updating its state, and returns an new action to be performed
//
// See: [ActionHandler]
func (s *Selection) Dispatch(action Action) Action {
	switch action := action.(type) {
	case SelectionActionInitSingle:
		return s.doInitSingle(action.Object, action.Mode, action.DragPos)
	case SelectionActionInitRectangle:
		return s.doInitRectangle(action.Rect, action.Mode, action.DragPos)
	case SelectionActionDelete:
		return s.doDelete()
	case SelectionActionBeginTransformation:
		return s.doBeginTransformation(action.Mode, action.Pos)
	case SelectionActionMoveTo:
		return s.doMoveTo(action.Pos)
	case SelectionActionRotate:
		return s.doRotate()
	case SelectionActionEndTransformation:
		return s.doEndTransformation(action.Discard)

	default:
		panic(fmt.Sprintf("Selection.Dispatch: cannot handle: %T", action))
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// Draw
////////////////////////////////////////////////////////////////////////////////////////////////////

func (s *selectionTransform) draw(state DrawState) {
	bounds := s.bounds
	px := 1 / camera.Zoom()
	if s.isValid {
		rl.DrawRectangleLinesEx(bounds, 3*px, colors.WithAlpha(colors.Blue500, 0.5))
	} else {
		rl.DrawRectangleLinesEx(bounds, 3*px, colors.WithAlpha(colors.Red500, 0.5))
		rl.DrawRectangleRec(bounds, colors.WithAlpha(colors.Red500, 0.1))
	}
	for i, p := range s.paths {
		if s.invalidPaths[i] {
			p.Draw(DrawInvalid)
		} else {
			p.Draw(state)
		}
	}
	for i, b := range s.buildings {
		if s.invalidBuildings[i] {
			b.Draw(DrawInvalid)
		} else {
			b.Draw(state)
		}
	}
}

// Draws the selection (if transformation if any)
func (s Selection) Draw() {
	switch s.mode {
	case SelectionNormal:
		// only draw the selection rectangle, buildings and paths are drawn in [Scene.Draw]
		px := 1 / camera.Zoom()
		rl.DrawRectangleLinesEx(s.bounds, 3*px, colors.WithAlpha(colors.Blue500, 0.5))
	case SelectionDrag:
		s.transform.draw(DrawClicked)
	case SelectionDuplicate:
		s.transform.draw(DrawNew)

	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// Iterators
////////////////////////////////////////////////////////////////////////////////////////////////////

// Iterate over a would-be mask from the true values indices.
type MaskIterator struct {
	// index for which to return true
	TrueIdxs []int
	// current index counter
	Idx int
	// index of the next true value in [MaskIterator.TrueIdxs]
	i int
}

// NewMaskIterator returns a new [MaskIterator] from the given true values indices.
//
// trueIdxs must be sorted in ascending order.
func NewMaskIterator(trueIdx []int) MaskIterator {
	return MaskIterator{TrueIdxs: trueIdx, Idx: 0, i: 0}
}

// Next returns the next value from the mask.
//
// It always returns false when all the true values have been iterated over, but keep incrementing
// [MaskIterator.Idx] counter.
func (it *MaskIterator) Next() bool {
	if it.i == len(it.TrueIdxs) {
		// no more selection
		it.Idx++
		return false
	}
	if it.Idx == it.TrueIdxs[it.i] {
		it.i++ // advance to next true value
		it.Idx++
		return true
	} else {
		it.Idx++
		return false
	}
}

// DrawStateIterator is a helper to get the next scene building/path draw state.
//
// It optimizes checking whether a scene object is in the selection assuming:
//   - [Selection.buildingsIdxs] and [Selection.pathsIdxs] are sorted
//   - Scene building/path are iterated in order
//
// This results in O(len(scene)) complexity.
type drawStateIterator struct {
	selected MaskIterator
	// state to return for selected buildings/paths
	state DrawState
}

func newDrawStateIterator(selectedIdxs []int, state DrawState) drawStateIterator {
	return drawStateIterator{
		selected: NewMaskIterator(selectedIdxs),
		state:    state,
	}
}

func (it *drawStateIterator) Next() DrawState {
	if it.selected.Next() {
		// it.selected.i is the index of the next selected building/path in selection
		// we want the current
		return it.state
	} else {
		// not selected, return normal state
		return DrawNormal
	}
}

// Returns a [drawStateIterator] for the given [ObjectType].
func (s *Selection) GetStateIterator(objectType ObjectType) drawStateIterator {
	var selectedIdxs []int
	switch objectType {
	case TypePath:
		selectedIdxs = s.pathsIdxs
	case TypeBuilding:
		selectedIdxs = s.buildingsIdxs
	default:
		panic("selectionTransform.GetStateIterator: invalid objectType")
	}

	var state DrawState
	switch s.mode {
	case SelectionNormal:
		state = DrawSelected
	case SelectionDrag:
		state = DrawShadow
	case SelectionDuplicate:
		state = DrawClicked
	default:
		panic("selectionTransform.GetStateIterator: invalid selection mode")
	}
	return newDrawStateIterator(selectedIdxs, state)
}
