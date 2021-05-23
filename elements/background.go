package elements

import (
	"encoding/json"

	"github.com/arovesto/gio/canvas"
	"github.com/arovesto/gio/math"
)

type StaticBackground struct {
	Where     math.Box
	TextureID string
	ID        int
}

func (s *StaticBackground) Draw(c canvas.Canvas) {
	c.DrawShape(s.TextureID, s.Where, math.Box{Size: s.Where.Size})
}

func (s *StaticBackground) GetID() int {
	return s.ID
}

func (s *StaticBackground) GetType() int {
	return StaticBackgroundType
}

func (s *StaticBackground) GetState() ([]byte, error) {
	return json.Marshal(s)
}

func (s *StaticBackground) SetState(bytes []byte) error {
	return json.Unmarshal(bytes, s)
}
