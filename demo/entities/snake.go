package entities

import (
	"encoding/json"
	"fmt"
	"image/color"
	"time"

	"github.com/arovesto/gio/canvas"
	"github.com/arovesto/gio/elements"
	"github.com/arovesto/gio/event"
	"github.com/arovesto/gio/input"
	"github.com/arovesto/gio/math"
	"github.com/arovesto/gio/misc"
)

const SnakeType = 232323

const orbsGoBackConstant = 0.2
const playerMoveAcc = 0.1
const playerMaxSpeed = 10
const dist = 10.0
const orbRadis = 25

type Snake struct {
	Orbs []math.Sphere
	ID   int

	Vel  math.Vector
	Dead bool

	I            PlayerInput
	Lost         bool
	LastCreated  time.Time
	MenusCreated int
}

func (s *Snake) Draw(c canvas.Canvas) {
	for i, o := range s.Orbs {
		if i == 0 {
			c.DrawColor(color.RGBA{G: 255, A: 255}, o, o)
		} else {
			c.DrawColor(color.RGBA{R: 255, A: 255}, o, o)
		}
	}
	c.DrawText(fmt.Sprintf("Создано %d менюшек", s.MenusCreated), s.Orbs[0].Center, "72px serif")
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

func (s *Snake) Input() ([]byte, error) {
	s.I.Target = input.MousePosition
	s.I.Moving = true
	s.I.GenNew = input.Pressed[input.KEY_SPACE]
	return json.Marshal(s.I)
}

func (s *Snake) SetInput(bytes []byte) error {
	return json.Unmarshal(bytes, &s.I)
}

func (s *Snake) Collide(other elements.Collidable) error {
	info := math.Collide(s.Orbs[0], other.Collider())
	if info.Collided {
		s.Dead = true
	}
	return nil
}

func (s *Snake) Collider() math.Shape {
	return s.Orbs
}

func (s *Snake) Move(duration time.Duration, processor elements.EventProcessor) error {
	//playerOrb := s.Orbs[0]

	if s.Dead && !s.Lost {
		s.Lost = true
		return processor.ProcessEvent(event.Event{Type: "lose", From: s.GetID()})
	}
	if s.I.GenNew && time.Since(s.LastCreated) > time.Millisecond*500 {
		s.LastCreated = time.Now()
		s.I.GenNew = false

		processor.NewElement(&elements.StaticBackground{
			Where:     math.Box{Corner: math.Vector{X: 100, Y: 100}, Size: math.Vector{X: 1464, Y: 720}},
			TextureID: "win.png",
			ID:        misc.NewID(),
		})
		s.MenusCreated++

		//var target math.Vector
		//if len(s.Orbs) >= 2 {
		//	prev := s.Orbs[len(s.Orbs)-1]
		//	subPrev := s.Orbs[len(s.Orbs)-2]
		//	target = prev.Center.Add(subPrev.Center.Sub(prev.Center).NormalizedTimes(-2*prev.R + dist))
		//} else {
		//	target = playerOrb.Center.Add(math.Vector{X: playerOrb.R, Y: playerOrb.R}).Add(math.Vector{X: dist, Y: dist})
		//	if s.Vel.Y != 0 && s.Vel.X != 0 {
		//		target = playerOrb.Center.Add(s.Vel.NormalizedTimes(-2*playerOrb.R + dist))
		//	}
		//}
		//s.Orbs = append(s.Orbs, math.Sphere{R: orbRadis, Center: target})
	}
	if s.I.Moving {
		s.Vel = math.Clamp(s.I.Target.Sub(s.Orbs[0].Center).Mul(playerMoveAcc), math.Vector{X: -playerMaxSpeed, Y: -playerMaxSpeed}, math.Vector{X: playerMaxSpeed, Y: playerMaxSpeed})
		s.Orbs[0].Center = s.Orbs[0].Center.Add(s.Vel)
		for i, o := range s.Orbs {
			if i != 0 {
				target := s.Orbs[i-1]
				target.Center = target.Center.Sub(target.Center.Sub(o.Center).NormalizedTimes(target.R + o.R + dist))
				s.Orbs[i].Center = o.Center.Add(target.Center.Sub(o.Center).Mul(orbsGoBackConstant))
			}
		}
	} else {
		s.Vel.X, s.Vel.Y = 0, 0
	}
	return nil
}

type PlayerInput struct {
	Target math.Vector
	GenNew bool
	Moving bool
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
