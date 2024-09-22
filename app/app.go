package app

import (
	"embed"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/bonoboris/satisfied/colors"
	"github.com/bonoboris/satisfied/log"
	tfd "github.com/bonoboris/satisfied/tinyfiledialogs"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	// Version of the save file format
	version          = 0
	windowTitle      = "Satisfied"
	extFilter        = "*.satisfied"
	extFilterDesc    = "Satisfied project"
	windowWidth      = 1080
	windowHeight     = 720
	windowFlags      = rl.FlagWindowResizable | rl.FlagWindowMaximized
	DefaultTargetFPS = 30
)

var app App

// App contains application global state
type App struct {
	// Application view mode
	Mode AppMode
	// File path
	filepath string
	// window title
	title string
	// draw counts
	drawCounts DrawCounts
	// whether the app has panicked
	hasPanicked bool
}

type DrawCounts struct {
	Buildings int
	Paths     int
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// Load / save
////////////////////////////////////////////////////////////////////////////////////////////////////

// save project to file and updates window title and [App.filepath] on success
func (a *App) saveFile(filepath string) error {
	log.Info("saving project", "path", filepath)
	file, err := os.Create(filepath)
	if err != nil {
		log.Error("cannot create file", "path", filepath, "err", err)
		return err
	}
	defer file.Close()
	err = scene.SaveToText(file)
	if err != nil {
		log.Error("cannot write to file", "path", filepath, "err", err)
		return err
	}
	a.filepath = filepath
	scene.ResetModified()
	log.Info("project saved", "path", filepath)
	return nil
}

func (a *App) loadFile(filepath string) error {
	// On error, log error, display message
	log.Info("loading project", "path", filepath)
	file, err := os.Open(filepath)
	if err != nil {
		log.Error("cannot open file", "path", filepath, "err", err)
		return err
	}
	defer file.Close()
	fileScene := Scene{}
	err = fileScene.LoadFromText(file)
	if err != nil {
		log.Error("error parsing project", "path", filepath, "err", err)
		msg := fmt.Sprintf("Cannot load project: %s\n\nError: %s", filepath, RemoveQuotes(err.Error()))
		tfd.MessageBox(windowTitle+" -Error loading file", msg, tfd.DialogOk, tfd.IconError, tfd.ButtonOkYes)
		return err
	}
	a.filepath = filepath
	scene = fileScene
	log.Info("project loaded", "path", filepath)
	return nil
}

// checkUnsavedChanges checks and asks the user for what to do if current project has unsaved changes.
// It returns whether we can continue or not.
//
// It ask the user in the following cases:
//   - current project exists on disk but has unsaved changes
//   - current project does not exist on disk and is not empty
//
// User can choose to:
//   - save the current changes, returns true
//   - discard the current changes, returns true
//   - cancel the operation, returns false
func (a *App) checkUnsavedChanges() bool {
	if a.filepath == "" {
		log.Debug("check unsaved changes", "project", "new", "isEmpty", scene.IsEmpty())
		if scene.IsEmpty() {
			return true
		} else {
			msg := "The current project is not saved.\n\nDo you want to save it ?"
			switch tfd.MessageBox(windowTitle+" - Unsaved project", msg, tfd.DialogYesNoCancel, tfd.IconWarning, tfd.ButtonOkYes) {
			case tfd.ButtonOkYes:
				log.Debug("check unsaved changes", "action", "saveAs")
				if err := a.saveFile(a.filepath); err != nil {
					msg := fmt.Sprintf("Cannot save project: %s", RemoveQuotes(err.Error()))
					tfd.MessageBox(windowTitle+" - Error saving file", msg, tfd.DialogOk, tfd.IconError, tfd.ButtonOkYes)
					log.Error("check unsaved changes", "action", "cancel", "err", err)
					return false
				}
				return true
			case tfd.ButtonNo:
				log.Debug("check unsaved changes", "action", "discard")
				return true
			case tfd.ButtonCancelNo:
				log.Debug("check unsaved changes", "action", "cancel")
				return false
			default:
				panic("invalid button")
			}
		}
	} else {
		log.Debug("check unsaved changes", "project", "existing", "isModified", scene.IsModified())
		if !scene.IsModified() {
			return true
		} else {
			msg := "The current project has unsaved changes.\n\nDo you want to save them ?"
			switch tfd.MessageBox(windowTitle+" - Unsaved project", msg, tfd.DialogYesNoCancel, tfd.IconWarning, tfd.ButtonOkYes) {
			case tfd.ButtonOkYes:
				log.Debug("check unsaved changes", "action", "save")
				if err := a.saveFile(a.filepath); err != nil {
					msg := fmt.Sprintf("Cannot save file: %s", RemoveQuotes(err.Error()))
					tfd.MessageBox(windowTitle+"Error saving file", msg, tfd.DialogOk, tfd.IconError, tfd.ButtonOkYes)
					log.Error("check unsaved changes", "action", "cancel", "err", err)
					return false
				}
				return true
			case tfd.ButtonNo:
				log.Debug("check unsaved changes", "action", "discard")
				return true
			case tfd.ButtonCancelNo:
				log.Debug("check unsaved changes", "action", "cancel")
				return false
			default:
				panic("invalid button")
			}
		}
	}
}

func (a *App) doNew() Action {
	log.Info("new project")
	if !a.checkUnsavedChanges() {
		return nil
	}
	app.filepath = ""
	scene.Buildings = scene.Buildings[:0]
	scene.Paths = scene.Paths[:0]
	return a.doSwitchMode(ModeNormal, ResetAll().WithCamera(true))
}

func (a *App) doOpen() Action {
	log.Info("open project")
	if !a.checkUnsavedChanges() {
		return nil
	}
	filepath, ok := tfd.OpenFileDialog("Open project", "", []string{extFilter}, extFilterDesc)
	if ok {
		log.Debug("open project", "action", "load", "path", filepath)
		if err := app.loadFile(filepath); err != nil {
			log.Error("open project", "action", "cancel", "err", err)
		} else {
			return a.doSwitchMode(ModeNormal, ResetAll().WithCamera(true))
		}
	} else {
		log.Debug("open project", "action", "cancel")
	}
	return nil
}

func (a *App) doSaveAs() Action {
	log.Info("save project as")
	filepath, ok := tfd.SaveFileDialog("Save project as...", a.filepath, []string{extFilter}, extFilterDesc)
	if ok {
		return a.doSave(filepath)
	}
	return nil
}

func (a *App) doSave(filepath string) Action {
	if filepath == "" {
		return a.doSaveAs()
	}
	if err := a.saveFile(filepath); err != nil {
		msg := fmt.Sprintf("Cannot save file: %s\n\nError: %s", filepath, RemoveQuotes(err.Error()))
		tfd.MessageBox(windowTitle+" - Error saving file", msg, tfd.DialogOk, tfd.IconError, tfd.ButtonOkYes)
		return nil
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// GUI actions
////////////////////////////////////////////////////////////////////////////////////////////////////

func (a *App) doUndo() Action {
	if a.Mode == ModeNormal || a.Mode == ModeSelection && selection.mode == SelectionNormal {
		scene.Undo()
	}
	return nil
}

func (a *App) doRedo() Action {
	if a.Mode == ModeNormal || a.Mode == ModeSelection && selection.mode == SelectionNormal {
		scene.Redo()
	}
	return nil
}

func (a *App) doDelete() Action {
	switch app.Mode {
	case ModeSelection:
		if selection.mode == SelectionNormal {
			return selection.doDelete()
		}
	}
	return nil
}

func (a *App) doRotate() Action {
	switch app.Mode {
	case ModeNewPath:
		return newPath.doReverse()
	case ModeNewBuilding:
		return newBuilding.doRotate()
	case ModeSelection:
		return selection.doRotate()
	}
	return nil
}

func (a *App) doDuplicate() Action {
	switch app.Mode {
	case ModeSelection:
		if selection.mode == SelectionNormal {
			return selection.doBeginTransformation(SelectionDuplicate, selection.Bounds.Center())
		}
	}
	return nil
}

func (a *App) doDrag() Action {
	switch app.Mode {
	case ModeSelection:
		if selection.mode == SelectionNormal {
			return selection.doBeginTransformation(SelectionDrag, selection.Bounds.Center())
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// Switch mode
////////////////////////////////////////////////////////////////////////////////////////////////////

// doSwitchMode sets [App.Mode] to the given [AppMode] and reset other modes state
//
// See: [ActionHandler]
func (a *App) doSwitchMode(mode AppMode, resets Resets) Action {
	log.Debug("switchAppMode", "mode", mode, "resets", resets)
	a.Mode = mode
	if resets.Selector {
		selector.Reset()
	}
	if resets.NewPath {
		newPath.Reset()
	}
	if resets.NewBuilding {
		newBuilding.Reset()
	}
	if resets.Selection {
		selection.Reset()
	}
	if resets.Gui {
		gui.Reset()
	}
	return nil
}

// Returns wether the app is in [ModeNormal] or [ModeSelection] with [SelectionNormal] sub-mode
func (a *App) isNormal() bool {
	return a.Mode == ModeNormal || a.Mode == ModeSelection && selection.mode == SelectionNormal
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// Main functions
////////////////////////////////////////////////////////////////////////////////////////////////////

type AppOptions struct {
	// A file to load
	File string
	// Target / Max FPS
	Fps int
}

// Init initializes the application.
//
// It loads assets, initializes the window, and sets up the default state.
//
// It must only be called once at startup.
func Init(assets embed.FS, opts *AppOptions) error {
	if opts == nil {
		opts = &AppOptions{
			Fps: DefaultTargetFPS,
		}
	}
	if opts.Fps <= 0 {
		opts.Fps = DefaultTargetFPS
	}

	log.Info("initializing application")
	log.Info("options", "targetFPS", opts.Fps)
	// Loading assets
	if err := LoadAssets(assets); err != nil {
		return err
	}
	log.Info("assets loaded")

	// Init window
	rl.SetConfigFlags(rl.FlagWindowHighdpi | rl.FlagMsaa4xHint)
	rl.InitWindow(windowWidth, windowHeight, windowTitle)
	rl.SetTargetFPS(int32(opts.Fps))
	rl.SetWindowState(windowFlags)
	rl.SetExitKey(rl.KeyNull)
	if icon, err := LoadIcon(assets); err == nil {
		rl.SetWindowIcon(*icon)
	}

	log.Info("window initialized")

	// Loading font
	if err := LoadFonts(assets); err != nil {
		return err
	}
	log.Info("fonts loaded")

	// Initializing state
	dims.Update()
	gui.Init()
	camera.doReset()
	log.Info("state initialized")

	app.Mode = ModeNormal

	if opts.File != "" {
		if err := app.loadFile(opts.File); err != nil {
			log.Error("init app with empty scene", "err", err)
		}
	}
	return nil
}

// Close cleanup resources used by the application before exiting.
func Close() {
	rl.UnloadFont(font)
	rl.UnloadFont(labelFont)
	rl.CloseWindow()
}

// ShouldQuit returns true if the application should exit.
func ShouldExit() bool {
	return rl.WindowShouldClose() || app.hasPanicked
}

// Step updates and draw a frame.
func Step() {
	defer panicHandler()
	update()
	rl.BeginDrawing()
	draw()
	updateAndDrawGui()
	rl.EndDrawing()
}

// getAction dispatches a call to the [GetActionFunc] of the current [AppMode].
//
// Most of the time it will returns `nil`.
func getAction() Action {
	switch app.Mode {
	case ModeNormal:
		return selector.GetAction()
	case ModeNewPath:
		return newPath.GetAction()
	case ModeNewBuilding:
		return newBuilding.GetAction()
	case ModeSelection:
		return selection.GetAction()
	default:
		panic("Invalid app mode")
	}
}

// Dispatch [TargetApp] actions their handler and returns the next [Action] to be performed.
func (app *App) dispatch(action Action) Action {
	switch action := action.(type) {
	case AppActionSwitchMode:
		return app.doSwitchMode(action.Mode, action.Resets)
	default:
		panic(fmt.Sprintf("appDispatch: cannot handle: %T", action))
	}
}

// Dispatch the [Action] to its target [ActionHandler] and returns the next [Action] to be performed.
//
// Most of the time it will returns `nil`.
func dispatchAction(action Action) Action {
	switch action.Target() {
	case TargetApp:
		return app.dispatch(action)
	case TargetGui:
		return gui.Dispatch(action)
	case TargetCamera:
		return camera.Dispatch(action)
	case TargetSelector:
		return selector.Dispatch(action)
	case TargetNewPath:
		return newPath.Dispatch(action)
	case TargetNewBuilding:
		return newBuilding.Dispatch(action)
	case TargetSelection:
		return selection.Dispatch(action)
	default:
		panic("Invalid action target")
	}
}

func (a *App) update() {
	// reset draw counts
	a.drawCounts = DrawCounts{}
	// update window title
	if a.filepath != "" {
		title := ""
		if scene.IsModified() {
			title += "•"
		}
		title += fmt.Sprintf("%s - %s", path.Base(a.filepath), windowTitle)
		if a.title != title {
			log.Debug("app.update", "title", title)
			rl.SetWindowTitle(title)
			a.title = title
		}
	} else {
		title := fmt.Sprintf("%s - %s", "Unsaved project", windowTitle)
		if a.title != title {
			log.Debug("app.update", "title", title)
			rl.SetWindowTitle(title)
			a.title = title
		}
	}
	// check for save shortcut
	if a.isNormal() && keyboard.Ctrl {
		switch keyboard.Pressed {
		case 's':
			if keyboard.Shift {
				if a.filepath == "" {
					a.doSaveAs()
				} else {
					a.doSave(a.filepath)
				}
			}
		}
	}
}

// Update the application state based on mouse and keyboard input(s).
func update() {
	// input updates
	animations.Update()
	keyboard.Update()

	dims.Update()
	camera.Update()
	mouse.Update()
	// FIXME: there some cyclic dependencies between mouse, camera and dims

	app.update()
	scene.Update()

	if keyboard.Pressed == rl.KeyEscape && app.isNormal() {
		app.doSwitchMode(ModeNormal, ResetAll())
	}

	for action := getAction(); action != nil; action = dispatchAction(action) {
		// empty loop body
		// [GetAction] is called once per frame
		// [Update] is called in a loop until action chain is terminated
		//
		// In most cases, [GetAction] will recursively handle the action chain and return nil.
	}
}

// Draw draws the scene without updating the state.
func draw() {
	rl.ClearBackground(colors.White)
	rl.BeginScissorMode(int32(dims.Scene.X), int32(dims.Scene.Y), int32(dims.Scene.Width), int32(dims.Scene.Height))
	camera.BeginMode2D()

	grid.Draw()

	// draw placed objects
	scene.Draw()

	switch app.Mode {
	case ModeNormal:
		selector.Draw()
	case ModeNewPath:
		newPath.Draw()
	case ModeNewBuilding:
		newBuilding.Draw()
	case ModeSelection:
		selection.Draw()
	}

	camera.EndMode2D()
	rl.EndScissorMode()
}

// UpdateAndDraw combines drawing the GUI, handling GUI inputs and returns an [Action] to be performed.
//
// See: [ActionHandler]
func updateAndDrawGui() {
	for action := gui.UpdateAndDraw(); action != nil; action = dispatchAction(action) {
		// empty loop body
		// [GetAction] is called once per frame
		// [Update] is called in a loop until action chain is terminated
		//
		// In most cases, [Gui.UpdateAndDraw] will recursively handle the action chain and return nil.
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////
// panic handler
////////////////////////////////////////////////////////////////////////////////////////////////////

const panicTitle = "Satisfied has crashed"

const panicMessageNoBackupFile = `Satisfied has crashed and failed to backup current project.

Error: %s

Failed to backup reason: Cannot find user home directory.

Sorry for the inconvenience, please report this issue to the developer.`

const panicMessageErrBackup = `Satisfied has crashed and failed to backup current project.

Error: %s

Tried to backup in file: %s
But failed because: %s

Sorry for the inconvenience, please report this issue to the developer.`

const panicMessageBackupOk = `Satisfied has crashed.

Error: %s

Current project has been saved in file: %s

Sorry for the inconvenience, please report this issue to the developer.`

func panicHandler() {
	// TODO: save logs, link repo in error message.
	panicErr := recover()
	if panicErr == nil {
		return
	}
	app.hasPanicked = true // schedule app exit
	log.Fatal("application panic", "err", panicErr)
	savepath := app.filepath
	if savepath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			msg := fmt.Sprintf(panicMessageNoBackupFile, panicErr)
			tfd.MessageBox(panicTitle, msg, tfd.DialogOk, tfd.IconError, tfd.ButtonOkYes)
			return
		}
		savepath = NormalizePath(filepath.Join(home, "recover.satisfied"))
	} else {
		savepath = NormalizePath(savepath)
		ext := filepath.Ext(savepath)
		savepath = savepath[:len(savepath)-len(ext)] + ".recover" + ext
	}

	if err := app.saveFile(savepath); err != nil {
		msg := fmt.Sprintf(panicMessageErrBackup, panicErr, savepath, err)
		tfd.MessageBox(panicTitle, msg, tfd.DialogOk, tfd.IconError, tfd.ButtonOkYes)
		return
	}

	msg := fmt.Sprintf(panicMessageBackupOk, panicErr, savepath)
	tfd.MessageBox(panicTitle, msg, tfd.DialogOk, tfd.IconError, tfd.ButtonOkYes)
}
