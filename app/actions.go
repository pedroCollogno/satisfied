package app

import rl "github.com/gen2brain/raylib-go/raylib"

// ActionTarget - action target
type ActionTarget int

const (
	// TargetApp - app level actions
	TargetApp ActionTarget = iota
	// TargetGui - gui level actions
	TargetGui
	// TargetCamera - camera level actions
	TargetCamera
	// TargetSelector - selector level actions
	TargetSelector
	// TargetNewPath - new path level actions
	TargetNewPath
	// TargetNewBuilding - new building level actions
	TargetNewBuilding
	// TargetNewTextBox - new text box level actions
	TargetNewTextBox
	// TargetSelection - selection level actions
	TargetSelection
)

// Action is an abstraction layer between inputs and state updates in [Update] step.
//
// It allows to decouple inputs and state updates, cleanly defining atomic state updates and
// each function / method perimeter.
//
// See:
// - [ActionHandler]
// - [GetActionFunc]
// - [getAction]
// - [DispatchAction]
type Action interface {
	// Returns which app mode the action is for
	Target() ActionTarget
}

// GetActionFunc are responsible for converting inputs (mouse and keyboard) into an [Action].
//
// Each [AppMode] has a corresponding struct implementing a [GetActionFunc] method.
// They do not modify the state directly.
//
// In most cases, the [GetActionFunc] methods do not returns an [Action] directly, but rather
// calls appropriate [ActionHandler] directly and returns their result.
//
// This prevents runtime reflection (switch action.(type) statements) at the expense of a deeper
// call stack.
type GetActionFunc func() Action

// ActionHandler are responsible for performing and an [Action], and may return a follow up [Action].
//
// Each [AppMode] has a corresponding struct implementing a Do [ActionHandler] method.
//
// In most cases, the [ActionHandler] only calls appropriate handler for the specific action
// they recieve.
// Those handler,in turn, often do not return an [Action] directly, but rather
// calls other handler directly and returns their result.
//
// This prevents runtime reflection (switch action.(type) statements) at the expense of a deeper
// call stack.
type ActionHandler func(Action) Action

////////////////////////////////////////////////////////////////////////////////////////////////////
// [TargetApp] actions
////////////////////////////////////////////////////////////////////////////////////////////////////

// AppActionSwitchMode - switch app mode
//
// This action terminates the action chain in the current frame.
type AppActionSwitchMode struct {
	Mode   AppMode
	Resets Resets
}

// AppActionNew - create a new project
type AppActionNew struct{}

// AppActionSave - save the current project to a file
type AppActionSave struct{ Filepath string }

// AppActionSaveAs - save the current project to a new file
type AppActionSaveAs struct{}

// AppActionOpen - open and load a project from a file
type AppActionOpen struct{}

func (a AppActionSwitchMode) Target() ActionTarget { return TargetApp }
func (a AppActionSave) Target() ActionTarget       { return TargetApp }
func (a AppActionSaveAs) Target() ActionTarget     { return TargetApp }
func (a AppActionOpen) Target() ActionTarget       { return TargetApp }

////////////////////////////////////////////////////////////////////////////////////////////////////
// [TargetGui] actions
////////////////////////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////////////////////////
// [TargetCamera] actions
////////////////////////////////////////////////////////////////////////////////////////////////////

// CameraActionReset - reset the camera to its default position and zoom
type CameraActionReset struct{}

// CameraActionZoom - zoom the camera
type CameraActionZoom struct {
	By float32
	At rl.Vector2
}

// CameraActionPan - pan the camera by the given vector
type CameraActionPan struct{ By rl.Vector2 }

func (a CameraActionReset) Target() ActionTarget { return TargetCamera }
func (a CameraActionZoom) Target() ActionTarget  { return TargetCamera }
func (a CameraActionPan) Target() ActionTarget   { return TargetCamera }

////////////////////////////////////////////////////////////////////////////////////////////////////
// [TargetSelector] actions
////////////////////////////////////////////////////////////////////////////////////////////////////

// SelectorActionInit - initialize the rectangle selector
type SelectorActionInit struct{ Pos rl.Vector2 }

// SelectorActionMoveTo - update the rectangle selector end corner position
type SelectorActionMoveTo struct{ Pos rl.Vector2 }

// SelectorActionSelect - apply the rectangle selector
type SelectorActionSelect struct{}

func (a SelectorActionInit) Target() ActionTarget   { return TargetSelector }
func (a SelectorActionMoveTo) Target() ActionTarget { return TargetSelector }
func (a SelectorActionSelect) Target() ActionTarget { return TargetSelector }

////////////////////////////////////////////////////////////////////////////////////////////////////
// [TargetNewPath] actions
////////////////////////////////////////////////////////////////////////////////////////////////////

// NewPathActionInit - initialize placing a new path
type NewPathActionInit struct{ DefIdx int }

// NewPathActionMoveTo - update the new path position (either start or end, depending on internal state)
type NewPathActionMoveTo struct{ Pos rl.Vector2 }

// NewPathActionReverse - reverse the new path direction
type NewPathActionReverse struct{}

