// keyboard - keyboard input state

package app

import rl "github.com/gen2brain/raylib-go/raylib"

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
		kb.Pressed = key
		kb.down = key
	} else if kb.down != rl.KeyNull {
		// no new key pressed and a key was down, check if it's still down
		if rl.IsKeyDown(kb.down) {
			// key is still down, check if it's a repeat
			if rl.IsKeyPressedRepeat(kb.down) {
				// key is a repeat, set Pressed
				kb.Pressed = kb.down
			}
		} else {
			// key is up, reset down, leave Pressed to KeyNull
			kb.down = rl.KeyNull
		}
	}
}
