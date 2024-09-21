/* Test / demo file for tinyfiledialogs
 *
 * Run it with:
 *
 *   go test -v -count=1 -run ^[TestName]$ github.com/bonoboris/satisfied/tinyfiledialogs
 */

package tinyfiledialogs_test

import (
	"testing"

	tfd "github.com/bonoboris/satisfied/tinyfiledialogs"
)

func TestGetDialogBackend(t *testing.T) {
	isGraphic, backend := tfd.GetDialogBackend()
	t.Logf("isGraphic=%v backend=%v", isGraphic, backend)
}

func TestBeep(t *testing.T) {
	tfd.Beep()
}

func TestNotifyPopup(t *testing.T) {
	tfd.NotifyPopup("NotifyPopup", "This is a notification popup", tfd.IconInfo)
}

func TestMessageBoxOk(t *testing.T) {
	button := tfd.MessageBox(
		"MessageBox (DialogOk)",
		"This is a message box",
		tfd.DialogOk, tfd.IconInfo, tfd.ButtonOkYes)
	t.Logf("button=%v", button)
}

func TestMessageBoxOkCancel(t *testing.T) {
	button := tfd.MessageBox(
		"MessageBox (DialogOkCancel)",
		"This is a message box",
		tfd.DialogOkCancel, tfd.IconInfo, tfd.ButtonOkYes)
	t.Logf("button=%v", button)
}

func TestMessageBoxYesNo(t *testing.T) {
	button := tfd.MessageBox(
		"MessageBox (DialogYesNo)",
		"This is a message box",
		tfd.DialogYesNo, tfd.IconInfo, tfd.ButtonOkYes)
	t.Logf("button=%v", button)
}

func TestMessageBoxYesNoCancel(t *testing.T) {
	button := tfd.MessageBox(
		"MessageBox (DialogYesNoCancel)",
		"This is a message box",
		tfd.DialogYesNoCancel, tfd.IconError, tfd.ButtonNo)
	t.Logf("button=%v", button)
}

func TestMessageBoxWithQuote(t *testing.T) {
	button := tfd.MessageBox(
		"Foo'",
		"The title contains a quote which is somehow not allowed.",
		tfd.DialogOk, tfd.IconInfo, tfd.ButtonOkYes)
	t.Logf("button=%v", button)
}

func TestInputBox(t *testing.T) {
	input, ok := tfd.InputBox("InputBox", "Type something:", "placeholder")
	t.Logf("input=%v ok=%v", input, ok)
}

func TestPasswordBox(t *testing.T) {
	password, ok := tfd.PasswordBox("PasswordBox", "Type a password:")
	t.Logf("input=%v ok=%v", password, ok)
}

func TestSaveFileDialog(t *testing.T) {
	filename, ok := tfd.SaveFileDialog(
		"Save file as",
		"somefile.txt",
		[]string{},
		"")
	t.Logf("filename=%v ok=%v", filename, ok)
}

func TestOpenFileDialog(t *testing.T) {
	filename, ok := tfd.OpenFileDialog(
		"Open document",
		"C:/Users/boris/Pictures/goCC_ink.svg",
		[]string{},
		"")
	t.Logf("filename=%v ok=%v", filename, ok)
}

func TestOpenFileMultipleDialog(t *testing.T) {
	filenames, ok := tfd.OpenFileMultipleDialog(
		"Open files",
		"file1.txt file2.txt",
		[]string{},
		"")
	t.Logf("filenames=%v ok=%v", filenames, ok)
}

func TestSelectFolderDialog(t *testing.T) {
	dirname, ok := tfd.SelectFolderDialog("Select folder", "C:\\Users\\boris\\go\\")
	t.Logf("dirname=%v ok=%v", dirname, ok)
}
