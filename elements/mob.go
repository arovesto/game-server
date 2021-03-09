package elements

import (
	"github.com/arovesto/gio"
	"github.com/arovesto/gio/math"
)

// simple, mob
// use gio.GenElements["mob"] = func()Element{return &Mob{}} to enable
type Mob struct {
	Where     math.Box
	Spd       math.Vector
	Acc       math.Vector
	TextureID string
	Grounded  bool
}

func (s *Mob) Draw(c *gio.Canvas) {
	c.DrawShape(s.TextureID, s.Where, math.Box{Size: s.Where.Size})
}

func (s *Mob) Update() error {
	s.Spd = s.Spd.Add(s.Acc)
	s.Where.Corner = s.Where.Corner.Add(s.Spd)
	if !s.Grounded {
		s.Acc.Y = 0.1
	}
	s.Grounded = false
	return nil
}

func (s *Mob) Collide(other gio.Element) error {
	info := math.Collide(s.Where, other.Collider())
	s.Where.Corner = s.Where.Corner.Add(info.Delta)
	s.Grounded = info.TouchDown
	s.Acc = info.Clamp(s.Acc)
	s.Spd = info.Clamp(s.Spd)
	return nil
}

func (s *Mob) Collider() math.Shape {
	return s.Where
}
