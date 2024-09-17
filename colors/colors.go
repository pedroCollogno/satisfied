package colors

import (
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func parseHex(hex string) uint8 {
	val, err := strconv.ParseInt(hex, 16, 32)
	if err != nil {
		panic(err)
	}
	return uint8(val)
}

func NewColorFromHex(hex string) rl.Color {
	return rl.Color{
		R: parseHex(hex[1:3]),
		G: parseHex(hex[3:5]),
		B: parseHex(hex[5:7]),
		A: 255,
	}
}

func lerpUint8(a, b uint8, t float32) uint8 {
	return uint8(float32(a)*(1-t) + float32(b)*t)
}

func Lerp(a, b rl.Color, t float32) rl.Color {
	return rl.Color{
		lerpUint8(a.R, b.R, t),
		lerpUint8(a.G, b.G, t),
		lerpUint8(a.B, b.B, t),
		lerpUint8(a.A, b.A, t),
	}
}

func WithAlpha(col rl.Color, a float32) rl.Color {
	return rl.Color{
		R: col.R,
		G: col.G,
		B: col.B,
		A: uint8(a * 255),
	}
}

var (
	Blank = rl.Color{0, 0, 0, 0}
	White = rl.Color{255, 255, 255, 255}
	Black = rl.Color{0, 0, 0, 255}

	// Colors from https://tailwindcss.com/docs/customizing-colors

	Gray100   = NewColorFromHex("#f3f4f6")
	Gray300   = NewColorFromHex("#d1d5db")
	Gray500   = NewColorFromHex("#6b7280")
	Gray700   = NewColorFromHex("#374151")
	Gray900   = NewColorFromHex("#111827")
	Blue500   = NewColorFromHex("#3b82f6")
	Blue700   = NewColorFromHex("#1d4ed8")
	Green500  = NewColorFromHex("#22c55e")
	Orange500 = NewColorFromHex("#f97316")
	Red500    = NewColorFromHex("#ef4444")
	Amber700  = NewColorFromHex("#b45309")
)
