// objects - Contains struct and functions to represent object(s) and objects selection

package app

import (
	"fmt"
	"slices"

	"github.com/bonoboris/satisfied/math32"
	rl "github.com/gen2brain/raylib-go/raylib"
)

////////////////////////////////////////////////////////////////////////////////////////////////////
// Object
////////////////////////////////////////////////////////////////////////////////////////////////////

// Object represents a building, a whole path, a path start or a path end in the scene
type Object struct {
	// Type of the object
	Type ObjectType
	// Index in either [Scene.Buildings], [Scene.Paths] or [Scene.TextBoxes]
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
	case TypePathStart:
		scene.Paths[o.Idx].DrawStart(state)
	case TypePathEnd:
		scene.Paths[o.Idx].DrawEnd(state)
	case TypeTextBox:
		scene.TextBoxes[o.Idx].Draw(state, false)
	}
}

// ObjectType enumerates the different types of objects
type ObjectType int

const (
	TypeInvalid ObjectType = iota
	TypeBuilding
	TypePath
	TypePathStart
	TypePathEnd
	TypeTextBox
)

func (ot ObjectType) String() string {
	switch ot {
	case TypeBuilding:
		return "TypeBuilding"
	case TypePath:
		return "TypePath"
	case TypePathStart:
		return "TypePathStart"
	case TypePathEnd:
		return "TypePathEnd"
	case TypeTextBox:
		return "TypeTextBox"
	default:
		return "TypeInvalid"
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// ObjectCollection
////////////////////////////////////////////////////////////////////////////////////////////////////

// ObjectCollection represents a collection of objects (buildings, & paths)
type ObjectCollection struct {
	Buildings []Building
	Paths     []Path
	TextBoxes []TextBox
}

// IsEmpty returns true if the collection is empty
func (oc ObjectCollection) IsEmpty() bool {
	return len(oc.Buildings) == 0 && len(oc.Paths) == 0 && len(oc.TextBoxes) == 0
}

func (oc ObjectCollection) clone() ObjectCollection {
	return ObjectCollection{
		Buildings: slices.Clone(oc.Buildings),
		Paths:     slices.Clone(oc.Paths),
		TextBoxes: slices.Clone(oc.TextBoxes),
	}
}

// SelectFromRect fills sel with the objects in the given rectangle and recomputes its bounding box
//
// sel must be empty, it is passed to avoid reallocating it
func (oc ObjectCollection) SelectFromRect(sel *ObjectSelection, rect rl.Rectangle) {
	xmin, ymin := math32.MaxFloat32, math32.MaxFloat32
	xmax, ymax := -math32.MaxFloat32, -math32.MaxFloat32
	for i, b := range scene.Buildings {
		bounds := b.Bounds()
		tl := bounds.TopLeft()
		br := bounds.BottomRight()
		if rect.CheckCollisionPoint(tl) && rect.CheckCollisionPoint(br) {
			sel.BuildingIdxs = append(sel.BuildingIdxs, i)
			xmin, ymin = min(xmin, tl.X), min(ymin, tl.Y)
			xmax, ymax = max(xmax, br.X), max(ymax, br.Y)
		}
	}
	for i, p := range scene.Paths {
		start := rect.CheckCollisionPoint(p.Start)
		end := rect.CheckCollisionPoint(p.End)
		if start && end {
			sel.PathIdxs = append(sel.PathIdxs, PathSel{Idx: i, Start: true, End: true})
			xmin, ymin = min(xmin, min(p.Start.X, p.End.X)), min(ymin, min(p.Start.Y, p.End.Y))
			xmax, ymax = max(xmax, max(p.Start.X, p.End.X)), max(ymax, max(p.Start.Y, p.End.Y))
		} else if start {
			sel.PathIdxs = append(sel.PathIdxs, PathSel{Idx: i, Start: true})
			xmin, ymin = min(xmin, p.Start.X), min(ymin, p.Start.Y)
			xmax, ymax = max(xmax, p.Start.X), max(ymax, p.Start.Y)
		} else if end {
			sel.PathIdxs = append(sel.PathIdxs, PathSel{Idx: i, End: true})
			xmin, ymin = min(xmin, p.End.X), min(ymin, p.End.Y)
			xmax, ymax = max(xmax, p.End.X), max(ymax, p.End.Y)
		}
	}

	for i, tb := range oc.TextBoxes {
		tl := tb.Bounds.TopLeft()
		br := tb.Bounds.BottomRight()
		if rect.CheckCollisionPoint(tl) && rect.CheckCollisionPoint(br) {
			sel.TextBoxIdxs = append(sel.TextBoxIdxs, i)
			xmin, ymin = min(xmin, tl.X), min(ymin, tl.Y)
			xmax, ymax = max(xmax, br.X), max(ymax, br.Y)
		}
	}

	if !sel.IsEmpty() {
		sel.Bounds = rl.NewRectangle(xmin, ymin, xmax-xmin, ymax-ymin)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// ObjectSelection
////////////////////////////////////////////////////////////////////////////////////////////////////

// ObjectSelection represents a selection of objects in a [ObjectCollection]
type ObjectSelection struct {
	// BuildingIdxs is the slice of selected buildings indices of a [ObjectCollection.Buildings]
	//
	// It must be sorted in ascending order.
	BuildingIdxs []int
	// PathIdxs is the slice of selected paths indices of a [ObjectCollection.Paths] and whether
	// the path start and/or end is selected
	//
	// It must be sorted in ascending order.
	PathIdxs []PathSel
	// TextBoxIdxs is the slice of selected text boxes indices of a [ObjectCollection.TextBoxes]
	//
	// It must be sorted in ascending order.
	TextBoxIdxs []int
	// Bounds is the bounding box of the selection
	Bounds rl.Rectangle
}

// IsEmpty returns true if the selection is empty
func (os ObjectSelection) IsEmpty() bool {
	return len(os.BuildingIdxs) == 0 && len(os.PathIdxs) == 0 && len(os.TextBoxIdxs) == 0
}

// Creates a clone of the selection
func (os ObjectSelection) clone() ObjectSelection {
	return ObjectSelection{
		BuildingIdxs: slices.Clone(os.BuildingIdxs),
		PathIdxs:     slices.Clone(os.PathIdxs),
		TextBoxIdxs:  slices.Clone(os.TextBoxIdxs),
		Bounds:       os.Bounds,
	}
}

// copy the selection into the given [ObjectSelection] (clears the target)
func (os ObjectSelection) copy(into *ObjectSelection) {
	into.BuildingIdxs = append(into.BuildingIdxs[:0], os.BuildingIdxs...)
	into.PathIdxs = append(into.PathIdxs[:0], os.PathIdxs...)
	into.TextBoxIdxs = append(into.TextBoxIdxs[:0], os.TextBoxIdxs...)
	into.Bounds = os.Bounds
}

// reset the selection
func (os *ObjectSelection) reset() {
	os.BuildingIdxs = os.BuildingIdxs[:0]
	os.PathIdxs = os.PathIdxs[:0]
	os.TextBoxIdxs = os.TextBoxIdxs[:0]
	os.Bounds = rl.NewRectangle(0, 0, 0, 0)
}

// FullPathIdxs returns the indices of the path with both start and end selected
func (os ObjectSelection) FullPathIdxs() []int {
	idxs := make([]int, 0, len(os.PathIdxs))
	for _, pSel := range os.PathIdxs {
		if pSel.Start && pSel.End {
			idxs = append(idxs, pSel.Idx)
		}
	}
	return idxs
}

// AnyPathIdxs returns the indices of the path with either start or end selected
func (os ObjectSelection) AnyPathIdxs() []int {
	idxs := make([]int, 0, len(os.PathIdxs))
	for _, pSel := range os.PathIdxs {
		idxs = append(idxs, pSel.Idx)
	}
	return idxs
}

// StartIdxs returns the indices of the path with start selected
func (os ObjectSelection) StartIdxs() []int {
	idxs := make([]int, 0, len(os.PathIdxs))
	for _, pSel := range os.PathIdxs {
		if pSel.Start {
			idxs = append(idxs, pSel.Idx)
		}
	}
	return idxs
}

// EndIdxs returns the indices of the path with end selected
func (os ObjectSelection) EndIdxs() []int {
	idxs := make([]int, 0, len(os.PathIdxs))
	for _, pSel := range os.PathIdxs {
		if pSel.End {
			idxs = append(idxs, pSel.Idx)
		}
	}
	return idxs
}

// recomputeBounds recomputes the selection bounding box
func (os *ObjectSelection) recomputeBounds(oc ObjectCollection) {
	xmin := math32.MaxFloat32
	ymin := math32.MaxFloat32
	xmax := -math32.MaxFloat32
	ymax := -math32.MaxFloat32
	for _, idx := range os.BuildingIdxs {
		bounds := oc.Buildings[idx].Bounds()
		xmin, xmax = min(xmin, bounds.X), max(xmax, bounds.X+bounds.Width)
		ymin, ymax = min(ymin, bounds.Y), max(ymax, bounds.Y+bounds.Height)
	}
	for _, pSel := range os.PathIdxs {
		p := oc.Paths[pSel.Idx]
		if pSel.Start {
			xmin, xmax = min(xmin, p.Start.X), max(xmax, p.Start.X)
			ymin, ymax = min(ymin, p.Start.Y), max(ymax, p.Start.Y)
		}
		if pSel.End {
			xmin, xmax = min(xmin, p.End.X), max(xmax, p.End.X)
			ymin, ymax = min(ymin, p.End.Y), max(ymax, p.End.Y)
		}
	}
	for _, idx := range os.TextBoxIdxs {
		bounds := oc.TextBoxes[idx].Bounds
		xmin, xmax = min(xmin, bounds.X), max(xmax, bounds.X+bounds.Width)
		ymin, ymax = min(ymin, bounds.Y), max(ymax, bounds.Y+bounds.Height)
	}
	os.Bounds = rl.NewRectangle(xmin, ymin, xmax-xmin, ymax-ymin)
}

// Contains returns true if the given object is in the selection (false for [TypeInvalid])
func (os ObjectSelection) Contains(obj Object) bool {
	switch obj.Type {
	case TypeBuilding:
		return SortedIntsIndex(os.BuildingIdxs, obj.Idx) >= 0
	case TypePath:
		return SortedIntsIndex(os.FullPathIdxs(), obj.Idx) >= 0
	case TypePathStart:
		return SortedIntsIndex(os.StartIdxs(), obj.Idx) >= 0
	case TypePathEnd:
		return SortedIntsIndex(os.EndIdxs(), obj.Idx) >= 0
	case TypeTextBox:
		return SortedIntsIndex(os.TextBoxIdxs, obj.Idx) >= 0
	default:
		return false
	}
}

// BuildingsIterator returns a mask iterator of the selected buildings
func (os ObjectSelection) BuildingsIterator() MaskIterator { return NewMaskIterator(os.BuildingIdxs) }

// PathsIterator returns a mask iterator of the selected paths
func (os ObjectSelection) PathsIterator() PathSelMaskIterator {
	return NewPathSelMaskIterator(os.PathIdxs)
}

// TextBoxesIterator returns a mask iterator of the selected text boxes
func (os ObjectSelection) TextBoxesIterator() MaskIterator { return NewMaskIterator(os.TextBoxIdxs) }

// PathSel represents a selected path in a [ObjectSelection]
//
// One of [Start] or [End] must be true.
type PathSel struct {
	// Idx is the index of the path in [ObjectCollection.Paths]
	Idx int
	// Start is true if the path start is selected
	Start bool
	// End is true if the path end is selected
	End bool
}

func (p PathSel) String() string {
	switch {
	case p.Start && p.End:
		return fmt.Sprintf("{%d start end}", p.Idx)
	case p.Start:
		return fmt.Sprintf("{%d start}", p.Idx)
	case p.End:
		return fmt.Sprintf("{%d end}", p.Idx)
	default:
		return fmt.Sprintf("{%d}", p.Idx)
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

// Iterate over a would-be mask from the true values.
type PathSelMaskIterator struct {
	// index for which to return true
	TrueIdxs []PathSel
	// current index counter
	Idx int
	// index of the next true value in [PathSelMaskIterator.TrueIdxs]
	i int
}

// NewPathSelMaskIterator returns a new [PathSelMaskIterator] from the given true values.
//
// trueIdxs must be sorted in ascending order.
func NewPathSelMaskIterator(trueIdx []PathSel) PathSelMaskIterator {
	return PathSelMaskIterator{TrueIdxs: trueIdx, Idx: 0, i: 0}
}

// Next returns the next pair of (start, end) from the mask.
//
// It always returns (false, flase) when all the true values have been iterated over
// [PathSelMaskIterator.Idx] counter.
func (it *PathSelMaskIterator) Next() (bool, bool) {
	if it.i == len(it.TrueIdxs) {
		// no more selection
		it.Idx++
		return false, false
	}
	if elt := it.TrueIdxs[it.i]; it.Idx == elt.Idx {
		it.i++ // advance to next true value
		it.Idx++
		return elt.Start, elt.End
	} else {
		it.Idx++
		return false, false
	}
}
