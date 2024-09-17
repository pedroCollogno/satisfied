package app

import (
	"embed"

	"github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	fontLoadSize = 32
	fontFilter   = rl.FilterBilinear
)

var (
	// Building defs
	buildingDefs BuildingDefs
	// Path defs
	pathDefs PathDefs
	// Application font
	font rl.Font
	// Label font
	labelFont rl.Font
)

func LoadFonts(assets embed.FS) {
	data, err := assets.ReadFile("assets/Roboto-Regular.ttf")
	if err != nil {
		panic(err)
	}
	font = rl.LoadFontFromMemory(".ttf", data, fontLoadSize, nil)
	rl.SetTextureFilter(font.Texture, fontFilter)
	raygui.SetFont(font)

	data, err = assets.ReadFile("assets/RobotoCondensed-Regular.ttf")
	if err != nil {
		panic(err)
	}
	labelFont = rl.LoadFontFromMemory(".ttf", data, fontLoadSize, nil)
	rl.SetTextureFilter(labelFont.Texture, fontFilter)
}

func LoadAssets(assets embed.FS) {
	data, err := assets.ReadFile("assets/building_defs.json")
	if err != nil {
		panic(err)
	}
	buildingDefs = ParseBuildingDefs(data)
	data, err = assets.ReadFile("assets/path_defs.json")
	if err != nil {
		panic(err)
	}
	pathDefs = ParsePathDefs(data)
}
