package app

import (
	"embed"
	"fmt"

	"github.com/bonoboris/satisfied/colors"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	windowTitle  = "Satisfied"
	windowWidth  = 1080
	windowHeight = 720
	windowFlags  = rl.FlagWindowResizable | rl.FlagWindowMaximized
	targetFPS    = 144
)

func LoadSomeScene() {
	for x := float32(0); x < 30; x += 10 {
		b1 := Building{DefIdx: 0, Pos: vec2(x+2, 0), Rot: 0}
		b2 := Building{DefIdx: 7, Pos: vec2(x, 20), Rot: 0}
		scene.Buildings = append(scene.Buildings, b1, b2)
		p := Path{DefIdx: 0, Start: vec2(x, 15), End: vec2(x, 7)}
		scene.Paths = append(scene.Paths, p)
	}
}

// Init initializes the application.
//
// It loads assets, initializes the window, and sets up the default state.
//
// It must only be called once at startup.
func Init(assets embed.FS) {
	// Loading assets
	LoadAssets(assets)

	// Init window
	rl.SetConfigFlags(rl.FlagWindowHighdpi | rl.FlagMsaa4xHint)
	rl.InitWindow(windowWidth, windowHeight, windowTitle)
	rl.SetTargetFPS(targetFPS)
	rl.SetWindowState(windowFlags)
	rl.SetExitKey(rl.KeyNull)

	// Loading font
	LoadFonts(assets)

	// Initializing state
	dims.Update()
	gui.Init()
	camera.doReset()
}

// Close cleanup resources used by the application before exiting.
func Close() {
	rl.UnloadFont(font)
	rl.UnloadFont(labelFont)
	rl.CloseWindow()
}

// ShouldQuit returns true if the application should exit.
func ShouldExit() bool {
	return rl.WindowShouldClose()
}

// Step updates and draw a frame.
func Step() {
	Update()
	rl.BeginDrawing()
	Draw()
	UpdateAndDrawGui()
	rl.EndDrawing()
}

// GetAction dispatches a call to the [GetActionFunc] of the current [AppMode].
//
// Most of the time it will returns `nil`.
func GetAction() Action {
	switch appMode {
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

// AppDispatch [TargetApp] actions their handler and returns the next [Action] to be performed.
func AppDispatch(action Action) Action {
	switch action := action.(type) {
	case ActionSwitchMode:
		return appMode.doSwitchMode(action.Mode, action.Resets)
	default:
		panic(fmt.Sprintf("AppDispatch: cannot handle: %T", action))
	}
}

// Dispatch the [Action] to its target [ActionHandler] and returns the next [Action] to be performed.
//
// Most of the time it will returns `nil`.
func DispatchAction(action Action) Action {
	// [ActionSwitchMode] is a special case, as it does not corresponds any [AppMode]
	if switchMode, ok := action.(ActionSwitchMode); ok {
		return appMode.doSwitchMode(switchMode.Mode, switchMode.Resets)
	}

	switch action.Target() {
	case TargetApp:
		return AppDispatch(action)
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

// Update the application state based on mouse and keyboard input(s).
func Update() {
	// input updates
	dims.Update()
	animations.Update()
	keyboard.Update()
	// FIXME: mouse.Update and camera.Update depends on each other (can introduce a 1 frame delay in state update)
	mouse.Update()

	camera.Update()

	if appMode == ModeNormal || appMode == ModeSelection && selection.mode == SelectionNormal {
		scene.Update(SceneIgnore{})
	} else {
		scene.Update(SceneIgnore{UndoRedo: true})
	}

	if keyboard.Pressed == rl.KeyEscape {
		appMode.doSwitchMode(ModeNormal, ResetAll())
	}

	for action := GetAction(); action != nil; action = DispatchAction(action) {
		// empty loop body
		// [GetAction] is called once per frame
		// [Update] is called in a loop until action chain is terminated
		//
		// In most cases, [GetAction] will recursively handle the action chain and return nil.
	}
}

// Draw draws the scene without updating the state.
func Draw() {
	rl.ClearBackground(colors.White)

	camera.BeginMode2D()

	grid.Draw()

	// draw placed objects
	scene.Draw()

	switch appMode {
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
}

// UpdateAndDraw combines drawing the GUI, handling GUI inputs and returns an [Action] to be performed.
//
// See: [ActionHandler]
func UpdateAndDrawGui() {
	for action := gui.UpdateAndDraw(); action != nil; action = DispatchAction(action) {
		// empty loop body
		// [GetAction] is called once per frame
		// [Update] is called in a loop until action chain is terminated
		//
		// In most cases, [Gui.UpdateAndDraw] will recursively handle the action chain and return nil.
	}
}
