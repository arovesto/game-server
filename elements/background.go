package elements

import (
	"encoding/json"

	"github.com/arovesto/gio/canvas"
	"github.com/arovesto/gio/math"
)

type StaticBackground struct {
	Where     math.Box
	Texture   math.Box
	TextureID string
	ID        int
	Layer     int
}

func (s *StaticBackground) Draw(c canvas.Canvas) {
	if s.Texture.Size.Abs() == 0 {
		s.Texture = math.Box{Size: s.Where.Size}
	}
	c.DrawShape(s.TextureID, s.Where, s.Texture)
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

func (s *StaticBackground) GetLayer() int {
	if s.Layer != 0 {
		return s.Layer
	}
	return 10
}
