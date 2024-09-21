package app

import (
	"encoding/json"
	"fmt"

	"github.com/bonoboris/satisfied/colors"
	"github.com/bonoboris/satisfied/matrix"
	rl "github.com/gen2brain/raylib-go/raylib"
)

////////////////////////////////////////////////////////////////////////////////////////////////////
// Path
////////////////////////////////////////////////////////////////////////////////////////////////////

type Path struct {
	DefIdx     int
	Start, End rl.Vector2
}

func (p Path) String() string {
	if p.DefIdx == -1 {
		return fmt.Sprintf("%s{%v %v %v %v}", "<invalid>", p.Start.X, p.Start.Y, p.End.X, p.End.Y)
	}
	return fmt.Sprintf("%s{%v %v %v %v}", p.Def().Class, p.Start.X, p.Start.Y, p.End.X, p.End.Y)
}

func (p Path) Def() PathDef { return pathDefs[p.DefIdx] }

func (p Path) DrawStart(state DrawState) {
	def := p.Def()
	if !dims.ExWorld.CheckCollisionPoint(p.Start) {
		// skip drawing if path start is outside of the scene
		return
	}
	color := state.transformColor(def.Color)
	// Path start
	rl.DrawCircleV(p.Start, def.Width/2, color)
}

func (p Path) DrawEnd(state DrawState) {
	def := p.Def()
	if !dims.ExWorld.CheckCollisionPoint(p.End) {
		// skip drawing if path end is outside of the scene
		return
	}
	color := state.transformColor(def.Color)
	// Path end
	rl.DrawCircleV(p.End, def.Width/2, color)
}

func (p Path) DrawBody(state DrawState) {
	def := p.Def()
	color := state.transformColor(def.Color)

	if !CheckCollisionRecLine(dims.ExWorld, p.Start, p.End) {
		// skip drawing if building is outside of the scene
		return
	}
	// we either call p.Draw(..) or p.DrawBody(..) p.DrawStart(..) and p.DrawEnd(..)
	app.drawCounts.Paths++

	// Path body
	rl.DrawLineEx(p.Start, p.End, def.Width, color)

	if !def.IsDirectional || state == DrawShadow {
		return
	}
	// Draw directional arrows
	color = state.transformColor(colors.Gray300)
	length := p.Start.Distance(p.End)
	angle := -p.Start.LineAngle(p.End)
	mat := matrix.NewTranslateV(p.Start).RotateRad(angle)
	for x := animations.BeltOffset; x < length; x += 1 {
		rl.DrawTriangle(
			mat.Apply(x-0.25, 0.5),
			mat.Apply(x+0.25, 0),
			mat.Apply(x-0.25, -0.5),
			color,
		)
	}
}

func (p Path) Draw(state DrawState) {
	def := p.Def()
	color := state.transformColor(def.Color)

	if !CheckCollisionRecLine(dims.ExWorld, p.Start, p.End) {
		// skip drawing if building is outside of the scene
		return
	}
	app.drawCounts.Paths++

	// Path start
	// FIXME: DrawCircle is very expensive, use shader instead
	rl.DrawCircleV(p.Start, def.Width/2, color)
	if p.Start.Equals(p.End) {
		return
	}
	// Path body
	rl.DrawLineEx(p.Start, p.End, def.Width, color)
	// Path end
	rl.DrawCircleV(p.End, def.Width/2, color)

	if !def.IsDirectional || state == DrawShadow {
		return
	}
	// Draw directional arrows
	color = state.transformColor(colors.Gray300)
	length := p.Start.Distance(p.End)
	angle := -p.Start.LineAngle(p.End)
	mat := matrix.NewTranslateV(p.Start).RotateRad(angle)
	for x := animations.BeltOffset; x < length; x += 1 {
		rl.DrawTriangle(
			mat.Apply(x-0.25, 0.5),
			mat.Apply(x+0.25, 0),
			mat.Apply(x-0.25, -0.5),
			color,
		)
	}
}

func (p Path) IsValid() bool {
	return p.Start != p.End
}

// Returns true if the given position is inside the path start.
func (p Path) CheckStartCollisionPoint(pos rl.Vector2) bool {
	return rl.CheckCollisionPointCircle(pos, p.Start, p.Def().Width/2)
}

// Returns true if the given position is inside the path end.
func (p Path) CheckEndCollisionPoint(pos rl.Vector2) bool {
	return rl.CheckCollisionPointCircle(pos, p.End, p.Def().Width/2)
}

// Returns true if the given position is inside the path body.
func (p Path) CheckCollisionPoint(pos rl.Vector2) bool {
	width := p.Def().Width
	lengthSqr := p.Start.DistanceSqr(p.End)
	angle := p.Start.LineAngle(p.End)
	transform := matrix.NewRotateRad(angle).TranslateV(p.Start.Negate())
	tpos := transform.ApplyV(pos)
	return tpos.X >= 0 && tpos.X*tpos.X <= lengthSqr && tpos.Y >= -width/2 && tpos.Y <= width/2
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// PathDef
////////////////////////////////////////////////////////////////////////////////////////////////////

type PathDef struct {
	Class         string
	Width         float32
	Color         rl.Color
	IsDirectional bool
}

func (def PathDef) String() string {
	return fmt.Sprintf("{%s W=%v, directional=%v}", def.Class, def.Width, def.IsDirectional)
}

func (def *PathDef) UnmarshalJSON(data []byte) error {
	type JsonPathDef struct {
		Class         string
		Width         float32
		Color         string
		IsDirectional bool
	}
	var jsonDef JsonPathDef
	err := json.Unmarshal(data, &jsonDef)
	if err != nil {
		return err
	}
	def.Class = jsonDef.Class
	def.Width = jsonDef.Width
	def.Color = colors.NewColorFromHex(jsonDef.Color)
	def.IsDirectional = jsonDef.IsDirectional
	return nil
}

type PathDefs []PathDef

func (defs PathDefs) Classes() []string {
	classes := make([]string, len(defs))
	for i, def := range defs {
		classes[i] = def.Class
	}
	return classes
}

func (defs PathDefs) Index(class string) int {
	for i, def := range defs {
		if def.Class == class {
			return i
		}
	}
	return -1
}
