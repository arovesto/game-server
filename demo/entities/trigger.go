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
	return 5
}

func (t *Trigger) Draw(c canvas.Canvas) {
	c.DrawColor(color.RGBA{R: 209, G: 187, B: 19, A: 255}, t.Gather, t.Gather)
	c.DrawColor(color.RGBA{R: 255, A: 255}, t.Start, t.Start)
	c.DrawText("Нажми красную кнопку чтоб начать", t.Gather.Corner, "32px serif")
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
			NewController(0, math.Box{Corner: math.Vector{X: 100, Y: 100}, Size: math.Vector{X: 3000, Y: 3000}}),
		})

		for pl := range t.Ready {
			id := room.NewID()
			room.NewElement(NewGuy(id))
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
