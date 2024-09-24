// keyboard - keyboard input state

package app

import (
	"fmt"

	"github.com/bonoboris/satisfied/log"
	rl "github.com/gen2brain/raylib-go/raylib"
)

var keyboard = Keyboard{}

type Keyboard struct {
	// Pressed key (handles repeated key presses)
	//
	// It is not 0 (KeyNull) on the first frame the key is down and if a key is maintained down
	// and no other key is pressed, it will periodically repeat that key.
	Pressed int32

	// True if left or right shift key is down
	Shift bool

	// True if left or right control key is down
	Ctrl bool

	// True if alt key is down
	Alt bool

	// Last pressed key (to check for key repeat)
	down int32
}

func (kb *Keyboard) traceState() {
	log.Trace("keyboard", "pressed", GetKeyName(kb.Pressed), "shift", kb.Shift, "ctrl", kb.Ctrl, "alt", kb.Alt, "down", GetKeyName(kb.down))
}

// Update keyboard state
//
// Do not depends on any other state
func (kb *Keyboard) Update() {
	kb.Shift = rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift)
	kb.Ctrl = rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl)
	kb.Alt = rl.IsKeyDown(rl.KeyLeftAlt) || rl.IsKeyDown(rl.KeyRightAlt)

	// reset pressed key
	kb.Pressed = rl.KeyNull

	// not KeyNull only on the first frame a key is pressed
	key := rl.GetKeyPressed()

	if key != rl.KeyNull {
		// key is pressed, set Pressed and down
		log.Debug("keyboard.pressed", "key", GetKeyName(key), "ctrl", kb.Ctrl, "alt", kb.Alt, "shift", kb.Shift)
		kb.Pressed = key
		kb.down = key
		kb.traceState()
	} else if kb.down != rl.KeyNull {
		// no new key pressed and a key was down, check if it's still down
		if rl.IsKeyDown(kb.down) {
			// key is still down, check if it's a repeat
			if rl.IsKeyPressedRepeat(kb.down) {
				switch kb.down {
				case rl.KeyLeftControl, rl.KeyRightControl, rl.KeyLeftShift, rl.KeyRightShift, rl.KeyLeftAlt, rl.KeyRightAlt:
					// don't log repeats of modifiers
				default:
					log.Debug("keyboard.repeat", "key", GetKeyName(kb.down), "ctrl", kb.Ctrl, "alt", kb.Alt, "shift", kb.Shift)
				}
				// key is a repeat, set Pressed
				kb.Pressed = kb.down
				kb.traceState()
			}
		} else {
			log.Debug("key released", "key", GetKeyName(kb.down))
			// key is up, reset down, leave Pressed to KeyNull
			kb.down = rl.KeyNull
			kb.traceState()
		}
	}
}

// Binding returns the key binding that matches the pressed key and modifiers
//
// If gui is capturing key presses, always returns [BindingNull]
func (kb Keyboard) Binding() KeyBinding {
	if gui.CapturesKeyPress() {
		return BindingNull
	}
	for i, pair := range keyBindings {
		if pair[0].Matches(kb.Pressed, kb.Ctrl, kb.Alt, kb.Shift) || pair[1].Matches(kb.Pressed, kb.Ctrl, kb.Alt, kb.Shift) {
			return KeyBinding(i)
		}
	}
	return BindingNull
}

type optBool uint8

const (
	Any optBool = iota
	No          // false
	Yes         // true
)

// keyBindingDef associates a key pressed (and modifiers) to a [keyBinding]
type keyBindingDef struct {
	code             int32
	ctrl, alt, shift optBool
}

func (kbd keyBindingDef) Matches(key int32, ctrl, alt, shift bool) bool {
	return kbd.code != rl.KeyNull && kbd.code == key &&
		(kbd.ctrl == Any || ctrl && kbd.ctrl == Yes || !ctrl && kbd.ctrl == No) &&
		(kbd.alt == Any || alt && kbd.alt == Yes || !alt && kbd.alt == No) &&
		(kbd.shift == Any || shift && kbd.shift == Yes || !shift && kbd.shift == No)
}

// KeyBinding represents a keyboard key binding
type KeyBinding int

const (
	BindingNull KeyBinding = iota
	BindingEscape
	BindingSave
	BindingSaveAs
	BindingUndo
	BindingRedo
	BindingDelete
	BindingDuplicate
	BindingRotate
	BindingDrag
	BindingUp
	BindingDown
	BindingLeft
	BindingRight
	BindingZoomIn
	BindingZoomOut
	BindingZoomReset
)

