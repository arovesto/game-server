package entities

import (
	"encoding/json"
	"time"

	"github.com/arovesto/gio/elements"
	"github.com/arovesto/gio/event"
	"github.com/arovesto/gio/math"
)

const ControllerType = 11112

const helpDuration = time.Second * 10
const maxLevel = 5

type Controller struct {
	ID               int
	Level            int
	SnakesLen        int
	SnakesCnt        int
	PlayersMaxHP     float64
	SnakesHeadRadius float64
	Arena            math.Box
	Snakes           map[int]struct{}
	HelpCoolDown     time.Time
}

func NewController(id int, arena math.Box) *Controller {
	return &Controller{
		ID:               id,
		SnakesLen:        3,
                SnakesCnt:        1,
		PlayersMaxHP:     10,
		SnakesHeadRadius: 40,
		Arena:            arena,
		Snakes:           map[int]struct{}{},
	}
}

func (c *Controller) GetID() int {
	return c.ID
}

func (c *Controller) GetType() int {
	return ControllerType
}

func (c *Controller) GetState() ([]byte, error) {
	return json.Marshal(c)
}

func (c *Controller) SetState(bytes []byte) error {
	return json.Unmarshal(bytes, &c)
}

func (c *Controller) Move(duration time.Duration, processor elements.EventProcessor) error {
	players := processor.Players()
	if len(players) == 0 {
		return nil
	}
	for s := range c.Snakes {
		el := processor.GetElement(s)
		if el == nil {
			delete(c.Snakes, s)
		}
		if sn, ok := el.(*Snake); ok {
			if sn.Dead {
				delete(c.Snakes, s)
			}
		}
	}
	if len(c.Snakes) == 0 {
		if c.Level == maxLevel {
			return processor.ProcessEvent(event.Event{Type: "win", From: c.ID})
		}
		if c.Level != 0 {
			for _, p := range players {
				el := processor.GetElement(p)
				if el != nil {
					p, ok := el.(*Guy)
					if ok {
						p.HP = c.PlayersMaxHP
					}
				}
			}
		}
		c.Level++
		c.PlayersMaxHP++
		if c.Level%2 == 0 {
			c.SnakesLen++
		}
		if c.Level%5 == 0 {
			c.SnakesCnt++
		}
		if c.Level%3 == 0 {
			c.SnakesHeadRadius += 5
		}
		snakes := c.SnakesCnt * len(players)
		if snakes <= 0 {
			snakes = 1
		}
		if snakes > 10 {
			snakes = 10
		}
		for i := 0; i < snakes; i++ {
			spd := math.ClampF(math.RandomF(float64(c.SnakesLen*2), c.SnakesHeadRadius), 5, 20)
			id := processor.NewID()
			processor.NewElement(&Snake{
				Layer:    1,
				Orbs:     genOrbs(c.Arena, math.Random(c.SnakesLen-2, c.SnakesLen+2), c.SnakesHeadRadius),
				ID:       id,
				MaxSpeed: spd,
				MaxAngle: 0.05 * math.ClampF(float64(c.Level)/3, 1, 3),
				DoDamage: math.ClampF(math.RandomF(float64(c.SnakesLen)/2, math.ClampF(spd/5, float64(c.SnakesLen)/2, 3)), 0.5, 3),
			})
			c.Snakes[id] = struct{}{}
		}
	}
	for _, i := range players {
		el := processor.GetElement(i)
		if el != nil {
			p, ok := el.(*Guy)
			if ok && p.HP < 5 && time.Since(c.HelpCoolDown) > helpDuration {
				c.HelpCoolDown = time.Now()
				processor.NewElement(&Apple{
					ID:  processor.NewID(),
					Pos: math.Sphere{R: 30, Center: math.RandomInBox(math.Box{Corner: p.Position.Corner.Sub(math.Vector{X: 500, Y: 500}), Size: math.Vector{X: 1000, Y: 1000}})},
				})
			}
		}
	}
	return nil
}

func genOrbs(where math.Box, len int, rad float64) (r []math.Sphere) {
	if len <= 3 {
		len = 3
	}

	p := math.RandomInBox(where)

	for i := 0; i < len; i++ {
		if i == 0 {
			r = append(r, math.Sphere{R: rad, Center: p})
			p = p.Add(math.Vector{X: math.RandomF(-2*rad, 2*rad), Y: math.RandomF(-2*rad, 2*rad)})
		} else {
			r = append(r, math.Sphere{R: rad * 0.75, Center: p})
			p = p.Add(math.Vector{X: math.RandomF(-1.5*rad, 1.5*rad), Y: math.RandomF(-1.5*rad, 1.5*rad)})
		}
	}
	return
}

func init() {
	elements.GenElements[ControllerType] = func() elements.Element {
		return &Controller{}
	}
}
