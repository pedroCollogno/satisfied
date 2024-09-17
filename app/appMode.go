package app

type AppMode int

var appMode AppMode

const (
	// ModeNormal is the default mode
	ModeNormal AppMode = iota
	// ModeNewPath is used when creating a new path
	ModeNewPath
	// ModeNewBuilding is used when creating a new building
	ModeNewBuilding
	// ModeSelection is used when one or many object are selected
	ModeSelection
)

func (mode AppMode) String() string {
	switch mode {
	case ModeNormal:
		return "Normal"
	case ModeNewPath:
		return "NewPath"
	case ModeNewBuilding:
		return "NewBuilding"
	case ModeSelection:
		return "Selected"
	default:
		return "Invalid"
	}
}

// Assert panics if [AppMode] is not the expected one
func (m AppMode) Assert(mode AppMode) {
	if m != mode {
		panic("Invalid: expected " + m.String() + ", got " + mode.String())
	}
}

// doSwitchMode sets [AppMode] to the given mode and reset other modes state
//
// [ActionHandler] for [ActionSwitchMode]
//
// See: [ActionHandler]
func (m *AppMode) doSwitchMode(mode AppMode, resets Resets) Action {
	*m = mode
	if resets.Selector {
		selector.Reset()
	}
	if resets.NewPath {
		newPath.Reset()
	}
	if resets.NewBuilding {
		newBuilding.Reset()
	}
	if resets.Selection {
		selection.Reset()
	}
	if resets.Gui {
		gui.Reset()
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// Resets for [AppMode.doSwitchMode]
////////////////////////////////////////////////////////////////////////////////////////////////////

type Resets struct {
	Selector    bool
	NewPath     bool
	NewBuilding bool
	Selection   bool
	Gui         bool
}

func ResetAll() Resets {
	return Resets{
		Selector:    true,
		NewPath:     true,
		NewBuilding: true,
		Selection:   true,
		Gui:         true,
	}
}

func (r Resets) WithSelector(v bool) Resets {
	r.Selector = v
	return r
}

func (r Resets) WithNewPath(v bool) Resets {
	r.NewPath = v
	return r
}

func (r Resets) WithNewBuilding(v bool) Resets {
	r.NewBuilding = v
	return r
}

func (r Resets) WithSelection(v bool) Resets {
	r.Selection = v
	return r
}

func (r Resets) WithGui(v bool) Resets {
	r.Gui = v
	return r
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// Helpers
////////////////////////////////////////////////////////////////////////////////////////////////////

// Returns whether to pan the camera on arrow keys press
func (m AppMode) CameraArrowsKeyEnabled() bool {
	return m == ModeNormal || m == ModeSelection && selection.mode == SelectionNormal
}

// Returns whether to undo/redo the scene on Ctrl+Z/Y keys press
func (m AppMode) UndoRedoKeyEnabled() bool {
	return m == ModeNormal || m == ModeSelection && selection.mode == SelectionNormal
}
