package input

import (
	"syscall/js"
)

var Pressed = map[int]bool{}

func init() {
	window := js.Global()

	window.Call(
		"addEventListener",
		"keyup",
		js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			e := args[0]
			e.Call("preventDefault")
			Pressed[e.Get("keyCode").Int()] = false
			return nil
		}))

	window.Call(
		"addEventListener",
		"keydown",
		js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			e := args[0]
			e.Call("preventDefault")
			Pressed[e.Get("keyCode").Int()] = true
			return nil
		}))
}
