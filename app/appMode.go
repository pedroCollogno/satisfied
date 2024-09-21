package app

import "strings"

type AppMode int

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
		return "New Path"
	case ModeNewBuilding:
		return "New Building"
	case ModeSelection:
		return "Selection"
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

////////////////////////////////////////////////////////////////////////////////////////////////////
// Resets for [App.doSwitchMode]
////////////////////////////////////////////////////////////////////////////////////////////////////

// TODO: Make it a flag set
type Resets struct {
	Selector    bool
	NewPath     bool
	NewBuilding bool
	Selection   bool
	Gui         bool
	Camera      bool
}

// ResetAll resets all states but the camera.
func ResetAll() Resets {
	return Resets{
		Selector:    true,
		NewPath:     true,
		NewBuilding: true,
		Selection:   true,
		Gui:         true,
		Camera:      false,
	}
}

func (r Resets) String() string {
	ons := make([]string, 0, 6)
	if r.Selector {
		ons = append(ons, "Selector")
	}
	if r.NewPath {
		ons = append(ons, "NewPath")
	}
	if r.NewBuilding {
		ons = append(ons, "NewBuilding")
	}
	if r.Selection {
		ons = append(ons, "Selection")
	}
	if r.Gui {
		ons = append(ons, "Gui")
	}
	if r.Camera {
		ons = append(ons, "Camera")
	}
	return strings.Join(ons, " ")
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

func (r Resets) WithCamera(v bool) Resets {
	r.Camera = v
	return r
}