// default key bindings
var keyBindings = [...][2]keyBindingDef{
	// defines as an array for performance and we are using the index syntax for readability and correctness
	// this is not a map
	BindingEscape:    {{code: rl.KeyEscape}},
	BindingDelete:    {{code: rl.KeyDelete}, {code: rl.KeyX}},
	BindingSave:      {{code: rl.KeyS, ctrl: Yes, shift: No}},
	BindingSaveAs:    {{code: rl.KeyS, ctrl: Yes, shift: Yes}},
	BindingUndo:      {{code: rl.KeyZ, ctrl: Yes, shift: No}},
	BindingRedo:      {{code: rl.KeyY, ctrl: Yes}, {code: rl.KeyZ, ctrl: Yes, shift: Yes}},
	BindingDuplicate: {{code: rl.KeyD}},
	BindingRotate:    {{code: rl.KeyR}},
	BindingDrag:      {{code: rl.KeyV}},
	BindingUp:        {{code: rl.KeyUp}},
	BindingDown:      {{code: rl.KeyDown}},
	BindingLeft:      {{code: rl.KeyLeft}},
	BindingRight:     {{code: rl.KeyRight}},
	BindingZoomIn:    {{code: rl.KeyEqual, shift: Yes}, {code: rl.KeyKpAdd}},
	BindingZoomOut:   {{code: rl.KeyMinus}, {code: rl.KeyKpSubtract}},
	BindingZoomReset: {{code: rl.KeyEqual, shift: No}, {code: rl.KeyKp0}},
}

func GetKeyName(key int32) string {
	switch key {
	case rl.KeyLeftControl, rl.KeyRightControl:
		return "<ctrl>"
	case rl.KeyLeftShift, rl.KeyRightShift:
		return "<shift>"
	case rl.KeyLeftAlt, rl.KeyRightAlt:
		return "<alt>"
	case rl.KeyBackspace:
		return "<backspace>"
	case rl.KeyEnter:
		return "<enter>"
	case rl.KeyEscape:
		return "<escape>"
	case rl.KeySpace:
		return "<space>"
	case rl.KeyTab:
		return "<tab>"
	case rl.KeyLeft:
		return "<left>"
	case rl.KeyRight:
		return "<right>"
	case rl.KeyUp:
		return "<up>"
	case rl.KeyDown:
		return "<down>"
	case rl.KeyInsert:
		return "<insert>"
	case rl.KeyDelete:
		return "<delete>"
	case rl.KeyA:
		return "a"
	case rl.KeyB:
		return "b"
	case rl.KeyC:
		return "c"
	case rl.KeyD:
		return "d"
	case rl.KeyE:
		return "e"
	case rl.KeyF:
		return "f"
	case rl.KeyG:
		return "g"
	case rl.KeyH:
		return "h"
	case rl.KeyI:
		return "i"
	case rl.KeyJ:
		return "j"
	case rl.KeyK:
		return "k"
	case rl.KeyL:
		return "l"
	case rl.KeyM:
		return "m"
	case rl.KeyN:
		return "n"
	case rl.KeyO:
		return "o"
	case rl.KeyP:
		return "p"
	case rl.KeyQ:
		return "q"
	case rl.KeyR:
		return "r"
	case rl.KeyS:
		return "s"
	case rl.KeyT:
		return "t"
	case rl.KeyU:
		return "u"
	case rl.KeyV:
		return "v"
	case rl.KeyW:
		return "w"
	case rl.KeyX:
		return "x"
	case rl.KeyY:
		return "y"
	case rl.KeyZ:
		return "z"
	case rl.KeyZero, rl.KeyKp0:
		return "0"
	case rl.KeyOne, rl.KeyKp1:
		return "1"
	case rl.KeyTwo, rl.KeyKp2:
		return "2"
	case rl.KeyThree, rl.KeyKp3:
		return "3"
	case rl.KeyFour, rl.KeyKp4:
		return "4"
	case rl.KeyFive, rl.KeyKp5:
		return "5"
	case rl.KeySix, rl.KeyKp6:
		return "6"
	case rl.KeySeven, rl.KeyKp7:
		return "7"
	case rl.KeyEight, rl.KeyKp8:
		return "8"
	case rl.KeyNine, rl.KeyKp9:
		return "9"
	case rl.KeyApostrophe:
		return "'"
	case rl.KeyComma:
		return ","
	case rl.KeyMinus:
		return "-"
	case rl.KeyPeriod:
		return "."
	case rl.KeySlash:
		return "/"
	case rl.KeySemicolon:
		return ";"
	case rl.KeyEqual:
		return "="
	case rl.KeyLeftBracket:
		return "["
	case rl.KeyRightBracket:
		return "]"
	case rl.KeyBackSlash:
		return "\\"
	case rl.KeyKpDecimal:
		return "."
	case rl.KeyKpDivide:
		return "/"
	case rl.KeyKpMultiply:
		return "*"
	case rl.KeyKpSubtract:
		return "-"
	case rl.KeyKpAdd:
		return "+"
	case rl.KeyKpEqual:
		return "="
	default:
		return fmt.Sprintf("\\x%x", key)
	}
}
