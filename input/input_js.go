// +build js

package input

import (
	"syscall/js"

	"github.com/arovesto/gio/math"
)

func init() {
	window := js.Global()
	document := window.Get("document")
	body := document.Get("body")

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
	body.Set("onmousedown", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		MousePressed = true
		return nil
	}))
	body.Set("onmouseup", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		MousePressed = false
		return nil
	}))
	document.Set("onmousemove", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		MousePosition = math.Vector{X: args[0].Get("clientX").Float(), Y: args[0].Get("clientY").Float()}
		return nil
	}))
}
