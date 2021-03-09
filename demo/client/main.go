package main

import (
	"context"
	"github.com/arovesto/gio"
	"github.com/arovesto/gio/elements"
	"github.com/arovesto/gio/input"
	"github.com/arovesto/gio/math"
	"log"

	"nhooyr.io/websocket"
)

type Me struct {
	elements.Mob
}

func (m *Me) Update() error {
	switch {
	case input.Pressed[input.KEY_D]:
		m.Acc.X = 0.2
	case input.Pressed[input.KEY_A]:
		m.Acc.X = -0.2
	default:
		m.Acc.X = -m.Spd.X/3
	}

	if input.Pressed[input.KEY_W] {
		m.Spd.Y = -5
	}

	return m.Mob.Update()
}

func main() {
	c, _, err := websocket.Dial(context.Background(), "localhost:8080", nil)
	if err != nil {
		log.Println(err)
		return
	}
	r := gio.NewCanvas(gio.Config{Server: "localhost:8080", FPSCap: 60})
	if err := r.LoadTextures([]gio.Texture {
		{"tank.png", "tmp"},
	}); err != nil {
		panic(err)
	}

	m := Me {Mob: elements.Mob{
		Where:     math.Box{Size: math.Vector{X: 128, Y: 64}},
		TextureID: "tmp",
		Grounded:  false,
	},
	}

	w := &elements.Wall{
		Where:   math.Box{Corner: math.Vector{X: 200, Y: 200}, Size: math.Vector{X: 300, Y: 100}},
		Texture: "tmp",
	}

	r.Start(func(c *gio.Canvas) (done bool) {
		c.Clear()
		w.Draw(c)
		m.Draw(c)

		if input.Pressed[input.KEY_ESCAPE] {
			return true
		}
		if input.Pressed[input.KEY_SPACE] {
			r.MoveTo(m.Where.Corner.Add((m.Where.Size).Mul(0.5)))
		}
		if err := m.Update(); err != nil {
			log.Println(err)
			return true
		}
		if err := w.Update(); err != nil {
			log.Println(err)
			return true
		}
		if err := m.Collide(w); err != nil {
			log.Println(err)
			return true
		}
		return
	})

	r.Stop()
}
