package tinyfiledialogs

/*
#cgo windows LDFLAGS: -lcomdlg32 -lole32
#include <stdlib.h>
#include "tinyfiledialogs.h"
*/
import "C"

import (
	"strings"
	"unsafe"
)

////////////////////////////////////////////////////////////////////////////////////////////////////
// Types and constants
////////////////////////////////////////////////////////////////////////////////////////////////////

// special title used to query the dialog backend
const backendQuery = "tinyfd_query"

// DialogType represents the type of a dialog (determine the displayed buttons)
type DialogType string

const (
	// DialogOk represents a dialog with only an Ok button
	DialogOk DialogType = "ok"
	// DialogOkCancel represents a dialog with an Ok and a Cancel button
	DialogOkCancel DialogType = "okcancel"
	// DialogYesNo represents a dialog with an Yes and a No button
	DialogYesNo DialogType = "yesno"
	// DialogYesNoCancel represents a dialog with an Yes, a No and a Cancel button
	DialogYesNoCancel DialogType = "yesnocancel"
)

// Button represents a button in a dialog (meaning depends on the [DialogType])
type Button int

const (
	// Reprensents the button:
	//   - [DialogOkCancel] -> Cancel
	//   - [DialogYesNoCancel] -> Cancel
	//   - [DialogYesNo] -> No
	ButtonCancelNo Button = 0
	// Represents the button:
	//   - [DialogOk] -> Ok
	//   - [DialogOkCancel] -> Ok
	//   - [DialogYesNo] -> Yes
	//   - [DialogYesNoCancel] -> Yes
	ButtonOkYes Button = 1
	// Represents the button:
	//   - [DialogYesNoCancel] -> No
	ButtonNo Button = 2
)

func (b Button) String() string {
	switch b {
	case ButtonCancelNo:
		return "Cancel/No"
	case ButtonOkYes:
		return "Ok/Yes"
	case ButtonNo:
		return "No"
	default:
		return "Invalid button value"
	}
}

type IconType string

const (
	IconInfo     IconType = "info"
	IconWarning  IconType = "warning"
	IconError    IconType = "error"
	IconQuestion IconType = "question"
)

// RGB represents a color in RGB format.
type RGB = [3]uint8

// crgb represents a color in RGB format.
type crgb = [3]C.uchar

////////////////////////////////////////////////////////////////////////////////////////////////////
// cgo helpers
////////////////////////////////////////////////////////////////////////////////////////////////////

// cstr casts a go string to a C string pointer and returns a function to free it.
//
// if s is empty, returns nil, noop function.
func cstr(s string) (*C.char, func()) {
	if s == "" {
		return nil, func() {}
	}
	c := C.CString(s)
	return c, func() { C.free(unsafe.Pointer(c)) }
}

// gostr casts a C string pointer to a go string, and returns whether it was not a nil pointer.
//
// Note: We don't want to free the C pointer, as it is a static buffer owned by the C code.
func gostr(c *C.char) (string, bool) {
	if c == nil {
		return "", false
	}
	return C.GoString(c), true
}

// cstrArray casts a Go []string to a C array of C string pointers and returns a function to free it.
func cstrArray(sa []string) (**C.char, func()) {
	n := len(sa)
	if n == 0 {
		return nil, func() {}
	}

	cArray := C.malloc(C.size_t(n) * C.size_t(unsafe.Sizeof(uintptr(0))))
	cStrings := (*[1 << 30]*C.char)(cArray)[:n:n]

	for i, s := range sa {
		cStr := C.CString(s)
		cStrings[i] = cStr
	}

	return (**C.char)(cArray), func() {
		// Call all the individual free functions for the C strings.
		for _, cStr := range cStrings {
			C.free(unsafe.Pointer(cStr))
		}
		// Free the array itself.
		C.free(unsafe.Pointer(cArray))
	}
}

// cRGB casts a RGB to a C array of uchar.
func cRGB(rgb RGB) crgb {
	return crgb{C.uchar(rgb[0]), C.uchar(rgb[1]), C.uchar(rgb[2])}
}

// goRGB casts a C array of uchar to a RGB.
func goRGB(cRgb crgb) RGB {
	return RGB{uint8(cRgb[0]), uint8(cRgb[1]), uint8(cRgb[2])}
}


////////////////////////////////////////////////////////////////////////////////////////////////////
// Bindings
////////////////////////////////////////////////////////////////////////////////////////////////////

// GetDialogBackend returns whether the dialog backend is graphical and the backend name.
func GetDialogBackend() (isGraphic bool, backend string) {
	cBackendQuery := C.CString(backendQuery)
	defer C.free(unsafe.Pointer(cBackendQuery))
	isGraphic = C.tinyfd_notifyPopup(cBackendQuery, nil, nil) != 0
	backend = C.GoString(&C.tinyfd_response[0])
	return isGraphic, backend
}

// Beep plays the system beep sound.
func Beep() {
	C.tinyfd_beep()
}

