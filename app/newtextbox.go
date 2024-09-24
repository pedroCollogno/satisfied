package app

import (
	"fmt"

	"github.com/bonoboris/satisfied/log"
	rl "github.com/gen2brain/raylib-go/raylib"
)

var newTextBox NewTextBox

type NewTextBox struct {
	TextBox           TextBox
	firstCornerPlaced bool
}

func (ntb NewTextBox) traceState(key, val string) {
	if key != "" && val != "" {
		log.Trace("newTextBox", key, val, "textBox", ntb.TextBox)
	} else {
		log.Trace("newTextBox", "textBox", ntb.TextBox)
	}
}

// Reset resets the [NewTextBox] state
func (ntb *NewTextBox) Reset() {
	ntb.traceState("before", "Reset")
	log.Debug("newTextBox.reset")
	ntb.TextBox = TextBox{Bounds: rl.NewRectangle(0, 0, 0, 0), Content: ""}
	ntb.firstCornerPlaced = false
	ntb.traceState("after", "Reset")
}

// GetAction processes inputs in [ModeNewTextBox], and returns an action to be performed.
//
// See: [GetActionFunc]
func (ntb *NewTextBox) GetAction() (action Action) {
	app.Mode.Assert(ModeNewTextBox)

	switch keyboard.Binding() {
	case BindingEscape:
		if ntb.firstCornerPlaced {
			return ntb.doInit()
		} else {
			return app.doSwitchMode(ModeNormal, ResetAll())
		}
	}
	if !mouse.InScene {
		return nil
	}

	if mouse.Left.Released {
		if !ntb.firstCornerPlaced {
			return ntb.doPlaceStart()
		} else {
			return ntb.doPlace()
		}
	}
	if !mouse.Left.Down {
		return ntb.doMoveTo(mouse.SnappedPos)
	}
	return nil
}

func (ntb *NewTextBox) doInit() Action {
	ntb.traceState("before", "doInit")
	log.Debug("newTextBox.doInit")
	ntb.TextBox = TextBox{
		Bounds:  rl.NewRectangle(0, 0, textBoxMinSize, textBoxMinSize),
		Content: textBoxDefaultText,
	}
	ntb.firstCornerPlaced = false
	ntb.traceState("after", "doInit")
	return app.doSwitchMode(ModeNewTextBox, ResetAll().WithNewTextBox(false))
}

func (ntb *NewTextBox) doMoveTo(pos rl.Vector2) Action {
	ntb.traceState("before", "doMoveTo")
	log.Trace("newTextBox.doMoveTo", "pos", pos) // moving by mouse -> tracing
	app.Mode.Assert(ModeNewTextBox)
	ntb.TextBox.Bounds = rl.NewRectangleCorners(ntb.TextBox.Bounds.TopLeft(), pos)
	ntb.traceState("after", "doMoveTo")
	return nil
}

func (ntb *NewTextBox) doPlaceStart() Action {
	ntb.traceState("before", "doPlaceStart")
	log.Debug("newTextBox.doPlaceStart")
	ntb.firstCornerPlaced = true
	ntb.traceState("after", "doPlaceStart")
	return nil
}

func (ntb *NewTextBox) doPlace() Action {
	ntb.traceState("before", "doPlace")
	log.Debug("newTextBox.doPlace")
	app.Mode.Assert(ModeNewTextBox)
	assert(ntb.firstCornerPlaced, "text box start not placed")
	if ntb.TextBox.Bounds.Width < textBoxMinSize {
		ntb.TextBox.Bounds.Width = textBoxMinSize
	}
	if ntb.TextBox.Bounds.Height < textBoxMinSize {
		ntb.TextBox.Bounds.Height = textBoxMinSize
	}
	scene.AddTextBox(ntb.TextBox)
	idx := len(scene.TextBoxes) - 1
	return selection.doInitSelection(ObjectSelection{TextBoxIdxs: []int{idx}})
}

// Dispatch performs an [NewTextBox] action, updating its state, and returns an new action to be performed
//
// See: [ActionHandler]
func (np *NewTextBox) Dispatch(action Action) Action {
	switch action := action.(type) {
	case NewTextBoxActionInit:
		return np.doInit()
	case NewTextBoxActionMoveTo:
		return np.doMoveTo(action.Pos)
	case NewTextBoxActionPlaceStart:
		return np.doPlaceStart()
	case NewTextBoxActionPlace:
		return np.doPlace()

	default:
		panic(fmt.Sprintf("NewTextBox.Dispatch: cannot handle: %T", action))
	}
}

func (np NewTextBox) Draw() {
	np.TextBox.Draw(DrawNew, false)
}
