package app

import (
	"embed"
	"encoding/json"

	"github.com/bonoboris/satisfied/log"
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

func readFile(fs embed.FS, path string) ([]byte, error) {
	log.Debug("assets.file", "status", "reading", "path", path)
	data, err := fs.ReadFile(path)
	if err != nil {
		log.Fatal("cannot read file", "path", path, "err", err)
		return nil, err
	}
	log.Info("assets.file", "status", "read", "path", path)
	return data, nil
}

func loadFont(fs embed.FS, path string) (rl.Font, error) {
	data, err := readFile(fs, path)
	if err != nil {
		return rl.Font{}, err
	}
	f := rl.LoadFontFromMemory(".ttf", data, fontLoadSize, nil)
	rl.SetTextureFilter(f.Texture, fontFilter)
	log.Debug("asset.font", "status", "loaded", "path", path)
	return f, nil
}

func LoadFonts(assets embed.FS) error {
	f, err := loadFont(assets, "assets/Roboto-Regular.ttf")
	if err != nil {
		log.Error("main font: using default", "err", err)
		font = rl.GetFontDefault()
		return err
	}
	font = f
	raygui.SetFont(f)
	log.Trace("assets.font", "status", "set in raygui")

	f, err = loadFont(assets, "assets/RobotoCondensed-Regular.ttf")
	if err != nil {
		log.Error("label font: using default", "err", err)
		labelFont = rl.GetFontDefault()
		return err
	}
	labelFont = f
	return nil
}

func LoadAssets(assets embed.FS) error {
	data, err := readFile(assets, "assets/building_defs.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &buildingDefs)
	if err != nil {
		log.Fatal("cannot parse building defs", "err", err)
		return err
	}
	log.Debug("assets.buildingsDefs", "status", "parsed", "count", len(buildingDefs))
	if log.WillTrace() {
		for i, def := range buildingDefs {
			log.Trace("assets.buildingDefs", "i", i, "value", def)
		}
	}

	data, err = readFile(assets, "assets/path_defs.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &pathDefs)
	if err != nil {
		log.Fatal("cannot parse path defs", "err", err)
		return err
	}
	log.Debug("assets.pathDefs", "status", "parsed", "count", len(pathDefs))
	if log.WillTrace() {
		for i, def := range pathDefs {
			log.Trace("assets.pathDefs", "i", i, "value", def)
		}
	}
	return nil
}

// LoadIcon loads the application icon
func LoadIcon(assets embed.FS) (*rl.Image, error) {
	data, err := readFile(assets, "assets/icon.png")
	if err != nil {
		return nil, err
	}
	img := rl.LoadImageFromMemory(".png", data, int32(len(data)))
	if img == nil {
		log.Error("cannot load icon as an image", "err", err)
	}
	return img, nil
}
