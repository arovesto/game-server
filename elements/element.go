package elements

import (
	"time"

	"github.com/arovesto/gio/canvas"
	"github.com/arovesto/gio/event"
	"github.com/arovesto/gio/math"
)

// add to this map your element types, type shell be int
var GenElements = map[int]func() Element{
	NoOpPlayerType: func() Element {
		return &NoOpPlayer{}
	},
	MobType: func() Element {
		return &Mob{}
	},
	WallType: func() Element {
		return &Wall{}
	},
	StaticBackgroundType: func() Element {
		return &StaticBackground{}
	},
}

const (
	NoOpPlayerType = iota
	MobType
	WallType
	StaticBackgroundType
)

type EventProcessor interface {
	ProcessEvent(e event.Event) error
	Transfer(id int, r EventProcessor) error
	GetElements() map[int]Element
	GetElement(id int) Element
	Players() (r []int)
	NewElement(e Element)
	NewID() int
}

// TODO remove SetState, GetState, move it is to separate interface, use json-all by default
// should be json + yaml marshalled to allow sending to web + storing in config files
type Element interface {
	GetID() int   // for cross element actions
	GetType() int // for cross element actions
	GetState() ([]byte, error)
	SetState([]byte) error
}

type Drawable interface {
	Element
	Draw(c canvas.Canvas) // all actions to draw the object
}

type GetLayer interface {
	GetLayer() int
}

type Playable interface {
	Element
	Input() ([]byte, error)
	SetInput([]byte) error
}

type Collidable interface {
	Element
	Collide(other Collidable) error // collision should be checked inside
	Collider() math.Shape           // useful in Collide
}

type Movable interface {
	Element
	Move(duration time.Duration, processor EventProcessor) error // all changes in object
}

type PreDraw interface {
	PreDraw(c canvas.Canvas)
}
