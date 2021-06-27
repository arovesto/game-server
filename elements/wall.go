package elements

import (
	"encoding/json"
	"image/color"
	"time"

	"github.com/arovesto/gio/canvas"
	"github.com/arovesto/gio/math"
)

// simple, immovable wall
// use gio.GenElements[1] = func()Element{return &Wall{}} to enable
type Wall struct {
	ID      int
	Where   math.Box
	Texture string
}

func (s *Wall) GetState() ([]byte, error) {
	return json.Marshal(s)
}

func (s *Wall) SetState(bytes []byte) error {
	return json.Unmarshal(bytes, s)
}

func (s *Wall) Playable() bool {
	return false
}

func (s *Wall) Draw(c canvas.Canvas) {
	if s.Texture != "" {
		c.DrawColor(color.RGBAModel.Convert(color.Black), s.Where, math.Box{Size: s.Where.Size})
	}
}

func (s *Wall) Move(duration time.Duration, processor EventProcessor) error {
	return nil
}

func (s *Wall) Collide(other Collidable) error {
	return nil
}

func (s *Wall) Collider() math.Shape {
	return s.Where
}

func (s *Wall) GetID() int {
	return s.ID
}

func (s *Wall) GetType() int {
	return WallType
}
