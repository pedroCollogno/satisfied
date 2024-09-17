# Satisfied (WIP)

Satisfied is a 2D factory planner tool for the [Satisfactory](https://www.satisfactorygame.com/).

This is a work in progress, and not usable yet.

## Quick start

For now, you need to build the project yourself.

### Building from source

#### Windows
I would recommend using [w64devkit](https://github.com/skeeto/w64devkit) to setup the build environment on windows.

Extract it as instructed in the README, and add the `.../w64devkit/bin` folder to your `PATH` environment variable.

#### Other platform, cross-compiling
Refers to the [raylib-go](https://github.com/bonoboris/raylib-go) instructions for your platform.

This should be enough as the project is not using any external dependencies other than raylib-go.

#### Debug build
```sh
go build
```
#### Release build

```sh
CGO_CPPFLAGS="-O3 -DNDEBUG -flto" go build -ldflags="-s -w -H=windowsgui"
```

## Why this project?

I'm learning [Go](https://go.dev/) and I had wanted to play with [Raylib](https://www.raylib.com/). 

When I started to write this tool, Coffe Stain Studios had announced the release of the much awaited
1.0 version of Satisfactory.
I have few hundreds of hours on the game and never found a tool/way to plan my factories the way I wanted so I decided to write one.

This is more of a **learning & personal** project right now, so **use it at your own risk**.

## V1 Features Roadmap

- [x] Infinite grid canvas
- [x] Place buildings
- [x] Draw paths (belt and pipes)
- [x] Snap to grid (resolution of 1 game meter)
- [x] Rotate by 90° increments
- [x] Single / multi selection
- [x] Click and drag to move selection
- [x] Delete selection
- [x] Undo / redo (WIP)
- [ ] Move paths by their ends
- [ ] Save and load projects
- [ ] Complete buildings list for Production / Power / Logistics related buildings
- [ ] Scroll bar in side panel
- [ ] Keybindings displayed somewhere (status bar or popup)
- [ ] Logs/crash reports 
- [ ] Free text box tool

### Other goals

A list of features that may or may not happen in the future.

- [ ] Anchor paths to building inputs / outputs
- [x] Add a `Action` interface:

    Action would be represent a state update to be preformed.
    The idea is to decorelate inputs and state updates, and truly limit to state written by .Update() methods to their own; it could also be nice for implementing undo / redo.

    Could split `func Update()` methods into `func ProcessInputs() Action` methods and `fn Update(Action) Action` methods, where the returned action (if any) would allow to change the app mode and set data in other state variables.

    Example:\
    `selector.ProcessInput()` could return a `SetSelectionFromRect` action, with the rect dimensions, it would be then be processed by `selection.Update()` to set the selection, which in turn
    would return a `SetMode` action to change the app mode to `Selection` if needed.

- [ ] Porting to app to the web, Rust + [raylib-rs](https://github.com/deltaphc/raylib-rs) + WASM
    - [ ] Local storage auto save
    - [ ] Import / export (to file or clipboard)
    - [ ] Deploy
- [ ] Make it look nice (game icons, texture for each building, nicer UI)
- [ ] Quick access bar
- [ ] Zones / groups to represents factories and/or production lines
- [ ] Add items and recipes (big one)
    - [ ] Add them for planning only
    - [ ] Item cost of factory / selection
    - [ ] Compute production (static)
- [ ] Settings / customization (only if this is used by anyone other than me)
    - [ ] Remap keybindings
    - [ ] Change fonts / colors


## Design / Architecture

The application is written in [Go](https://go.dev/) and uses [Raylib](https://www.raylib.com/) for rendering.

I use a slightly modified version of [raylib-go](https://github.com/bonoboris/raylib-go) bindings 
(check out the [original here](https://github.com/gen2brain/raylib-go)).

The codebase is split into files by topic; each of them may contains update and/or draw logic.

**TODO**: Update this part: 
    - new files
    - Kbd/mouseInput -> `GetAction` -> action -> `Dispatch` -> `doXXX` methods 
    - `Dispatch` bypass: `GetAction` directly calls `doXXX` methods

#
*main package*
- `main.go`: entry point, contains the render loop, calls into `app/app.go`
#
*app package (most of the code)*
- `app/app.go`: defines the main functions:
    - `Init()`: initializes the application & creates a window
    - `Close()`: destroys the window and cleans up resources
    - `Step()`: the render loop body, performs updates and draw a frame
        - `Update()`: updates the application state based on mouse and keyboard input(s)
        - `Draw()`: draws the scene without updating the state
        - `DrawGUI()`: draws the GUI and may update the state based on GUI controls events

- `app/appMode.go`: AppMode enum definition and methods and appMode state variable
- `app/drawState.go`: DrawState enum definition and methods (normal, new, selected, hovered, shadow, ...)
- `app/assets.go`: static assets (fonts, buildings definitions, etc.)

- `app/gui.go`: GUI (topbar, sidebar, statusbar) related code
    - `Draw() GuiEvent`: draws the GUI and returns an GUI event

#
*update only*
- `app/animations.go`: animations timers
- `app/dims.go`: Screen and scene dimensions
- `app/mouse.go`: holds mouse state (position, button press, down, release, ...)
- `app/keyboard.go`: holds keyboard state (pressed key, shift, ctrl, alt, ...)
- `app/camera.go`: camera state (position, zoom, rotation)
    - `Update()`: updates the camera state based on mouse and keyboard input
    - `BeginMode2D()`: starts a 2D camera mode
    - `EndMode2D()`: ends a 2D camera mode
    - `WorldPos()`: converts a screen position to a world position
    - `ScreenPos()`: converts a world position to a screen position
#
*draw only*
- `app/grid.go`: grid drawing code
- `app/scene.go`: holds the scene objects (buildings and pipes) and some helper functions
    - `Draw()`: draws the 'normal' scene objects (those not handled in other places)
- `app/buildings.go`: buildings struct and drawing code
    - `Draw(DrawState)`: draws the building in a given state 
- `app/path.go`: paths (belts and pipes) struct and drawing code
    - `Draw(DrawState)`: draws the path in a given state (normal, new, selected, hovered, shadow, ...)
#
*app mode specific; update and draw*
- `app/newobj.go`: handles placing a new buildings
    - `Update()`: move/rotate the new object(s) to be placed, on click add them to the scene
    - `Draw()`: draw the new object(s) to be placed following the mouse 
- `app/selector.go`: handles creating an selection and hovering
    - `Update()`: find hovered object, create a selection (-> `selection.go`)
    - `Draw()`: draw the hovered object and the selection rectangle
- `app/selection.go`: handles the selection
    - `Update()`: manipulate the selection, delete, duplicate (-> `newobj.go`), drag and move/rotate
    - `Draw()`: draw the selected objects, the bounding box, the shadow of the original object when dragging
#
*other packages*
- `matrix`: 3x3 transform matrix (translation, rotation, scaling)
- `colors` package: color palette
- `math32` package: some math functions on `float32` (std `math` only supports `float64`)


Each of the update related files usually defines a main struct representing its topic state, 
and instanciate a global variable of that type, (eg: in `mouse.go`, there is a `Mouse` struct and a `mouse` global variable of that type).


As most of the code belongs to the `main`, we avoid cyclic dependencies and passing state around 
(eg: the `mouse` variable is available anywhere in the codebase) which makes for a consise code. 
But we also need extra care on how we update those variable state to keep the data flow reasonable.

Because the codebase is mainly a single package, every variable, functions, and types are accessible from anywhere.
I use exported indentifier as a indication that it is safe to access from outside its file
(exceptions includes the global variable).

## License

This project is licensed under the [GNU Affero General Public License v3.0](LICENSE).