const invalidTitleMsg = `Cannot use tinyfd_query as a title.

This title is reserved for tinyfiledialogs internal usage.

If you want to query the dialog backend, use the function: GetDialogBackend()
`

// NotifyPopup displays a popup notification.
func NotifyPopup(title, message string, iconType IconType) {
	if title == backendQuery {
		NotifyPopup("Invalid title", invalidTitleMsg, IconError)
		return
	}
	cTitle, freeTitle := cstr(title)
	defer freeTitle()
	cMessage, freeMessage := cstr(message)
	defer freeMessage()
	cIconType, freeIconType := cstr(string(iconType))
	defer freeIconType()
	C.tinyfd_notifyPopup(cTitle, cMessage, cIconType)
}

// MessageBox displays a message box and returns the [Button] that was clicked.
//
// Note: closing the message box with the escape Key / X button:
//   - returns [ButtonOkYes] if dialofType is [DialogOk]
//   - is disabled if dialogType is [DialogYesNo]
//   - returns [ButtonCancelNo] if dialogType is [DialogOkCancel] or [DialogYesNoCancel]
func MessageBox(
	title, message string,
	dialogType DialogType, iconType IconType,
	defaultButton Button,
) Button {
	if title == backendQuery {
		NotifyPopup("Invalid title", invalidTitleMsg, IconError)
		return defaultButton
	}

	cTitle, freeTitle := cstr(title)
	defer freeTitle()
	cMessage, freeMessage := cstr(message)
	defer freeMessage()
	cDialogType, freeDialogType := cstr(string(dialogType))
	defer freeDialogType()
	cIconType, freeIconType := cstr(string(iconType))
	defer freeIconType()

	cRet := C.tinyfd_messageBox(cTitle, cMessage, cDialogType, cIconType, C.int(defaultButton))
	return Button(cRet)
}

// InputBox displays a input box,
// and returns either (input, true) or ("", false) on cancel.
func InputBox(title, message, defaultInput string) (string, bool) {
	if title == backendQuery {
		NotifyPopup("Invalid title", invalidTitleMsg, IconError)
		return "", false
	}

	cTitle, freeTitle := cstr(title)
	defer freeTitle()
	cMessage, freeMessage := cstr(message)
	defer freeMessage()
	cDefaultInput := C.CString(defaultInput) // We don't want to cast empty string to nil here
	defer C.free(unsafe.Pointer(cDefaultInput))

	cRet := C.tinyfd_inputBox(cTitle, cMessage, cDefaultInput)
	return gostr(cRet)
}

// PasswordBox displays a password input box,
// and returns either (password, true) or ("", false) on cancel.
//
// Note: this use tinyfd_inputBox, with a NULL pointer for the default input.
func PasswordBox(title, message string) (string, bool) {
	if title == backendQuery {
		NotifyPopup("Invalid title", invalidTitleMsg, IconError)
		return "", false
	}

	cTitle, freeTitle := cstr(title)
	defer freeTitle()
	cMessage, freeMessage := cstr(message)
	defer freeMessage()

	cRet := C.tinyfd_inputBox(cTitle, cMessage, nil)
	return gostr(cRet)
}

// SaveFileDialog displays a save file dialog,
// and returns either (filename, true) or ("", false) on cancel.
//
//   - defaultPath: can end with "/" to only set a default directory;
//     can be a file name only, in which case the backend will select the default directory
//   - filePatterns: a list of file patterns to filter the displayed files.
//   - filterDescription: a description for the filePatterns; ignored if filePatterns is empty.
//
// Note: tinyfiledialogs will add an "All File" entry if filePatterns is not empty, but will not add
// anything if filePatterns is empty, making for an empty combobox.
//
// Example:
//
//	filename, ok := SaveFileDialog("Save document as", "somefile.txt", []string{"*.txt"}, "Text files")
func SaveFileDialog(
	title, defaultPath string,
	filePatterns []string, filterDescription string,
) (string, bool) {
	if title == backendQuery {
		NotifyPopup("Invalid title", invalidTitleMsg, IconError)
		return "", false
	}

	cTitle, freeTitle := cstr(title)
	defer freeTitle()
	cDefaultPath, freeDefaultPath := cstr(defaultPath)
	defer freeDefaultPath()
	cFilePatterns, freeFilePatterns := cstrArray(filePatterns)
	defer freeFilePatterns()
	cFilterDescription := C.CString(filterDescription)
	defer C.free(unsafe.Pointer(cFilterDescription))

	cRet := C.tinyfd_saveFileDialog(
		cTitle, cDefaultPath,
		C.int(len(filePatterns)), cFilePatterns,
		cFilterDescription)

	return gostr(cRet)
}

