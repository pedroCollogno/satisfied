// animations - Animations timers
package app

import (
	"github.com/bonoboris/satisfied/math32"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// Animations state
var animations Animations

type Animations struct {
	// Elapsed time since window initialization (seconds)
	Timer float32
	// Brightness factor for DrawSelected
	SelectedLerp float32
	// Belt arrows offset
	BeltOffset float32
}

// Update animations state
//
// Do not depends on any other state
func (a *Animations) Update() {
	t := float32(rl.GetTime())
	a.Timer = t
	a.SelectedLerp = 0.1 + 0.2*math32.Sin(2.*rl.Pi*t)
	a.BeltOffset = math32.Mod(t, 1.)
}
