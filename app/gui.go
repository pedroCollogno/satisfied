package app

import (
	"fmt"
	"slices"
	"strings"

	"github.com/bonoboris/satisfied/colors"
	"github.com/bonoboris/satisfied/log"
	"github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	// Top bar height in px
	TopbarHeight = 50.0
	// Side bar width in px
	SidebarWidth = 300.0
	// Status bar height in px
	StatusBarHeight = 30.0
)

var gui = Gui{}

// Gui is a container struct for all the GUI elements
type Gui struct {
	Topbar    guiTopbar
	Sidebar   guiSidebar
	Statusbar guiStatusbar
}

// Precompute and store some static data
func (gui *Gui) Init() {
	gui.Sidebar.init()
}

// UpdateAndDraw combines drawing the GUI, handling GUI inputs and returns an [Action] to be performed.
//
// Note: Raygui being an immediate mode GUI, we cannot seperate the update and draw steps.
func (g *Gui) UpdateAndDraw() (action Action) {
	// At most a single non nil action by frame should be returned from updateAndDraw calls
	// we cannot press 2 buttons at the same time
	action = orAction(action, g.Statusbar.updateAndDraw())
	action = orAction(action, g.Sidebar.updateAndDraw())
	action = orAction(action, g.Topbar.updateAndDraw())
	return action
}

// Reset resets GUI state
func (g *Gui) Reset() {
	g.Sidebar.Reset()
}

func (g *Gui) traceState() {
	log.Trace("gui.sidebar", "activePath", g.Sidebar.activePath, "activeCategory", g.Sidebar.activeCategory, "activeBuilding", g.Sidebar.activeBuilding)
}

// Dispatch performs an [Gui] action, updating its state, and returns an new action to be performed
//
// See: [ActionHandler]
func (g *Gui) Dispatch(action Action) Action {
	switch action := action.(type) {
	default:
		panic(fmt.Sprintf("Gui.Dispatch: cannot handle: %T", action))
	}
}

type guiTopbar struct{}

func (tb *guiTopbar) updateAndDraw() (action Action) {
	dims := rl.NewRectangle(0, 0, dims.Screen.X, TopbarHeight)

	rl.DrawRectangleRec(dims, colors.Gray100)
	rl.DrawLineV(dims.BottomLeft(), dims.BottomRight(), colors.Gray300)

	// Enable tooltip & set font size to 16
	pTextSize := raygui.GetStyle(raygui.DEFAULT, raygui.TEXT_SIZE)
	raygui.SetStyle(raygui.DEFAULT, raygui.TEXT_SIZE, 16)
	raygui.EnableTooltip()

	bounds := rl.NewRectangle(20, 10, 30, 30)
	raygui.SetTooltip("New file")
	if raygui.Button(bounds, raygui.IconText(raygui.ICON_FILE_NEW, "")) {
		log.Debug("topbar new file clicked")
		action = app.doNew()
	}

	bounds.X += 50
	raygui.SetTooltip("Open file")
	if raygui.Button(bounds, raygui.IconText(raygui.ICON_FILE_OPEN, "")) {
		log.Debug("topbar open file clicked")
		action = app.doOpen()
	}

	bounds.X += 50
	raygui.SetTooltip("Save file (Ctrl+S)")
	if raygui.Button(bounds, raygui.IconText(raygui.ICON_FILE_SAVE, "")) {
		log.Debug("topbar save file clicked")
		action = app.doSave(app.filepath)
	}

	bounds.X += 50
	raygui.SetTooltip("Save file as...")
	if raygui.Button(bounds, raygui.IconText(raygui.ICON_FILE_SAVE_CLASSIC, "")) {
		log.Debug("topbar save file as clicked")
		action = app.doSaveAs()
	}

	// Reset style and tooltip
	raygui.DisableTooltip()
	raygui.SetStyle(raygui.DEFAULT, raygui.TEXT_SIZE, pTextSize)

	return action
}