// OpenFileDialog displays a open file dialog, and returns either (filename, true) or ("", false) on cancel.
//
//   - defaultPath: can end with "/" to only set a default directory.
//     can be a file name only, in which case the backend will select the default directory
//   - filePatterns: a list of file patterns to filter the displayed files.
//   - filterDescription: a description for the filePatterns; ignored if filePatterns is empty.
//
// Note: tinyfiledialogs will add an "All File" entry if filePatterns is not empty, but will not add
// anything if filePatterns is empty, making for an empty combobox.
//
// Example:
//
//	filename, ok := OpenFileDialog("Open document", "somefile.txt", []string{"*.txt"}, "Text files")
func OpenFileDialog(
	title, defaultPath string,
	filePatterns []string, filterDescription string,
) (string, bool) {
	if title == backendQuery {
		NotifyPopup("Invalid title", invalidTitleMsg, IconError)
		return "", false
	}

	cTitle, freeTitle := cstr(title)
	defer freeTitle()
	cDefaultPath, freeDefaultPath := cstr(defaultPath)
	defer freeDefaultPath()
	cFilePatterns, freeFilePatterns := cstrArray(filePatterns)
	defer freeFilePatterns()
	cFilterDescription, freeFilterDescription := cstr(filterDescription)
	defer freeFilterDescription()

	cRet := C.tinyfd_openFileDialog(
		cTitle, cDefaultPath,
		C.int(len(filePatterns)), cFilePatterns,
		cFilterDescription, 0)

	return gostr(cRet)
}

// OpenFileMultipleDialog displays a open file(s) dialog,
// and returns either (filenames, true) or ([], false) on cancel.
//
//   - defaultPath: can end with "/" to only set a default directory.
//     can be a file name only, in which case the backend will select the default directory
//   - filePatterns: a list of file patterns to filter the displayed files.
//   - filterDescription: a description for the filePatterns; ignored if filePatterns is empty.
//
// Note: tinyfiledialogs will add an "All File" entry if filePatterns is not empty, but will not add
// anything if filePatterns is empty, making for an empty combobox.
//
// Example:
//
//	filenames, ok := OpenFileMultipleDialog("Open documents", "", []string{"*.txt"}, "Text files")
func OpenFileMultipleDialog(
	title, defaultPath string,
	filePatterns []string, filterDescription string,
) ([]string, bool) {
	if title == backendQuery {
		NotifyPopup("Invalid title", invalidTitleMsg, IconError)
		return []string{}, false
	}

	cTitle, freeTitle := cstr(title)
	defer freeTitle()
	cDefaultPath, freeDefaultPath := cstr(defaultPath)
	defer freeDefaultPath()
	cFilePatterns, freeFilePatterns := cstrArray(filePatterns)
	defer freeFilePatterns()
	cFilterDescription, freeFilterDescription := cstr(filterDescription)
	defer freeFilterDescription()

	cRet := C.tinyfd_openFileDialog(
		cTitle, cDefaultPath,
		C.int(len(filePatterns)), cFilePatterns,
		cFilterDescription, 1)

	ret, ok := gostr(cRet)
	return strings.Split(ret, "|"), ok
}

// SelectFolderDialog displays a select folder dialog,
//
//   - defaultPath: must end with "/"
//
// Example:
//
//	defaultPath := path.Join(os.UserHomeDir(), "MyApp") + "/"
//	dirname, ok := SelectFolderDialog("Open files", defaultPath)
func SelectFolderDialog(title, defaultPath string) (string, bool) {
	if title == backendQuery {
		NotifyPopup("Invalid title", invalidTitleMsg, IconError)
		return "", false
	}

	cTitle, freeTitle := cstr(title)
	defer freeTitle()
	cDefaultPath, freeDefaultPath := cstr(defaultPath)
	defer freeDefaultPath()

	cRet := C.tinyfd_selectFolderDialog(cTitle, cDefaultPath)
	return gostr(cRet)
}

// ColorChooserRGB displays a color picker dialog,
// and returns either (color, true) or (RGB{}, false) on cancel.
//
// Note: this use tinyfd_colorChooser.
func ColorChooserRGB(title string, defaultRgb RGB) (RGB, bool) {
	if title == backendQuery {
		NotifyPopup("Invalid title", invalidTitleMsg, IconError)
		return RGB{}, false
	}

	cTitle, freeTitle := cstr(title)
	defer freeTitle()

	cDefaultRgb := cRGB(defaultRgb)
	var cResultRgb crgb
	cRet := C.tinyfd_colorChooser(cTitle, nil, &cDefaultRgb[0], &cResultRgb[0])
	if cRet == nil {
		return RGB{}, false
	}
	return goRGB(cResultRgb), true
}

// ColorChooserHex displays a color picker dialog,
// and returns either (hexColor, true) or ("", false) on cancel.
//
// Note: this use tinyfd_colorChooser.
func ColorChooserHex(title, defaultHex string) (string, bool) {
	if title == backendQuery {
		NotifyPopup("Invalid title", invalidTitleMsg, IconError)
		return "", false
	}

	cTitle, freeTitle := cstr(title)
	defer freeTitle()

	cDefaultHex, freeDefaultHex := cstr(defaultHex)
	defer freeDefaultHex()

	var cDefaultRgb, cResultRgb crgb

	cRet := C.tinyfd_colorChooser(cTitle, cDefaultHex, &cDefaultRgb[0], &cResultRgb[0])
	return gostr(cRet)
}
