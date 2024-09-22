// selection - Handle selection

package app

import (
	"fmt"
	"slices"

	"github.com/bonoboris/satisfied/colors"
	"github.com/bonoboris/satisfied/log"
	"github.com/bonoboris/satisfied/matrix"
	rl "github.com/gen2brain/raylib-go/raylib"
)

var selection Selection

////////////////////////////////////////////////////////////////////////////////////////////////////
// Selection
////////////////////////////////////////////////////////////////////////////////////////////////////

type Selection struct {
	ObjectSelection
	// Selection mode
	mode SelectionMode
	// transformed selection data (in drag or duplicate mode)
	transform selectionTransform
}

func (s Selection) traceState(key, val string) {
	if log.WillTrace() {
		if key != "" && val != "" {
			log.Trace("selection", key, val, "mode", s.mode)
		}
		log.Trace("selection", "buildingIdxs", s.BuildingIdxs)
		log.Trace("selection", "pathIdxs", s.PathIdxs)
		log.Trace("selection", "mode", s.mode, "bounds", s.Bounds)
		s.transform.traceState()
	}
}

func (s *Selection) Reset() {
	log.Debug("selection.reset")
	s.BuildingIdxs = s.BuildingIdxs[:0]
	s.PathIdxs = s.PathIdxs[:0]
	s.Bounds = rl.NewRectangle(0, 0, 0, 0)
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
	ObjectCollection
	// invalid transformed paths mask
	invalidPaths []bool
	// invalid transformed buildings mask
	invalidBuildings []bool
	// whether the every transformed object is valid
	isValid bool
	// bounds of the transformed selection
	bounds rl.Rectangle

	// transformed building bounds buffer (reduce allocs)
	_buildingBounds []rl.Rectangle
}

func (st selectionTransform) traceState() {
	if log.WillTrace() {
		for i, p := range st.Paths {
			log.Trace("selectionTransform.paths", "i", i, "value", p, "invalid", st.invalidPaths[i])
		}
		for i, b := range st.Buildings {
			log.Trace("selectionTransform.buildings", "i", i, "value", b, "invalid", st.invalidBuildings[i])
		}
		log.Trace("selectionTransform", "isValid", st.isValid, "bounds", st.bounds)
		log.Trace("selectionTransform", "rot", st.rot, "startPos", st.startPos, "endPos", st.endPos)
	}
}

func (st *selectionTransform) reset() {
	st.rot = 0
	st.startPos = rl.Vector2{}
	st.endPos = rl.Vector2{}

	st.Paths = st.Paths[:0]
	st.Buildings = st.Buildings[:0]
	st.invalidPaths = st.invalidPaths[:0]
	st.invalidBuildings = st.invalidBuildings[:0]
	st.isValid = false
	st.bounds = rl.Rectangle{}
}

func (s selectionTransform) isIdentity() bool {
	translate := grid.Snap(s.endPos.Subtract(s.startPos))
	return s.rot%360 == 0 && translate.X == 0 && translate.Y == 0
}

// Matrix returns the rotation matrix of the selection
func (s selectionTransform) transformMatrix(baseBounds rl.Rectangle) matrix.Matrix {
	center := baseBounds.Center()
	translate := grid.Snap(s.endPos.Subtract(s.startPos))
	return matrix.NewTranslateV(translate.Add(center)).Rotate(s.rot).TranslateV(center.Negate())
}

