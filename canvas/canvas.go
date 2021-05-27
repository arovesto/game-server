package canvas

import (
	"image/color"

	"github.com/arovesto/gio/math"
)

type Canvas interface {
	DrawShape(id string, world, texture math.Box)
	DrawColor(cl color.Color, world, texture math.Shape)
	DrawText(text string, where math.Vector, font string)
}
