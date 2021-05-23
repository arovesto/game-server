package elements

import (
	"encoding/json"
	"time"

	"github.com/arovesto/gio/canvas"
	"github.com/arovesto/gio/math"
)

// simple, mob
// use gio.GenElements[0] = func()Element{return &Mob{}} to enable
type Mob struct {
	ID        int
	Where     math.Box
	Spd       math.Vector
	Acc       math.Vector
	TextureID string
	Grounded  bool
}

func (s *Mob) GetState() ([]byte, error) {
	return json.Marshal(s)
}

func (s *Mob) SetState(bytes []byte) error {
	return json.Unmarshal(bytes, s)
}

func (s *Mob) Draw(c canvas.Canvas) {
	c.DrawShape(s.TextureID, s.Where, math.Box{Size: s.Where.Size})
}

func (s *Mob) Move(duration time.Duration, processor EventProcessor) error {
	s.Spd = s.Spd.Add(s.Acc)
	s.Where.Corner = s.Where.Corner.Add(s.Spd)
	if !s.Grounded {
		s.Acc.Y = 0.1
	}
	s.Grounded = false
	return nil
}

func (s *Mob) Collide(other Collidable) error {
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

func (s *Mob) GetID() int {
	return s.ID
}

func (s *Mob) GetType() int {
	return MobType
}
