package entities

import (
	"encoding/json"
	"image/color"
	"time"

	"github.com/arovesto/gio/canvas"
	"github.com/arovesto/gio/elements"
	"github.com/arovesto/gio/input"
	"github.com/arovesto/gio/math"
)

const SnakeType = 232323

const orbsGoBackConstant = 0.2
const dist = 5.0
const snakeDamageCoolDown = time.Second

type Snake struct {
	Orbs           []math.Sphere
	ID             int
	TargetID       int
	Layer          int
	TargetPosition math.Vector

	Vel      math.Vector
	Dead     bool
	MaxSpeed float64
	DoDamage float64
	MaxAngle float64

	DamageCoolDown time.Time
	Damaged        bool
}

func (s *Snake) Draw(c canvas.Canvas) {
	for i, o := range s.Orbs {
		el := math.Ellipse{Radius: math.Vector{X: o.R * 0.5, Y: o.R / 6}, Center: math.Vector{X: o.Center.X, Y: o.Center.Y + o.R}}
		c.DrawColor(color.RGBA{A: 10}, el, el)
		if i == 0 {
			if s.Damaged {
				c.DrawColor(color.RGBA{G: 200, R: 200, B: 200, A: 255}, o, o)
			} else {
				c.DrawColor(color.RGBA{G: 255, A: 255}, o, o)
			}
			eyeColor := color.RGBA{G: 250, R: 251, B: 255, A: 100}
			eyeLitColor := color.RGBA{G: 230, R: 237, B: 255, A: 100}

			eye := o.Center.Add(s.TargetPosition.Sub(o.Center).NormalizedTimes(o.R * 0.3))
			c.DrawColor(eyeColor, math.Sphere{Center: eye, R: 8}, math.Sphere{Center: eye, R: 8})
			c.DrawColor(eyeLitColor, math.Sphere{Center: eye, R: 3}, math.Sphere{Center: eye, R: 3})

			eye = o.Center.Add(s.TargetPosition.Sub(o.Center).NormalizedTimes(o.R * 0.8))
			c.DrawColor(eyeColor, math.Sphere{Center: eye, R: 8}, math.Sphere{Center: eye, R: 8})
			c.DrawColor(eyeLitColor, math.Sphere{Center: eye, R: 3}, math.Sphere{Center: eye, R: 3})
		} else {
			if s.Damaged {
				c.DrawColor(color.RGBA{G: 200, R: 200, B: 200, A: 255}, o, o)
			} else {
				c.DrawColor(color.RGBA{R: 255, A: 255}, o, o)
			}
		}
	}
}

func (s *Snake) GetID() int {
	return s.ID
}

func (s *Snake) GetType() int {
	return SnakeType
}

func (s *Snake) GetState() ([]byte, error) {
	return json.Marshal(s)
}

func (s *Snake) SetState(bytes []byte) error {
	return json.Unmarshal(bytes, s)
}

func (s *Snake) Collide(other elements.Collidable) error {
	return nil
}

func (s *Snake) Collider() math.Shape {
	return s.Orbs
}

func (s *Snake) Move(duration time.Duration, processor elements.EventProcessor) error {
	if time.Since(s.DamageCoolDown) > snakeDamageCoolDown && !s.Dead {
		s.Damaged = false
	}

	if s.Dead {
		return nil
	}

	t := processor.GetElement(s.TargetID)
	if t == nil || t.GetID() == s.ID {
		if len(processor.Players()) == 0 {
			return nil
		}
		s.TargetID = processor.Players()[0]
		return nil
	}
	c, ok := t.(elements.Collidable)
	if !ok {
		if len(processor.Players()) == 0 {
			return nil
		}
		s.TargetID = processor.Players()[0]
		return nil
	}
	s.TargetPosition = math.CenterOf(c.Collider())
	desiredVelocity := s.TargetPosition.Sub(s.Orbs[0].Center).NormalizedTimes(s.MaxSpeed)
	if s.Vel.X == 0 && s.Vel.Y == 0 {
		s.Vel = desiredVelocity
	} else {
		s.Vel = s.Vel.Rotate(math.ClampF(math.AngleBetween(s.Vel, desiredVelocity), -s.MaxAngle, s.MaxAngle))
	}
	s.Orbs[0].Center = s.Orbs[0].Center.Add(s.Vel)
	for i, o := range s.Orbs {
		if i != 0 {
			target := s.Orbs[i-1]
			target.Center = target.Center.Sub(target.Center.Sub(o.Center).NormalizedTimes(target.R + o.R + dist))
			s.Orbs[i].Center = o.Center.Add(target.Center.Sub(o.Center).Mul(orbsGoBackConstant))
		}
	}
	return nil
}

func (s *Snake) GetLayer() int {
	return s.Layer
}

func (s *Snake) IncreaseLength() {
	last := s.Orbs[len(s.Orbs)-1]
	s.Orbs = append(s.Orbs, math.Sphere{R: last.R, Center: last.Center.Add(s.Vel.NormalizedTimes(-(last.R*2 + dist)))})
}

func (s *Snake) Damage() {
	if time.Since(s.DamageCoolDown) > snakeDamageCoolDown {
		s.DamageCoolDown = time.Now()
		s.Damaged = true
		if len(s.Orbs) == 1 {
			s.Dead = true
		} else {
			s.Orbs = s.Orbs[:len(s.Orbs)-1]
		}
	}
}

const GameOverPlayerType = 123

type GameOverPlayer struct {
	elements.NoOpPlayer
	I     bool
	Lobby elements.EventProcessor
}

func (g *GameOverPlayer) SetInput(d []byte) error {
	return json.Unmarshal(d, &g.I)
}

func (g *GameOverPlayer) Input() ([]byte, error) {
	return json.Marshal(input.MousePressed)
}

func (g *GameOverPlayer) Move(d time.Duration, p elements.EventProcessor) error {
	if g.I {
		return p.Transfer(g.GetID(), g.Lobby)
	}
	return nil
}

func (g *GameOverPlayer) GetType() int {
	return GameOverPlayerType
}

func init() {
	elements.GenElements[SnakeType] = func() elements.Element {
		return &Snake{}
	}
	elements.GenElements[GameOverPlayerType] = func() elements.Element {
		return &GameOverPlayer{}
	}
}
