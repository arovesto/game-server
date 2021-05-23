package canvas

import (
	"image/color"

	"github.com/arovesto/gio/math"
)

type Canvas interface {
	DrawShape(id string, world, texture math.Shape)
	DrawColor(cl color.Color, world, texture math.Shape)
}
