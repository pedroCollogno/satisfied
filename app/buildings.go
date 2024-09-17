package app

import (
	"encoding/json"
	"slices"
	"sort"
	"strings"

	"github.com/bonoboris/satisfied/colors"
	"github.com/bonoboris/satisfied/matrix"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Color = rl.Color

type inputOutput struct {
	Pos rl.Vector2
	Rot int32
}

func (io inputOutput) matrix() matrix.Matrix {
	return matrix.NewTranslateV(io.Pos).Rotate(io.Rot)
}

func (inputOutput) drawTri(mat matrix.Matrix, x, y float32, c rl.Color) {
	rl.DrawTriangle(
		mat.Apply(x-0.25, y+0.25),
		mat.Apply(x+0.25, y+0.25),
		mat.Apply(x, y-0.25),
		c,
	)
}

func (io inputOutput) drawBeltInput(mat matrix.Matrix, state DrawState) {
	mat = mat.Mult(io.matrix())
	bounds := rl.NewRectangle(-1, -0.5, 2, 0.5)
	rl.DrawRectangleRec(mat.ApplyRecRec(bounds), state.transformColor(colors.Orange500))
	c := state.transformColor(colors.Black)
	io.drawTri(mat, -0.5, -0.25, c)
	io.drawTri(mat, 0, -0.25, c)
	io.drawTri(mat, 0.5, -0.25, c)
}

func (io inputOutput) drawBeltOutput(mat matrix.Matrix, state DrawState) {
	mat = mat.Mult(io.matrix())
	bounds := rl.NewRectangle(-1, 0, 2, 0.5)
	rl.DrawRectangleRec(mat.ApplyRecRec(bounds), state.transformColor(colors.Green500))
	c := state.transformColor(colors.Black)
	io.drawTri(mat, -0.5, 0.25, c)
	io.drawTri(mat, 0, 0.25, c)
	io.drawTri(mat, 0.5, 0.25, c)
}

func (io inputOutput) drawPipeInput(mat matrix.Matrix, state DrawState) {
	mat = mat.Mult(io.matrix())
	bounds := rl.NewRectangle(-0.5, -1, 1, 0.5)
	rl.DrawRectangleRec(mat.ApplyRecRec(bounds), state.transformColor(colors.Orange500))
	c := state.transformColor(colors.Black)
	io.drawTri(mat, 0, -0.25, c)
}

func (io inputOutput) drawPipeOutput(mat matrix.Matrix, state DrawState) {
	mat = mat.Mult(io.matrix())
	bounds := rl.NewRectangle(-0.5, 0, 1, 0.5)
	rl.DrawRectangleRec(mat.ApplyRecRec(bounds), state.transformColor(colors.Green500))
	c := state.transformColor(colors.Black)
	io.drawTri(mat, 0, 0.25, c)
}

type BuildingDef struct {
	Class    string
	Category string
	Dims     rl.Vector2
	BeltIn   []inputOutput
	BeltOut  []inputOutput
	PipeIn   []inputOutput
	PipeOut  []inputOutput
}

type Building struct {
	DefIdx int
	Pos    rl.Vector2
	Rot    int32
}

func (b Building) Def() BuildingDef { return buildingDefs[b.DefIdx] }

func (b Building) matrix() matrix.Matrix {
	mid := grid.Snap(b.Def().Dims.Scale(0.5))
	return matrix.NewTranslateV(b.Pos).Rotate(b.Rot).TranslateV(mid.Negate())
}

func (b Building) Bounds() rl.Rectangle {
	dims := b.Def().Dims
	return b.matrix().ApplyRec(0, 0, dims.X, dims.Y)
}

// func (b Building) DrawLabel(bounds rl.Rectangle) {
// 	zoom := camera.Zoom()
// 	fontSize := float32(24) / zoom
// 	spacing := float32(1) / zoom
// 	lines := strings.Split(b.Def().Class, " ")
// 	center := bounds.Center()
// 	iOff := float32(len(lines)-1)/2 + 0.5
// 	for i, line := range lines {
// 		width := rl.MeasureTextEx(font, line, fontSize, spacing).X
// 		pos := center.Sub(width/2, (iOff-float32(i))*fontSize)
// 		rl.DrawTextEx(font, line, pos, fontSize, spacing, rl.Color{0, 0, 0, 127})

//		}
//	}
const (
	labelFontSize    = 24.
	labelLineSpacing = -5.
)

var labelColor = rl.Color{0, 0, 0, 127}

func (b Building) DrawLabel(bounds rl.Rectangle) {
	bounds.X += 0.5
	bounds.Y += 0.5
	bounds.Width -= 1
	bounds.Height -= 1
	zoom := camera.Zoom()
	screenPos := camera.ScreenPos(bounds.TopLeft())
	screenSize := bounds.Size().Scale(zoom)

	lines := strings.Split(b.Def().Class, " ")
	n := float32(len(lines))
	fontSize := labelFontSize / zoom
	dy := (labelFontSize + labelLineSpacing) / zoom
	yOff := max(0, (bounds.Height-n*dy)/2)

	rl.BeginScissorMode(int32(screenPos.X), int32(screenPos.Y), int32(screenSize.X), int32(screenSize.Y))
	for i, line := range lines {
		width := rl.MeasureTextEx(labelFont, line, fontSize, 0).X
		xOff := max(0, (bounds.Width-width)/2)
		pos := bounds.TopLeft().Add(vec2(xOff, yOff+float32(i)*dy))
		rl.DrawTextEx(labelFont, line, pos, fontSize, 0, labelColor)
	}
	rl.EndScissorMode()
}

func (b Building) Draw(state DrawState) {
	mat := b.matrix()
	def := b.Def()
	bounds := mat.ApplyRec(0, 0, def.Dims.X, def.Dims.Y)
	rl.DrawRectangleRec(bounds, state.transformColor(colors.Blue500))

	if state == DrawShadow {
		return
	}

	rl.DrawRectangleLinesEx(bounds, 0.5, state.transformColor(colors.Blue700))

	for _, input := range def.BeltIn {
		input.drawBeltInput(mat, state)
	}
	for _, output := range def.BeltOut {
		output.drawBeltOutput(mat, state)
	}
	for _, input := range def.PipeIn {
		input.drawPipeInput(mat, state)
	}
	for _, output := range def.PipeOut {
		output.drawPipeOutput(mat, state)
	}
	b.DrawLabel(bounds)
}

func ParseBuildingDefs(data []byte) BuildingDefs {
	var defs BuildingDefs
	err := json.Unmarshal(data, &defs)
	if err != nil {
		panic(err)
	}
	return defs
}

type BuildingDefs []BuildingDef

func (defs BuildingDefs) Classes() []string {
	classes := make([]string, len(defs))
	for i, def := range defs {
		classes[i] = def.Class
	}
	return classes
}

func (defs BuildingDefs) Categories() []string {
	var categories []string
	for _, def := range defs {
		if !slices.Contains(categories, def.Category) {
			categories = append(categories, def.Category)
		}
	}
	sort.Strings(categories)
	return categories
}
