package app

import (
	"github.com/bonoboris/satisfied/colors"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// DrawState enumerates object drawing states
type DrawState uint8

const (
	// States

	DrawNormal   DrawState = 0
	DrawNew      DrawState = 1
	DrawSelected DrawState = 2
	DrawInvalid  DrawState = 3
	DrawShadow   DrawState = 4
	DrawSkip     DrawState = 5

	// Modifiers

	DrawHovered DrawState = 1 << 6
	DrawClicked DrawState = 2 << 6
)

const (
	drawStateMask    = 0b00111111
	drawModifierMask = 0b11000000
)

// transformColor returns a color modified according to the draw state
func (state DrawState) transformColor(color rl.Color) rl.Color {
	// State
	switch state & drawStateMask {
	case DrawNormal:
	case DrawSelected:
		color = rl.ColorBrightness(color, animations.SelectedLerp)
	case DrawNew:
		color = rl.ColorBrightness(color, 0.5)
	case DrawInvalid:
		color = colors.Lerp(color, colors.Red500, 0.5)
	case DrawShadow:
		color = colors.Gray300
	case DrawSkip:
		return colors.Blank // FIXME: should panic ?
	default:
		panic("transformColor: invalid ToolState")
	}
	// Modifier
	switch state & drawModifierMask {
	case DrawClicked:
		color = rl.ColorBrightness(color, -0.2)
	case DrawHovered:
		color = rl.ColorBrightness(color, 0.2)
	}
	return color
}
