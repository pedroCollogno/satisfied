package app

import (
	"encoding/json"

	"github.com/bonoboris/satisfied/colors"
	"github.com/bonoboris/satisfied/matrix"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type PathDef struct {
	Class         string
	Width         float32
	Color         rl.Color
	IsDirectional bool
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

type Path struct {
	DefIdx     int
	Start, End rl.Vector2
}

func (p Path) Def() PathDef { return pathDefs[p.DefIdx] }

func (p Path) Draw(state DrawState) {
	def := p.Def()
	color := state.transformColor(def.Color)
	// Path start
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

// Returns true if the given position is inside the path body.
func (p Path) CheckCollisionPoint(pos rl.Vector2) bool {
	width := p.Def().Width
	lengthSqr := p.Start.DistanceSqr(p.End)
	angle := p.Start.LineAngle(p.End)
	transform := matrix.NewRotateRad(angle).TranslateV(p.Start.Negate())
	tpos := transform.ApplyV(pos)
	return tpos.X >= 0 && tpos.X*tpos.X <= lengthSqr && tpos.Y >= -width/2 && tpos.Y <= width/2
}

func ParsePathDefs(data []byte) PathDefs {
	var defs PathDefs
	err := json.Unmarshal(data, &defs)
	if err != nil {
		panic(err)
	}
	return defs
}

type PathDefs []PathDef

func (defs PathDefs) Classes() []string {
	classes := make([]string, len(defs))
	for i, def := range defs {
		classes[i] = def.Class
	}
	return classes
}
