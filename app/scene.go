package app

import (
	"bufio"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/bonoboris/satisfied/log"
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

	// History position the last time the scene was saved
	savedHistoryPos int

	// The scene object currently hovered by the mouse
	Hovered Object
	// Placed buildings
	Buildings []Building
	// Placed paths
	Paths []Path
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

type sceneOp interface {
	// Do the operation
	do(*Scene)
	// Redo undone operation
	redo(*Scene) Action
	// Undo done operation
	undo(*Scene) Action
	// trace state
	traceState()
	// name (for logging)
	name() string
}

type sceneOpAdd struct {
	paths     []Path
	buildings []Building
}

func (op sceneOpAdd) name() string { return "add" }

func (op sceneOpAdd) traceState() {
	if log.WillTrace() {
		for i, p := range op.paths {
			log.Trace("scene.operation.add.paths", "i", i, "value", p)
		}
		for i, b := range op.buildings {
			log.Trace("scene.operation.add.buildings", "i", i, "value", b)
		}
	}
}

func (op *sceneOpAdd) do(s *Scene) {
	s.Paths = append(s.Paths, op.paths...)
	s.Buildings = append(s.Buildings, op.buildings...)
}

func (op *sceneOpAdd) redo(s *Scene) Action {
	s.traceState("before", "add.redo")
	op.traceState()
	log.Debug("scene.operation.add", "action", "redo", "num_paths", len(op.paths), "num_buildings", len(op.buildings))
	op.do(s)
	ss := sceneSubset{
		buildingsIdxs: Range(len(s.Buildings)-len(op.buildings), len(s.Buildings)),
		pathsIdxs:     Range(len(s.Paths)-len(op.paths), len(s.Paths)),
	}
	ss.updateBounds()
	s.traceState("after", "add.redo")
	return selection.doInitSceneSubset(ss)
}

func (op *sceneOpAdd) undo(s *Scene) Action {
	s.traceState("before", "add.undo")
	op.traceState()
	log.Debug("scene.operation.add", "action", "redo", "num_paths", len(op.paths), "num_buildings", len(op.buildings))
	s.Paths = s.Paths[:len(s.Paths)-len(op.paths)]
	s.Buildings = s.Buildings[:len(s.Buildings)-len(op.buildings)]
	s.traceState("after", "add.undo")
	return app.doSwitchMode(ModeNormal, ResetAll())
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

func (op sceneOpDelete) name() string { return "delete" }

func (op sceneOpDelete) traceState() {
	if log.WillTrace() {
		for i, p := range op.paths {
			log.Trace("scene.operation.delete.paths", "i", i, "sceneIdx", op.pathIdxs[i], "value", p)
		}
		for i, b := range op.buildings {
			log.Trace("scene.operation.delete.buildings", "i", i, "sceneIdx", op.buildingIdxs[i], "value", b)
		}
	}
}

func (op *sceneOpDelete) do(s *Scene) {
	s.Paths = SwapDeleteMany(s.Paths, op.pathIdxs)
	s.Buildings = SwapDeleteMany(s.Buildings, op.buildingIdxs)
}

func (op *sceneOpDelete) redo(s *Scene) Action {
	s.traceState("before", "delete.redo")
	op.traceState()
	log.Debug("scene.operation.delete", "action", "redo", "pathIdxs", op.pathIdxs, "buildingIdxs", op.buildingIdxs)
	op.do(s)
	s.traceState("after", "delete.redo")
	return app.doSwitchMode(ModeNormal, ResetAll())
}

func (op *sceneOpDelete) undo(s *Scene) Action {
	s.traceState("before", "delete.undo")
	op.traceState()
	log.Debug("scene.operation.delete", "action", "undo", "pathIdxs", op.pathIdxs, "buildingIdxs", op.buildingIdxs)
	s.Paths = SwapInsertMany(s.Paths, op.pathIdxs, op.paths)
	s.Buildings = SwapInsertMany(s.Buildings, op.buildingIdxs, op.buildings)
	ss := sceneSubset{
		buildingsIdxs: op.buildingIdxs,
		pathsIdxs:     op.pathIdxs,
	}
	ss.updateBounds()
	s.traceState("after", "delete.undo")
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

func (op sceneOpModify) name() string { return "modify" }

func (op sceneOpModify) traceState() {
	if log.WillTrace() {
		for i, idx := range op.pathIdxs {
			log.Trace("scene.operation.modify.paths", "i", i, "sceneIdx", idx, "old", op.oldPaths[i], "new", op.newPaths[i])
		}
		for i, idx := range op.buildingsIdxs {
			log.Trace("scene.operation.modify.buildings", "i", i, "sceneIdx", idx, "old", op.oldBuildings[i], "new", op.newBuildings[i])
		}
	}
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
	s.traceState("before", "modify.redo")
	op.traceState()
	log.Debug("scene.operation.modify", "action", "redo", "pathIdxs", op.pathIdxs, "buildingIdxs", op.buildingsIdxs)
	op.do(s)
	ss := sceneSubset{
		buildingsIdxs: op.buildingsIdxs,
		pathsIdxs:     op.pathIdxs,
	}
	ss.updateBounds()
	s.traceState("after", "modify.redo")
	return selection.doInitSceneSubset(ss)
}

func (op *sceneOpModify) undo(s *Scene) Action {
	s.traceState("before", "modify.undo")
	op.traceState()
	log.Debug("scene.operation.modify", "action", "undo", "pathIdxs", op.pathIdxs, "buildingIdxs", op.buildingsIdxs)
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
	s.traceState("after", "modify.undo")
	return selection.doInitSceneSubset(ss)
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// Scene Modifiers methods
////////////////////////////////////////////////////////////////////////////////////////////////////

func (s *Scene) addSceneOp(op sceneOp) {
	s.traceState("before", op.name()+".do")
	op.traceState()
	log.Info("scene.operation", "op", op.name())
	s.history = s.history[:s.historyPos] // trim any undone operations
	op.do(s)                             // actually perform the operation
	s.history = append(s.history, op)    // append the operation to the history
	s.historyPos++                       // increment history position
	s.traceState("after", op.name()+".do")
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
	if scene.Hovered.Type == TypeBuilding && SortedIntsIndex(buildingIdxs, scene.Hovered.Idx) >= 0 ||
		scene.Hovered.Type == TypePath && SortedIntsIndex(pathIdxs, scene.Hovered.Idx) >= 0 {
		// hovered object was deleted
		scene.Hovered = scene.GetObjectAt(mouse.Pos)
	}
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
		log.Info("scene.operation", "undo", op.name())
		return true, op.undo(s)
	}
	log.Warn("cannot undo operation", "reason", "no more operations to undo")
	return false, nil
}

// Redo tries to redo the last undone operation, and returns whether it has, and the action to be performed.
func (s *Scene) Redo() (bool, Action) {
	if s.historyPos < len(s.history) {
		op := s.history[s.historyPos]
		log.Info("scene.operation", "undo", op.name())
		s.historyPos++ // increment history position
		return true, op.redo(s)
	}
	log.Warn("cannot redo operation", "reason", "no more operations to redo")
	return false, nil
}

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

func (s Scene) IsEmpty() bool {
	return len(s.Buildings) == 0 && len(s.Paths) == 0
}

// Equals returns true if the 2 scenes buildings and paths are equal (element-wise)
func (s Scene) Equals(other Scene) bool {
	return slices.Equal(s.Buildings, other.Buildings) && slices.Equal(s.Paths, other.Paths)
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