// recompute recomputes the transformed objects and whether they are valid
func (st *selectionTransform) recompute(sel ObjectSelection, mode SelectionMode) {
	// fast path for identity transform
	// TODO: not copying anything would be faster
	if st.isIdentity() {
		switch mode {
		case SelectionDuplicate:
			pathIdxs := sel.FullPathIdxs()
			st.Buildings = CopyIdxs(st.Buildings, scene.Buildings, sel.BuildingIdxs)
			st.invalidBuildings = Repeat(st.invalidBuildings, true, len(sel.BuildingIdxs))
			st.Paths = CopyIdxs(st.Paths, scene.Paths, pathIdxs)
			st.invalidPaths = Repeat(st.invalidPaths, true, len(pathIdxs))
			st.isValid = false
			st.bounds = sel.Bounds
		case SelectionDrag, SelectionNormal:
			pathIdxs := sel.AnyPathIdxs()
			st.Buildings = CopyIdxs(st.Buildings, scene.Buildings, sel.BuildingIdxs)
			st.invalidBuildings = Repeat(st.invalidBuildings, false, len(sel.BuildingIdxs))
			st.Paths = CopyIdxs(st.Paths, scene.Paths, pathIdxs)
			st.invalidPaths = Repeat(st.invalidPaths, false, len(pathIdxs))
			st.isValid = true
			st.bounds = sel.Bounds
		}
		return
	}

	// TODO: store transformMatrix and recompute only when needed

	st.isValid = true

	var pathIdxs []int
	if mode == SelectionDuplicate {
		// we only want to duplicate paths that are entirely inside the selection
		pathIdxs = sel.FullPathIdxs()
	} else {
		pathIdxs = sel.AnyPathIdxs()
	}

	nb := len(sel.BuildingIdxs)
	np := len(pathIdxs)

	// clears slices
	st.Buildings = slices.Grow(st.Buildings[:0], nb)
	st.invalidBuildings = slices.Grow(st.invalidBuildings[:0], nb)
	st.Paths = slices.Grow(st.Paths[:0], np)
	st.invalidPaths = slices.Grow(st.invalidPaths[:0], np)
	st._buildingBounds = slices.Grow(st._buildingBounds[:0], nb)

	mat := st.transformMatrix(sel.Bounds)
	st.bounds = mat.ApplyRecRec(sel.Bounds)

	// Buildings
	for _, idx := range sel.BuildingIdxs {
		b := scene.Buildings[idx]
		b.Pos = mat.ApplyV(b.Pos)
		b.Rot = (b.Rot + st.rot) % 360
		st.Buildings = append(st.Buildings, b)
	}

	// Paths & invalidPaths
	switch mode {
	case SelectionDuplicate:
		for _, idx := range pathIdxs {
			p := scene.Paths[idx]
			p.Start = mat.ApplyV(p.Start)
			p.End = mat.ApplyV(p.End)
			st.Paths = append(st.Paths, p)
			if p.IsValid() {
				st.invalidPaths = append(st.invalidPaths, false)
			} else {
				st.invalidPaths = append(st.invalidPaths, true)
				st.isValid = false
			}
		}
	case SelectionDrag, SelectionNormal:
		for _, elt := range sel.PathIdxs {
			p := scene.Paths[elt.Idx]
			if elt.Start {
				p.Start = mat.ApplyV(p.Start)
			}
			if elt.End {
				p.End = mat.ApplyV(p.End)
			}
			st.Paths = append(st.Paths, p)
			if p.IsValid() {
				st.invalidPaths = append(st.invalidPaths, false)
			} else {
				st.invalidPaths = append(st.invalidPaths, true)
				st.isValid = false
			}
		}
	}

	// precompute transformed building bounds & initialize invalid to false
	for i := range nb {
		st._buildingBounds = append(st._buildingBounds, st.Buildings[i].Bounds())
		st.invalidBuildings = append(st.invalidBuildings, false)
	}

	isSelectedIt := NewMaskIterator(sel.BuildingIdxs)
	for _, sb := range scene.Buildings {
		sb := sb.Bounds()
		if mode != SelectionDuplicate && isSelectedIt.Next() || !st.bounds.CheckCollisionRec(sb) {
			// skip:
			//   - scene building in selection (except when duplicating)
			//   - scene building outside transformation outer bounds
			continue
		} else {
			// check against every transformed building
			for i, bounds := range st._buildingBounds {
				// no need to call CheckCollisionRec if st.invalidBuildings[i] is already true
				if !st.invalidBuildings[i] && bounds.CheckCollisionRec(sb) {
					st.isValid = false
					st.invalidBuildings[i] = true
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
	app.Mode.Assert(ModeSelection)
	switch s.mode {
	case SelectionNormal:
		switch keyboard.Pressed {
		case rl.KeyEscape:
			return app.doSwitchMode(ModeNormal, ResetAll())
		case rl.KeyD:
			// Duplicate use center of current selection as start position
			return s.doBeginTransformation(SelectionDuplicate, s.Bounds.Center())
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
		switch keyboard.Pressed {
		case rl.KeyEscape:
			return s.doEndTransformation(true)
		case rl.KeyR:
			return s.doRotate()
		}
		if mouse.Left.Released {
			return s.doEndTransformation(false)
		}
		if mouse.InScene && !mouse.Left.Down {
			return s.doMoveTo(mouse.Pos)
		}
	case SelectionDrag:
		switch keyboard.Pressed {
		case rl.KeyEscape:
			return s.doEndTransformation(true)
		case rl.KeyR:
			return s.doRotate()
		}
		if mouse.Left.Released {
			return s.doEndTransformation(false)
		}
		if mouse.InScene {
			return s.doMoveTo(mouse.Pos)
		}
	}
	return nil
}

func (s *Selection) doInitSingle(obj Object, mode SelectionMode, dragPos rl.Vector2) Action {
	log.Debug("selection.doInitSingle", "obj", obj, "mode", mode, "dragPos", dragPos)
	s.Reset()

	switch obj.Type {
	case TypeBuilding:
		s.BuildingIdxs = append(s.BuildingIdxs, obj.Idx)
		s.Bounds = scene.Buildings[obj.Idx].Bounds()
	case TypePath:
		s.PathIdxs = append(s.PathIdxs, PathSel{Idx: obj.Idx, Start: true, End: true})
		s.Bounds = rl.NewRectangleCorners(scene.Paths[obj.Idx].Start, scene.Paths[obj.Idx].End)
	case TypePathStart:
		s.PathIdxs = append(s.PathIdxs, PathSel{Idx: obj.Idx, Start: true})
		s.Bounds = rl.NewRectangleV(scene.Paths[obj.Idx].Start, rl.Vector2{})
	case TypePathEnd:
		s.PathIdxs = append(s.PathIdxs, PathSel{Idx: obj.Idx, End: true})
		s.Bounds = rl.NewRectangleV(scene.Paths[obj.Idx].End, rl.Vector2{})

	default:
		panic("invalid object type")
	}
	s.mode = mode
	if mode == SelectionDrag {
		s.transform.startPos = dragPos
		s.transform.endPos = dragPos
	} else {
		center := s.Bounds.Center()
		s.transform.startPos = center
		s.transform.endPos = center
	}
	s.transform.recompute(s.ObjectSelection, mode) // noop transformation ->uses fast path

	s.traceState("after", "doBeginTransformation")
	return app.doSwitchMode(ModeSelection, ResetAll().WithSelection(false))
}

// doInitSelection initializes a new selection from an [ObjectSelection], in [SelectionNormal] mode
func (s *Selection) doInitSelection(sel ObjectSelection) Action {
	log.Debug("selection.doInitSelection", "selected", sel)
	s.Reset()

	if sel.IsEmpty() {
		s.traceState("after", "doInitSelection")
		return app.doSwitchMode(ModeNormal, ResetAll())
	}

	s.BuildingIdxs = append(s.BuildingIdxs, sel.BuildingIdxs...)
	s.PathIdxs = append(s.PathIdxs, sel.PathIdxs...)
	s.Bounds = sel.Bounds
	s.mode = SelectionNormal
	s.transform.startPos = s.Bounds.Center()
	s.transform.endPos = s.Bounds.Center()
	s.transform.recompute(s.ObjectSelection, s.mode) // noop transformation ->uses fast path

	s.traceState("after", "doInitSelection")
	return app.doSwitchMode(ModeSelection, ResetAll().WithSelection(false))
}

func (s *Selection) doInitRectangle(rect rl.Rectangle, mode SelectionMode, dragPos rl.Vector2) Action {
	log.Debug("selection.doInitRectangle", "rect", rect, "mode", mode, "dragPos", dragPos)
	s.Reset()
	scene.SelectFromRect(&s.ObjectSelection, rect)

	if s.IsEmpty() {
		s.traceState("after", "doInitRectangle")
		return app.doSwitchMode(ModeNormal, ResetAll())
	}
	s.mode = mode
	if mode == SelectionDrag {
		s.transform.startPos = dragPos
		s.transform.endPos = dragPos
	} else {
		center := s.Bounds.Center()
		s.transform.startPos = center
		s.transform.endPos = center
	}
	s.transform.recompute(s.ObjectSelection, mode) // noop transformation -> uses fast path

	s.traceState("after", "doInitSingle")
	return app.doSwitchMode(ModeSelection, ResetAll().WithSelection(false))
}

func (s *Selection) doDelete() Action {
	s.traceState("before", "doDelete")
	log.Debug("selection.doDelete")
	app.Mode.Assert(ModeSelection)
	assert(s.mode == SelectionNormal, "cannot delete selection in "+s.mode.String())

	scene.DeleteObjects(s.ObjectSelection)

	s.traceState("after", "doDelete")
	return app.doSwitchMode(ModeNormal, ResetAll())
}

func (s *Selection) doBeginTransformation(mode SelectionMode, pos rl.Vector2) Action {
	s.traceState("before", "doBeginTransformation")
	log.Debug("selection.doBeginTransformation", "mode", mode, "pos", pos)
	app.Mode.Assert(ModeSelection)

	// special cases for only path start/end selected for duplicate
	if mode == SelectionDuplicate && len(s.BuildingIdxs) == 0 {
		noFullPath := true
		for _, elt := range s.PathIdxs {
			if elt.Start && elt.End {
				noFullPath = false
				break
			}
		}
		if noFullPath {
			if len(s.PathIdxs) == 1 {
				// only one path start/end selected switch to newpath mode
				log.Debug("selection.doBeginTransformation", "action", "newpath", "mode", mode, "reason", "single path ending selected")
				return newPath.doInit(s.PathIdxs[0].Idx)
			} else {
				// TODO: would be nice to have a multi newpath mode
				// for now do nothing
				log.Debug("selection.doBeginTransformation", "action", "skipped", "mode", mode, "reason", "only path endings selected")
				s.traceState("after", "doBeginTransformation")
				return nil
			}
		}
	}

	s.transform.reset()

	s.mode = mode
	s.transform.startPos = pos
	s.transform.endPos = pos
	s.transform.recompute(s.ObjectSelection, mode) // noop transformation -> uses fast path

	s.traceState("after", "doBeginTransformation")
	return nil
}

func (s *Selection) doMoveBy(delta rl.Vector2) Action {
	s.traceState("before", "doMoveBy")
	log.Debug("selection.doMoveBy", "delta", delta, "selection.mode", s.mode)
	app.Mode.Assert(ModeSelection)

	if s.mode == SelectionNormal {
		center := s.Bounds.Center()
		s.transform.startPos = center
		s.transform.endPos = center.Add(delta)
		s.transform.rot = 0
		s.transform.recompute(s.ObjectSelection, s.mode)
		s.traceState("after", "doMoveBy")
		return s.doEndTransformation(false)
	}
	s.transform.endPos = s.transform.endPos.Add(delta)

	s.traceState("after", "doMoveBy")
	return nil
}

func (s *Selection) doMoveTo(pos rl.Vector2) Action {
	s.traceState("before", "doMoveTo")
	log.Trace("selection.doMoveTo", "pos", pos) // moving by mouse -> tracing
	app.Mode.Assert(ModeSelection)

	s.transform.endPos = pos
	s.transform.recompute(s.ObjectSelection, s.mode)

	s.traceState("after", "doMoveTo")
	return nil
}

func (s *Selection) doRotate() Action {
	s.traceState("before", "doRotate")
	log.Debug("selection.doRotate", "selection.mode", s.mode)
	app.Mode.Assert(ModeSelection)

	if s.mode == SelectionNormal {
		// try rotate the selection
		center := s.Bounds.Center()
		s.transform.startPos = center
		s.transform.endPos = center
		s.transform.rot = 90
		s.transform.recompute(s.ObjectSelection, s.mode)
		// will apply the rotation if it's valid and reset transformation in any case
		s.traceState("after", "doRotate")
		return s.doEndTransformation(false)
	}

	s.transform.rot += 90
	s.transform.recompute(s.ObjectSelection, s.mode)

	s.traceState("after", "doRotate")
	return nil
}

func (s *Selection) doEndTransformation(discard bool) Action {
	s.traceState("before", "doEndTransformation")
	log.Debug("selection.doEndTransformation", "discard", discard, "selection.mode", s.mode)
	app.Mode.Assert(ModeSelection)

	switch s.mode {

	case SelectionNormal, SelectionDrag:
		if !discard && !s.transform.isIdentity() && s.transform.isValid {
			scene.ModifyObjects(s.ObjectSelection, s.transform.ObjectCollection)
			s.Bounds = s.transform.bounds
		}
		// in any case, reset mode & transform
		s.transform.reset()
		s.mode = SelectionNormal

	case SelectionDuplicate:
		if discard {
			// on discard -> reset mode & transform
			s.transform.reset()
			s.mode = SelectionNormal
		} else if !s.transform.isIdentity() && s.transform.isValid {
			// stays in [SelectionDuplicate] mode
			scene.AddObjects(s.transform.ObjectCollection)
		}
	}

	s.traceState("after", "doEndTransformation")
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

// draws the transformed selection
func (s *selectionTransform) draw(state DrawState) {
	bounds := s.bounds
	px := 1 / camera.Zoom()
	if s.isValid {
		rl.DrawRectangleLinesEx(bounds, 3*px, colors.WithAlpha(colors.Blue500, 0.5))
	} else {
		rl.DrawRectangleLinesEx(bounds, 3*px, colors.WithAlpha(colors.Red500, 0.5))
		rl.DrawRectangleRec(bounds, colors.WithAlpha(colors.Red500, 0.15))
	}
	for i, p := range s.Paths {
		if s.invalidPaths[i] {
			p.Draw(DrawInvalid)
		} else {
			p.Draw(state)
		}
	}
	for i, b := range s.Buildings {
		if s.invalidBuildings[i] {
			b.Draw(DrawInvalid)
		} else {
			b.Draw(state)
		}
	}
}

// Draws the selection (and the transformed selection if any)
func (s Selection) Draw() {
	switch s.mode {
	case SelectionNormal:
		// only draw the selection rectangle, buildings and paths are drawn in [Scene.Draw]
		px := 1 / camera.Zoom()
		rl.DrawRectangleLinesEx(s.Bounds, 3*px, colors.WithAlpha(colors.Blue500, 0.5))
	case SelectionDrag:
		s.transform.draw(DrawClicked)
	case SelectionDuplicate:
		s.transform.draw(DrawNew)

	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// Iterators
////////////////////////////////////////////////////////////////////////////////////////////////////

// BuildingDrawStateIterator returns an iterator over the draw state of scene buildings
func (s Selection) BuildingDrawStateIterator() buildingDrawStateIterator {
	var state DrawState
	switch s.mode {
	case SelectionNormal:
		state = DrawSkip // drawn later on top
	case SelectionDrag:
		state = DrawShadow
	case SelectionDuplicate:
		state = DrawClicked
	default:
		panic("invalid selection mode")
	}
	return buildingDrawStateIterator{
		selectedIt: NewMaskIterator(s.BuildingIdxs),
		state:      state,
	}
}

// PathDrawStateIterator returns an iterator over the draw state of scene paths start, end and body
func (s Selection) PathDrawStateIterator() pathDrawStateIterator {
	var state DrawState
	switch s.mode {
	case SelectionNormal:
		state = DrawSkip // drawn later on top
	case SelectionDrag:
		state = DrawShadow
	case SelectionDuplicate:
		state = DrawClicked
	default:
		panic("invalid selection mode")
	}
	return pathDrawStateIterator{
		pathsIdxs: s.PathIdxs,
		state:     state,
		mode:      s.mode,
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// pathDrawStateIterator
////////////////////////////////////////////////////////////////////////////////////////////////////

type pathDrawStateIterator struct {
	// selected paths
	pathsIdxs []PathSel
	// selection mode
	mode SelectionMode
	// state to return for selected path start/end/body
	state DrawState
	// Current scene path index to consider
	idx int
	// current index in pathsIdxs
	i int
}

// Next returns the triplet of (start, end, body) draw state
func (it *pathDrawStateIterator) Next() (DrawState, DrawState, DrawState) {
	if it.i == len(it.pathsIdxs) {
		// no more selection
		it.idx++
		return DrawNormal, DrawNormal, DrawNormal
	}
	if elt := it.pathsIdxs[it.i]; elt.Idx == it.idx {
		// current path is in selection
		it.i++ // advance to in selected paths
		it.idx++
		switch {
		case it.mode == SelectionDrag || elt.Start && elt.End:
			// when dragging we want to fully gray out the scene path
			// regardless of wheter only start/end is selected
			return it.state, it.state, it.state
		case elt.Start:
			return it.state, DrawNormal, DrawNormal
		case elt.End:
			return DrawNormal, it.state, DrawNormal
		default:
			panic("selection.pathIdx contains an empty path (neiher start nor end)")
		}
	} else {
		// not selected, return normal state
		it.idx++
		return DrawNormal, DrawNormal, DrawNormal
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// buildingDrawStateIterator
////////////////////////////////////////////////////////////////////////////////////////////////////

// buildingDrawStateIterator is a helper to get the next scene building draw state.
type buildingDrawStateIterator struct {
	selectedIt MaskIterator
	// state to return for selected buildings/paths
	state DrawState
}

func (it *buildingDrawStateIterator) Next() DrawState {
	if it.selectedIt.Next() {
		// it.selected.i is the index of the next selected building/path in selection
		// we want the current
		return it.state
	} else {
		// not selected, return normal state
		return DrawNormal
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// MaskIterator
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