// NewPathActionPlaceStart - place a new path start
type NewPathActionPlaceStart struct{}

// NewPathActionPlace - add a new path to the scene
type NewPathActionPlace struct{}

func (a NewPathActionInit) Target() ActionTarget       { return TargetNewPath }
func (a NewPathActionMoveTo) Target() ActionTarget     { return TargetNewPath }
func (a NewPathActionReverse) Target() ActionTarget    { return TargetNewPath }
func (a NewPathActionPlaceStart) Target() ActionTarget { return TargetNewPath }
func (a NewPathActionPlace) Target() ActionTarget      { return TargetNewPath }

////////////////////////////////////////////////////////////////////////////////////////////////////
// [TargetNewBuilding] actions
////////////////////////////////////////////////////////////////////////////////////////////////////

// NewBuildingActionInit - initialize placing a new path
type NewBuildingActionInit struct{ DefIdx int }

// NewBuildingActionMoveTo - update the new building position
type NewBuildingActionMoveTo struct{ Pos rl.Vector2 }

// NewBuildingActionRotate - rotate the new building direction
type NewBuildingActionRotate struct{}

// NewBuildingActionPlace - place a new building
type NewBuildingActionPlace struct{}

func (a NewBuildingActionInit) Target() ActionTarget   { return TargetNewBuilding }
func (a NewBuildingActionMoveTo) Target() ActionTarget { return TargetNewBuilding }
func (a NewBuildingActionRotate) Target() ActionTarget { return TargetNewBuilding }
func (a NewBuildingActionPlace) Target() ActionTarget  { return TargetNewBuilding }

////////////////////////////////////////////////////////////////////////////////////////////////////
// [TargetNewTextBox] actions
////////////////////////////////////////////////////////////////////////////////////////////////////

// NewTextBoxActionInit - initialize placing a new text box
type NewTextBoxActionInit struct{ DefIdx int }

// NewTextBoxActionMoveTo - update the new text box position (either start or end, depending on internal state)
type NewTextBoxActionMoveTo struct{ Pos rl.Vector2 }

// NewTextBoxActionReverse - reverse the new text box direction
type NewTextBoxActionReverse struct{}

// NewTextBoxActionPlaceStart - place a new text box start
type NewTextBoxActionPlaceStart struct{}

// NewTextBoxActionPlace - add a new text box to the scene
type NewTextBoxActionPlace struct{}

func (a NewTextBoxActionInit) Target() ActionTarget       { return TargetNewTextBox }
func (a NewTextBoxActionMoveTo) Target() ActionTarget     { return TargetNewTextBox }
func (a NewTextBoxActionReverse) Target() ActionTarget    { return TargetNewTextBox }
func (a NewTextBoxActionPlaceStart) Target() ActionTarget { return TargetNewTextBox }
func (a NewTextBoxActionPlace) Target() ActionTarget      { return TargetNewTextBox }

////////////////////////////////////////////////////////////////////////////////////////////////////
// [TargetSelection] actions
////////////////////////////////////////////////////////////////////////////////////////////////////

// SelectionActionInitSingleDrag - initialize a new selection with a single object in [SelectionDrag] mode
type SelectionActionInitSingleDrag struct {
	// Object to select
	Object Object
	// Start position [SelectionDrag]
	Pos rl.Vector2
}

// SelectionActionInitSelection - initialize a new selection from a subset of the scene
type SelectionActionInitSelection struct{ Selection ObjectSelection }

// SelectionActionDelete - delete the current selection ([SelectionNormal])
type SelectionActionDelete struct{}

// SelectionActionBeginTransformation - switch selection to either [SelectionDrag] or [SelectionDuplicate]
type SelectionActionBeginTransformation struct {
	// Selection mode
	Mode SelectionMode
	// Start position for transformation
	Pos rl.Vector2
	// If true, the transformation will track the mouse even on mouse down
	MoveOnMouseDown bool
}

// SelectionActionMoveTo - update the selection transformation position
type SelectionActionMoveTo struct{ Pos rl.Vector2 }

// SelectionActionMoveBy - translate the selection by a given delta
type SelectionActionMoveBy struct{ Delta rl.Vector2 }

// SelectionActionRotate - rotate the selection transformation
type SelectionActionRotate struct{}

// SelectionActionEndTransformation - commit the selection transformation to the scene
type SelectionActionEndTransformation struct {
	// If true, discard the transformation regardless of its validity
	Discard bool
}

func (a SelectionActionInitSingleDrag) Target() ActionTarget      { return TargetSelection }
func (a SelectionActionInitSelection) Target() ActionTarget       { return TargetSelection }
func (a SelectionActionDelete) Target() ActionTarget              { return TargetSelection }
func (a SelectionActionBeginTransformation) Target() ActionTarget { return TargetSelection }
func (a SelectionActionMoveTo) Target() ActionTarget              { return TargetSelection }
func (a SelectionActionMoveBy) Target() ActionTarget              { return TargetSelection }
func (a SelectionActionRotate) Target() ActionTarget              { return TargetSelection }
func (a SelectionActionEndTransformation) Target() ActionTarget   { return TargetSelection }
