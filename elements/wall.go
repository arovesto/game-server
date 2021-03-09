package elements

import (
	"github.com/arovesto/gio"
	"github.com/arovesto/gio/math"
	"image/color"
)

// simple, immovable wall
// use gio.GenElements["wall"] = func()Element{return &Wall{}} to enable
type Wall struct {
	Where   math.Box
	Texture string
}

func (s *Wall) Draw(c *gio.Canvas) {
	c.DrawColor(color.RGBAModel.Convert(color.Black), s.Where, math.Box{Size: s.Where.Size})
}

func (s *Wall) Update() error {
	return nil
}

func (s *Wall) Collide(other gio.Element) error {
	return nil
}

func (s *Wall) Collider() math.Shape {
	return s.Where
}