// guiSidebar represents the sidebar of the application
type guiSidebar struct {
	// Active path index
	activePath int32
	// Active category index
	activeCategory int32
	// Active building index (in active category)
	activeBuilding int32

	// Number of paths
	numPath int32
	// Number of categories
	numCategory int32
	// Path toggle group text
	pathText string
	// Category toggle group text
	categoryText string
	// Building toggle group texts
	buildingTexts []string
	// table of indices
	//
	// buildingIndices[activeCategory][activeBuilding]
	// is the index of the active building in buildingDefs
	buildingIndices [][]int
}

// Reset resets sidebar state
//
// We don't reset the active category because it's not what we usually want
func (sb *guiSidebar) Reset() {
	sb.activePath = -1
	sb.activeBuilding = -1
}

func (sb *guiSidebar) init() {
	paths := pathDefs.Classes()
	sb.numPath = int32(len(paths))
	sb.pathText = strings.Join(paths, ";")

	log.Trace("gui.sidebar", "paths", paths)
	log.Trace("gui.sidebar", "numPath", sb.numPath)
	log.Trace("gui.sidebar", "pathText", sb.pathText)

	categories := buildingDefs.Categories()
	sb.numCategory = int32(len(categories))
	sb.categoryText = strings.Join(categories, "\n")
	sb.buildingIndices = make([][]int, len(categories))
	log.Trace("gui.sidebar", "categories", categories)
	log.Trace("gui.sidebar", "numCategory", sb.numCategory)
	log.Trace("gui.sidebar", "categoryText", sb.categoryText)
	for i, def := range buildingDefs {
		idx := slices.Index(categories, def.Category)
		sb.buildingIndices[idx] = append(sb.buildingIndices[idx], i)
	}
	sb.buildingTexts = make([]string, len(categories))
	for i, cat_idxs := range sb.buildingIndices {
		classes := make([]string, len(cat_idxs))
		for j, idx := range cat_idxs {
			classes[j] = buildingDefs[idx].Class
		}
		sb.buildingTexts[i] = strings.Join(classes, "\n")
		log.Trace("gui.sidebar", "i", i, "buildingIndices[i]", cat_idxs)
		log.Trace("gui.sidebar", "i", i, "buildingTexts[i]", sb.buildingTexts[i])
	}
	sb.activePath = -1
	sb.activeCategory = -1
	sb.activeBuilding = -1
	gui.traceState()
}

func (sb *guiSidebar) drawLine(dims rl.Rectangle, yOffset float32) {
	start := vec2(dims.X, dims.Y+yOffset)
	end := vec2(dims.X+dims.Width, dims.Y+yOffset)
	rl.DrawLineEx(start, end, 2, colors.Gray300)
}

func (sb *guiSidebar) drawPathsControls(dims rl.Rectangle, yOffset float32) Action {
	bounds := rl.NewRectangle(dims.X, dims.Y+yOffset, dims.Width/float32(sb.numPath), 40)
	newActive := raygui.ToggleGroup(bounds, sb.pathText, int32(sb.activePath))
	if newActive != sb.activePath {
		// newActive is guaranteed to be != -1 because ToggleGroup returns the index of the newly
		// active toggle (after a click) and we cannot goes from an active one (sb.activePath != 1)
		// to an inactive one (newActive == -1) by clicking on the same toggle
		sb.activePath = newActive
		sb.activeCategory = -1
		sb.activeBuilding = -1
		log.Debug("sidebar path clicked", "defIdx", newActive)
		gui.traceState()
		// newActive matches with actual index in [pathDefs]
		return newPath.doInit(int(newActive))
	}
	return nil
}

func (sb *guiSidebar) drawCategoryControls(dims rl.Rectangle, yOffset float32) Action {
	bounds := rl.NewRectangle(dims.X, dims.Y+yOffset, dims.Width, 40)
	newActive := raygui.ToggleGroup(bounds, sb.categoryText, sb.activeCategory)
	if newActive != sb.activeCategory {
		// newActive is guaranteed to be != -1 because ToggleGroup returns the index of the newly
		// active toggle (after a click) and we cannot goes from an active one (sb.activeCategory != 1)
		// to an inactive one (newActive == -1) by clicking on the same toggle
		sb.activeCategory = newActive
		sb.activePath = -1
		sb.activeBuilding = -1
		log.Debug("sidebar category clicked", "catIdx", sb.activeCategory)
		gui.traceState()
		return app.doSwitchMode(ModeNormal, ResetAll().WithGui(false))
	}
	return nil
}

