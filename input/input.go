package input

import (
	"github.com/arovesto/gio/math"
)

var Pressed = map[int]bool{}
var MousePressed = false
var MousePosition math.Vector

func ResetPressed() {
	Pressed = map[int]bool{}
	MousePressed = false
}
