package entities

import (
	"encoding/json"
	"image/color"
	"log"
	"time"

	"github.com/arovesto/gio/canvas"
	"github.com/arovesto/gio/elements"
	"github.com/arovesto/gio/math"
	"github.com/arovesto/gio/misc"
	"github.com/arovesto/gio/server"
)

const TriggerType = 1232

type Trigger struct {
	ID     int
	Gather math.Box
	Start  math.Box

	Ready    map[int]struct{}
	Starting bool
}

func (t *Trigger) GetLayer() int {
	return 7
}

func (t *Trigger) Draw(c canvas.Canvas) {
	c.DrawColor(color.RGBA{R: 200, G: 224, B: 110, A: 100}, t.Gather, t.Gather)
	c.DrawColor(color.RGBA{A: 255}, t.Start, t.Start)
	c.DrawText("Зайди в пещеру чтобы начать.", t.Gather.Corner, "32px serif")
	c.DrawText("Кто стоит на входе пойдет следом", t.Gather.Corner.Add(math.Vector{Y: 40}), "32px serif")
}

func (t *Trigger) Collide(other elements.Collidable) error {
	info := math.Collide(t.Gather, other.Collider())
	if info.Collided {
		t.Ready[other.GetID()] = struct{}{}
	} else {
		delete(t.Ready, other.GetID())
	}
	info = math.Collide(t.Start, other.Collider())
	if info.Collided {
		t.Starting = true
	}
	return nil
}

func (t *Trigger) Collider() math.Shape {
	return t.Gather
}

func (t *Trigger) GetID() int {
	return t.ID
}

func (t *Trigger) GetType() int {
	return TriggerType
}

func (t *Trigger) GetState() ([]byte, error) {
	return json.Marshal(t)
}

func (t *Trigger) SetState(bytes []byte) error {
	return json.Unmarshal(bytes, &t)
}

func (t *Trigger) Move(duration time.Duration, processor elements.EventProcessor) error {
	if t.Starting {
		t.Starting = false

		room := server.NewBasicRoom(misc.NewID(), "snake", []elements.Element{
			NewController(0, math.Box{Size: math.Vector{X: 256 * 15, Y: 144 * 15}}),
			&elements.StaticBackground{
				Where:     math.Box{Size: math.Vector{X: 256 * 15, Y: 144 * 15}},
				Texture:   math.Box{Size: math.Vector{X: 1280, Y: 720}},
				TextureID: "game-background.png",
				ID:        1,
			},
			&elements.Wall{
				ID:    2,
				Where: math.Box{Corner: math.Vector{X: 43 * 15, Y: 42 * 15}, Size: math.Vector{X: 10, Y: 60 * 15}},
			},
			&elements.Wall{
				ID:    3,
				Where: math.Box{Corner: math.Vector{X: 176 * 15, Y: 42 * 15}, Size: math.Vector{X: 10, Y: 60 * 15}},
			},
			&elements.Wall{
				ID:    4,
				Where: math.Box{Corner: math.Vector{X: 43 * 15, Y: 42 * 15}, Size: math.Vector{X: 130 * 15, Y: 10}},
			},
			&elements.Wall{
				ID:    5,
				Where: math.Box{Corner: math.Vector{X: 43 * 15, Y: 102 * 15}, Size: math.Vector{X: 130 * 15, Y: 10}},
			},
		})

		for pl := range t.Ready {
			id := room.NewID()
			room.NewElement(NewGuy(id, math.Vector{X: 1500, Y: 1000}))
			if err := processor.Transfer(pl, room); err != nil {
				log.Println("failed to transfer guy", pl, "to new room", err)
			}
		}
		t.Ready = map[int]struct{}{}
	}
	return nil
}

func init() {
	elements.GenElements[TriggerType] = func() elements.Element {
		return &Trigger{}
	}
}