func (sb *guiSidebar) drawBuildingControls(dims rl.Rectangle, yOffset float32) Action {
	bounds := rl.NewRectangle(dims.X, dims.Y+yOffset, dims.Width, 40)
	newActive := raygui.ToggleGroup(bounds, sb.buildingTexts[sb.activeCategory], sb.activeBuilding)
	if newActive != sb.activeBuilding {
		// newActive is guaranteed to be != -1 because ToggleGroup returns the index of the newly
		// active toggle (after a click) and we cannot goes from an active one (sb.activeBuilding != 1)
		// to an inactive one (newActive == -1) by clicking on the same toggle
		sb.activeBuilding = newActive
		sb.activePath = -1
		defIdx := sb.buildingIndices[sb.activeCategory][newActive]
		log.Debug("sidebar building clicked", "defIdx", defIdx)
		gui.traceState()
		return newBuilding.doInit(defIdx)
	}
	return nil
}

func (sb *guiSidebar) updateAndDraw() (action Action) {
	dims := rl.NewRectangle(0, TopbarHeight, SidebarWidth, dims.Screen.Y-TopbarHeight-StatusBarHeight)

	rl.DrawRectangleRec(dims, colors.Gray100)
	rl.DrawLineV(dims.TopRight(), dims.BottomRight(), colors.Gray300)

	padding := float32(10)

	pPadding := raygui.GetStyle(raygui.TOGGLE, raygui.GROUP_PADDING)
	pTextSize := raygui.GetStyle(raygui.DEFAULT, raygui.TEXT_SIZE)
	raygui.SetStyle(raygui.TOGGLE, raygui.GROUP_PADDING, int64(padding))
	raygui.SetStyle(raygui.DEFAULT, raygui.TEXT_SIZE, 32)

	// padded dimensions
	pdims := rl.NewRectangle(dims.X+20, dims.Y+20, dims.Width-40, dims.Height-40)

	yOffset := float32(0)
	action = orAction(action, sb.drawPathsControls(pdims, yOffset))
	yOffset += 60

	sb.drawLine(pdims, yOffset)
	yOffset += 20

	action = orAction(action, sb.drawCategoryControls(pdims, yOffset))
	yOffset += float32(sb.numCategory) * 50

	if sb.activeCategory > -1 {
		sb.drawLine(pdims, yOffset)
		yOffset += 10

		action = orAction(action, sb.drawBuildingControls(pdims, yOffset))
	}

	raygui.SetStyle(raygui.TOGGLE, raygui.GROUP_PADDING, pPadding)
	raygui.SetStyle(raygui.DEFAULT, raygui.TEXT_SIZE, pTextSize)
	return action
}

type guiStatusbar struct{}

func (sb *guiStatusbar) updateAndDraw() Action {
	dims := rl.NewRectangle(0, dims.Screen.Y-StatusBarHeight, dims.Screen.X, StatusBarHeight)
	rl.DrawRectangleRec(dims, colors.Gray100)
	rl.DrawLineV(dims.TopLeft(), dims.TopRight(), colors.Gray300)

	// left aligned text
	lpos := dims.TopLeft().Add(vec2(5, 5))
	ltext := fmt.Sprintf("FPS=% 3d | %12v | Building Draws=%d | Path Draws=%d", int(rl.GetFPS()), app.Mode, app.drawCounts.Buildings, app.drawCounts.Paths)
	rl.DrawTextEx(font, ltext, lpos, 24, 1, colors.Gray700)

	// right aligned text
	rtext := fmt.Sprintf("x=%- 4d  y=%- 4d", int(mouse.SnappedPos.X), int(mouse.SnappedPos.Y))
	width := rl.MeasureTextEx(font, rtext, 24, 1).X
	rpos := dims.TopRight().Add(vec2(-5-width, 5))
	rl.DrawTextEx(font, rtext, rpos, 24, 1, colors.Gray700)

	return nil
}

// orAction returns the first non nil action, or nil if both are nil
func orAction(a, b Action) Action {
	if a != nil {
		return a
	}
	return b
}
