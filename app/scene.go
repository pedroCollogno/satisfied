package app

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/bonoboris/satisfied/log"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// Scene holds the scene objects (buildings and paths)
var scene Scene

// Scene holds the scene objects (buildings and paths)
type Scene struct {
	ObjectCollection
	// History of scene operations (undo / redo)
	history []sceneOp
	// Current history position:
	//   - history[:historyPos] all have been done
	//   - history[historyPos:] all have been undone (if existing)
	historyPos int

	// History position the last time the scene was saved
	savedHistoryPos int

	// The scene object currently hovered by the mouse
	Hovered Object
	// was in modified state last frame
	wasModified bool
}

func (s Scene) traceState(key, val string) {
	if log.WillTrace() {
		if key != "" && val != "" {
			log.Trace("scene", key, val)
		}

		for i, b := range s.Buildings {
			log.Trace("scene.buildings", "i", i, "value", b)
		}
		for i, p := range s.Paths {
			log.Trace("scene.paths", "i", i, "value", p)
		}
		log.Trace("scene", "wasModified", s.wasModified, "historyPos", s.historyPos, "savedHistoryPos", s.savedHistoryPos)
		for i, op := range s.history {
			log.Trace("scene.history", "i", i, "op", op)
		}
		log.Trace("scene", "hovered", s.Hovered)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// sceneOp (undo / redo)
////////////////////////////////////////////////////////////////////////////////////////////////////

type sceneOpType string

const (
	SceneOpAdd    sceneOpType = "add"
	SceneOpDelete sceneOpType = "delete"
	SceneOpModify sceneOpType = "modify"
)

// sceneOp represents a scene operation
type sceneOp struct {
	// Type is the type of the operation
	Type sceneOpType
	// Sel is the selection the operation acts on (empty for [SceneOpAdd])
	Sel ObjectSelection
	// Old is the objects before the operation (empty for [SceneOpAdd])
	//
	// - in [SceneOpDelete] Old.Paths contains only the deleted paths ([ObjectSelection.FullPathIdxs])
	// - in [SceneOpModify] Old.Paths contains all the paths ([ObjectSelection.AnyPathIdxs])
	Old ObjectCollection
	// New is the objects after the operation (empty for [SceneOpDelete])
	New ObjectCollection
}

func (op sceneOp) traceState() {
	switch op.Type {
	case SceneOpAdd:
		log.Trace("scene.operation", "type", "add", "New", op.New)
	case SceneOpDelete:
		log.Trace("scene.operation", "type", "delete", "Sel", op.Sel, "Old", op.Old)
	case SceneOpModify:
		log.Trace("scene.operation", "type", "modify", "Sel", op.Sel, "Old", op.Old, "New", op.New)
	default:
		panic("invalid scene operation type")
	}
}

// do performs the operation
func (op sceneOp) do(s *Scene) {
	s.traceState("before", "sceneOp.do")
	op.traceState()
	log.Info("scene.operation", "do", string(op.Type))
	switch op.Type {

	case SceneOpAdd:
		log.Debug("scene.operation.add", "action", "do", "num_paths", len(op.New.Paths), "num_buildings", len(op.New.Buildings))
		s.Paths = append(s.Paths, op.New.Paths...)
		s.Buildings = append(s.Buildings, op.New.Buildings...)

	case SceneOpDelete:
		pathIdxs := op.Sel.FullPathIdxs()
		log.Debug("scene.operation.delete", "action", "do", "paths", pathIdxs, "buildinds", op.Sel.BuildingIdxs)
		s.Paths = SwapDeleteMany(s.Paths, op.Sel.FullPathIdxs())
		s.Buildings = SwapDeleteMany(s.Buildings, op.Sel.BuildingIdxs)

	case SceneOpModify:
		pathIdxs := op.Sel.AnyPathIdxs()
		log.Debug("scene.operation.modify", "action", "do", "paths", pathIdxs, "buildings", op.Sel.BuildingIdxs)
		for i, idx := range pathIdxs {
			s.Paths[idx] = op.New.Paths[i]
		}
		for i, idx := range op.Sel.BuildingIdxs {
			s.Buildings[idx] = op.New.Buildings[i]
		}

	default:
		panic("invalid scene operation type")
	}
	s.traceState("after", "sceneOp.do")
}

// redo performs the operation and returns the new selection, if any
func (op sceneOp) redo(s *Scene) ObjectSelection {
	s.traceState("before", "sceneOp.redo")
	op.traceState()
	log.Info("scene.operation", "redo", string(op.Type))

	var newSel ObjectSelection

	switch op.Type {

	case SceneOpAdd:
		log.Debug("scene.operation.add", "action", "redo", "num_paths", len(op.New.Paths), "num_buildings", len(op.New.Buildings))
		s.Paths = append(s.Paths, op.New.Paths...)
		s.Buildings = append(s.Buildings, op.New.Buildings...)

		newSel = ObjectSelection{
			BuildingIdxs: Range(len(s.Buildings)-len(op.New.Buildings), len(s.Buildings)),
		}
		n := len(s.Paths)
		for i := range n - len(op.New.Paths) {
			newSel.PathIdxs = append(newSel.PathIdxs, PathSel{Idx: n + i, Start: true, End: true})
		}
		newSel.recomputeBounds(s.ObjectCollection)

	case SceneOpDelete:
		pathIdxs := op.Sel.FullPathIdxs()
		log.Debug("scene.operation.delete", "action", "redo", "paths", pathIdxs, "buildinds", op.Sel.BuildingIdxs)
		s.Paths = SwapDeleteMany(s.Paths, pathIdxs)
		s.Buildings = SwapDeleteMany(s.Buildings, op.Sel.BuildingIdxs)

	case SceneOpModify:
		pathIdxs := op.Sel.AnyPathIdxs()
		log.Debug("scene.operation.modify", "action", "redo", "paths", pathIdxs, "buildings", op.Sel.BuildingIdxs)
		for i, idx := range pathIdxs {
			s.Paths[idx] = op.New.Paths[i]
		}
		for i, idx := range op.Sel.BuildingIdxs {
			s.Buildings[idx] = op.New.Buildings[i]
		}
		newSel = op.Sel
		newSel.recomputeBounds(s.ObjectCollection)

	default:
		panic("invalid scene operation type")
	}
	s.traceState("after", "sceneOp.redo")
	return newSel
}

// undo performs the operation and returns the new selection, if any
func (op sceneOp) undo(s *Scene) ObjectSelection {
	s.traceState("before", "sceneOp.undo")
	op.traceState()
	log.Info("scene.operation", "undo", string(op.Type))

	var newSel ObjectSelection

	switch op.Type {
	case SceneOpAdd:
		log.Debug("scene.operation.add", "action", "undo", "num_paths", len(op.New.Paths), "num_buildings", len(op.New.Buildings))
		s.Paths = s.Paths[:len(s.Paths)-len(op.Old.Paths)]
		s.Buildings = s.Buildings[:len(s.Buildings)-len(op.Old.Buildings)]

	case SceneOpDelete:
		pathIdxs := op.Sel.FullPathIdxs()
		log.Debug("scene.operation.delete", "action", "undo", "paths", pathIdxs, "buildinds", op.Sel.BuildingIdxs)
		s.Paths = SwapInsertMany(s.Paths, pathIdxs, op.Old.Paths)
		s.Buildings = SwapInsertMany(s.Buildings, op.Sel.BuildingIdxs, op.Old.Buildings)

		newSel = op.Sel

	case SceneOpModify:
		pathIdxs := op.Sel.AnyPathIdxs()
		log.Debug("scene.operation.modify", "action", "redo", "paths", pathIdxs, "buildings", op.Sel.BuildingIdxs)
		for i, idx := range pathIdxs {
			s.Paths[idx] = op.Old.Paths[i]
		}
		for i, idx := range op.Sel.BuildingIdxs {
			s.Buildings[idx] = op.Old.Buildings[i]
		}

		newSel = op.Sel
		newSel.recomputeBounds(s.ObjectCollection)

	default:
		panic("invalid scene operation type")
	}
	s.traceState("after", "sceneOp.undo")
	return newSel
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// Scene Modifiers methods
////////////////////////////////////////////////////////////////////////////////////////////////////

// doSceneOp adds the given operation to the scene history and performs it
func (s *Scene) doSceneOp(op sceneOp) {
	s.history = s.history[:s.historyPos] // trim any undone operations
	op.do(s)                             // actually perform the operation
	s.history = append(s.history, op)    // append the operation to the history
	s.historyPos++                       // increment history position
	s.Hovered = Object{}                 // invalidate hovered object just in case
}

// AddPath adds the given path to the scene.
//
// No validity check is performed.
func (s *Scene) AddPath(path Path) {
	s.doSceneOp(sceneOp{Type: SceneOpAdd, New: ObjectCollection{Paths: []Path{path}}})
}

// AddBuilding adds the given building to the scene.
//
// No validity check is performed.
func (s *Scene) AddBuilding(building Building) {
	s.doSceneOp(sceneOp{Type: SceneOpAdd, New: ObjectCollection{Buildings: []Building{building}}})
}

// AddObjects adds the given paths and buildings to the scene.
//
// No validity checks is performed.
func (s *Scene) AddObjects(col ObjectCollection) {
	s.doSceneOp(sceneOp{Type: SceneOpAdd, New: col.clone()})
}

// DeleteObjects deletes the given paths and buildings from the scene.
func (s *Scene) DeleteObjects(sel ObjectSelection) {
	sel = sel.clone()
	op := sceneOp{Type: SceneOpDelete, Sel: sel}
	op.Old.Buildings = CopyIdxs(op.Old.Buildings, s.Buildings, sel.BuildingIdxs)
	op.Old.Paths = CopyIdxs(op.Old.Paths, s.Paths, sel.FullPathIdxs())

	s.doSceneOp(op)
}

// ModifyObjects updates the given paths and buildings in the scene.
//
// No validity checks is performed.
func (s *Scene) ModifyObjects(sel ObjectSelection, new ObjectCollection) {
	sel = sel.clone()
	op := sceneOp{Type: SceneOpModify, Sel: sel, New: new.clone()}
	op.Old.Buildings = CopyIdxs(op.Old.Buildings, s.Buildings, sel.BuildingIdxs)
	op.Old.Paths = CopyIdxs(op.Old.Paths, s.Paths, sel.AnyPathIdxs())
	s.doSceneOp(op)
}

// Undo tries to undo the last operation, and returns whether it has, and the action to be performed.
func (s *Scene) Undo() (bool, Action) {
	if s.historyPos > 0 {
		s.historyPos-- // decrement history position
		op := s.history[s.historyPos]
		s.Hovered = Object{} // invalidate hovered object just in case
		// will switch to [ModeSelection] or [ModeNormal] if new selection is empty
		return true, selection.doInitSelection(op.undo(s))
	}
	log.Warn("cannot undo operation", "reason", "no more operations to undo")
	return false, nil
}

// Redo tries to redo the last undone operation, and returns whether it has, and the action to be performed.
func (s *Scene) Redo() (bool, Action) {
	if s.historyPos < len(s.history) {
		op := s.history[s.historyPos]
		s.historyPos++       // increment history position
		s.Hovered = Object{} // invalidate hovered object just in case
		// will switch to [ModeSelection] or [ModeNormal] if new selection is empty
		return true, selection.doInitSelection(op.redo(s))
	}
	log.Warn("cannot redo operation", "reason", "no more operations to redo")
	return false, nil
}

// HasUndo returns true if there are more undo operations to perform
func (s *Scene) HasUndo() bool { return s.historyPos > 0 }

// HasRedo returns true if there are more redo operations to perform
func (s *Scene) HasRedo() bool { return s.historyPos < len(s.history) }

// IsModified returns true if the scene has been modified since last save
func (s *Scene) IsModified() bool {
	return s.historyPos != s.savedHistoryPos
}

// ResetModified resets the scene modified flag
func (s *Scene) ResetModified() {
	s.traceState("before", "ResetModified")
	s.savedHistoryPos = s.historyPos
	log.Debug("scene.resetModified", "savedHistoryPos", s.savedHistoryPos)
	s.traceState("after", "ResetModified")
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// Scene other methods
////////////////////////////////////////////////////////////////////////////////////////////////////

// GetObjectAt returns the object at the given position (world coordinates)
//
// If multiple objects are at the position returns first one on this list:
//   - selected path with highest index: start / end over body
//   - selected building with highest index
//   - normal path with highest index: start / end over body
//   - normal building with highest index
//
// This is (mostly) the reverse of [Scene.Draw] order to make viewing/selecting masked objects easier.
//
// If no object is found, returns an zero-valued [Object]
func (s Scene) GetObjectAt(pos rl.Vector2) Object {
	for i := len(selection.PathIdxs) - 1; i >= 0; i-- {
		elt := selection.PathIdxs[i]
		p := s.Paths[elt.Idx]
		if elt.Start && p.CheckStartCollisionPoint(pos) {
			return Object{Type: TypePathStart, Idx: elt.Idx}
		}
		if elt.End && p.CheckEndCollisionPoint(pos) {
			return Object{Type: TypePathEnd, Idx: elt.Idx}
		}
		if p.CheckCollisionPoint(pos) {
			return Object{Type: TypePath, Idx: elt.Idx}
		}
	}

	for i := len(selection.BuildingIdxs) - 1; i >= 0; i-- {
		if s.Buildings[selection.BuildingIdxs[i]].Bounds().CheckCollisionPoint(pos) {
			return Object{Type: TypeBuilding, Idx: selection.BuildingIdxs[i]}
		}
	}

	// TODO: do not check selected paths / buildings again ?
	for i := len(s.Paths) - 1; i >= 0; i-- {
		p := s.Paths[i]
		if p.CheckStartCollisionPoint(pos) {
			return Object{Type: TypePathStart, Idx: i}
		}
		if p.CheckEndCollisionPoint(pos) {
			return Object{Type: TypePathEnd, Idx: i}
		}
		if p.CheckCollisionPoint(pos) {
			return Object{Type: TypePath, Idx: i}
		}
	}

	for i := len(s.Buildings) - 1; i >= 0; i-- {
		if s.Buildings[i].Bounds().CheckCollisionPoint(pos) {
			return Object{Type: TypeBuilding, Idx: i}
		}
	}
	return Object{}
}

// Update hovered object
func (s *Scene) Update() (action Action) {
	s.Hovered = s.GetObjectAt(mouse.Pos)

	if app.isNormal() && keyboard.Ctrl {
		switch keyboard.Pressed {
		case rl.KeyZ:
			_, action = s.Undo()
		case rl.KeyY:
			_, action = s.Redo()
		}
	}
	return action
}

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
	if app.Mode == ModeSelection {
		pathStates := selection.PathDrawStateIterator()
		for _, b := range s.Paths {
			start, end, body := pathStates.Next()
			b.DrawStart(start)
			b.DrawEnd(end)
			b.DrawBody(body)
		}
		buildingStates := selection.BuildingDrawStateIterator()
		for _, b := range s.Buildings {
			b.Draw(buildingStates.Next())
		}
	} else {
		for _, b := range s.Paths {
			b.Draw(DrawNormal)
		}
		for _, b := range s.Buildings {
			b.Draw(DrawNormal)
		}
	}

	if app.Mode == ModeSelection && selection.mode == SelectionNormal {
		// draw selection on top
		for _, idx := range selection.BuildingIdxs {
			s.Buildings[idx].Draw(DrawSelected)
		}

		for _, elt := range selection.PathIdxs {
			p := s.Paths[elt.Idx]
			if elt.Start && elt.End {
				p.DrawBody(DrawSelected)
			}
			if elt.Start {
				p.DrawStart(DrawSelected)
			}
			if elt.End {
				p.DrawEnd(DrawSelected)
			}
		}
	}

	// draw hovered object
	if !s.Hovered.IsEmpty() {
		if app.Mode == ModeNormal && !selector.selecting {
			s.Hovered.Draw(DrawNormal | DrawHovered)
		} else if app.Mode == ModeSelection && selection.mode == SelectionNormal {
			if selection.Contains(s.Hovered) {
				s.Hovered.Draw(DrawSelected | DrawHovered)
			} else {
				s.Hovered.Draw(DrawNormal | DrawHovered)
			}
		}
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// Save / Load
////////////////////////////////////////////////////////////////////////////////////////////////////

const (
	tagVersion = "#VERSION"
)

// SaveToText saves the scene into text format.
//
// All errors originate from the underlying [io.Writer].
func (s *Scene) SaveToText(w io.Writer) error {
	// // bufSize is kind of low estimation of actual size of the save
	// //   - version line is minimum 10 chars + '\n'
	// //   - the minimum building line is 7 chars + '\n'
	// //   - the minimum path line is 10 chars + '\n'
	// //
	// // Most of the actual lines will be longer as classes are more than 1 char long
	// // and numbers will have multiple digits.
	// bufSize := 10 * (len(s.Paths) + len(s.Buildings) + 1)
	// br := bufio.NewWriterSize(w, bufSize)
	br := bufio.NewWriter(w)
	defer br.Flush()
	// version
	_, err := br.WriteString(fmt.Sprintf("%s=%d\n", tagVersion, version))
	if err != nil {
		return err
	}
	// buildings
	for _, b := range s.Buildings {
		_, err := br.WriteString(fmt.Sprintf("%s %v %v %d\n", b.Def().Class, b.Pos.X, b.Pos.Y, b.Rot))
		if err != nil {
			return err
		}
	}
	// paths
	for _, p := range s.Paths {
		_, err := br.WriteString(fmt.Sprintf("%s %v %v %v %v\n", p.Def().Class, p.Start.X, p.Start.Y, p.End.X, p.End.Y))
		if err != nil {
			return err
		}
	}
	return nil
}

type DecodeTextError struct {
	Msg     string
	Err     error
	Line    int
	Version int
}

const (
	msgEmpty                = "empty file"
	msgInvalidVersionLine   = "invalid first line, expected '#VERSION=x'"
	msgInvalidVersionNumber = "invalid version, expected a positive integer"
	msgVersionTooHigh       = "version is too high"
	msgInvalidPath          = "invalid path line expected '[class] [startX] [startY] [endX] [endY]'"
	msgInvalidBuilding      = "invalid building line expected '[class] [posX] [posY] [rotation]'"
	msgInvalidClass         = "unknown class"
)

func (e DecodeTextError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("line %d: %s (%s)", e.Line, e.Msg, e.Err.Error())
	}
	return fmt.Sprintf("line %d: %s", e.Line, e.Msg)
}

func (s *Scene) LoadFromText(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	scanner.Scan()
	line := scanner.Text()
	if err := scanner.Err(); err != nil {
		return err
	}
	if len(line) == 0 {
		return DecodeTextError{Msg: msgEmpty}
	}
	// parse version
	var ver int
	if _, err := fmt.Sscanf(string(line), tagVersion+"=%d", &ver); err != nil {
		return DecodeTextError{Msg: msgInvalidVersionLine, Line: 1, Err: err}
	}
	if ver < 0 {
		return DecodeTextError{Msg: msgInvalidVersionNumber, Line: 1}
	}
	// call version specific function
	switch ver {
	case 0:
		return s.decodeText(scanner, ver)
	default:
		return DecodeTextError{Msg: msgVersionTooHigh, Version: ver, Line: 1}
	}
}

func (s *Scene) decodeText(scanner *bufio.Scanner, ver int) error {
	no := 2
	var (
		p Path
		b Building
	)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		class, fields, _ := strings.Cut(line, " ")
		if defIdx := pathDefs.Index(string(class)); defIdx >= 0 {
			p.DefIdx = defIdx
			if _, err := fmt.Sscanf(fields, "%f %f %f %f", &p.Start.X, &p.Start.Y, &p.End.X, &p.End.Y); err != nil {
				return DecodeTextError{Msg: msgInvalidPath, Line: no, Err: err, Version: ver}
			}
			s.Paths = append(s.Paths, p)
		} else if defIdx := buildingDefs.Index(string(class)); defIdx >= 0 {
			b.DefIdx = defIdx
			if _, err := fmt.Sscanf(fields, "%f %f %d", &b.Pos.X, &b.Pos.Y, &b.Rot); err != nil {
				return DecodeTextError{Msg: msgInvalidBuilding, Line: no, Err: err, Version: ver}
			}
			s.Buildings = append(s.Buildings, b)
		} else {
			return DecodeTextError{Msg: msgInvalidClass, Line: no, Version: ver}
		}
		no++
	}

	return nil
}
