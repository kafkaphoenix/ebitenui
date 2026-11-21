//go:build !js
// +build !js

package jsUtil

import "github.com/ebitenui/ebitenui/utilities/mobile"

func IsMobileBrowser() bool {
	return false
}

func Prompt(_ mobile.InputMode, _ string, _ string, _ int, _ int, _ InsertCallBack, _ SelectAllCallback) {
}

func SetCursorPosition(_ int, _ int) {}

func GetCursorPosition() int {
	return 0
}
