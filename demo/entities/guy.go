package entities

import (
	"encoding/json"
	"fmt"
	"image/color"
	math2 "math"
	"time"

	"github.com/arovesto/gio/canvas"
	"github.com/arovesto/gio/elements"
	"github.com/arovesto/gio/event"
	"github.com/arovesto/gio/input"
	"github.com/arovesto/gio/math"
)

const GuyType = 6543

const (
	GuyIdle = iota
	GuyMoveRight
	GuyMoveLeft
	GuyMoveUp
	GuyMoveDown
	GuyJump
)

const guyAnimDurationGo = time.Millisecond * 50
const guyAnimDurationIdle = time.Millisecond * 100
const guyAnimDurationJump = time.Millisecond * 50
const guyAnimAttackDuration = time.Millisecond * 100
const guyAnimAttackCoolDown = time.Millisecond * 1000

const guyMoveSpeed = 400.0
const guyFlySpeed = 400.0

const jumpSpeed = 700
const gravity = 1500

type GuyInput struct {
	Jump    bool
	Attack  bool
	MoveDir math.Vector
}

type Guy struct {
	ID                int
	TextureID         string
	SwordTextureID    string
	SwordTextureShape math.Box
	SwordPosition     math.Box
	Position          math.Box
	TextureShape      math.Box
	T                 time.Time
	AnimState         int
	JumpDirection     math.Vector
	Flying            bool
	Attacking         bool
	LastAttack        time.Time
	HP                float64
	DamageCoolDown    time.Time
	LastKnownPosition math.Box

	I GuyInput
}

func NewGuy(id int) *Guy {
	return &Guy{
		ID:                id,
		TextureID:         "guy.png",
		SwordTextureID:    "sword.png",
		SwordTextureShape: math.Box{Size: math.Vector{X: 32, Y: 32}},
		SwordPosition:     math.Box{Size: math.Vector{X: 64, Y: 64}},
		Position:          math.Box{Size: math.Vector{X: 128, Y: 128}},
		TextureShape:      math.Box{Size: math.Vector{X: 64, Y: 64}},
		HP:                10,
	}
}

func (g *Guy) Collide(other elements.Collidable) error {
	t, ok := other.(*Snake)
	if !ok {
		if _, ok := other.(*Trigger); ok {
			return nil
		}
		if _, ok := other.(*Guy); ok {
			return nil
		}
		info := math.Collide(g.Collider(), other.Collider())
		g.Position.Corner = g.Position.Corner.Add(info.Delta)
		return nil
	}
	if t.Dead {
		return nil
	}
	if g.Attacking {
		info := math.Collide(g.SwordPosition, t.Orbs)
		if info.Collided {
			t.Damage()
		}
	}
	info := math.Collide(math.Box{Corner: g.Position.Corner.Add(math.Vector{X: g.Position.Size.X * 0.25}), Size: g.Position.Size.Add(math.Vector{X: -g.Position.Size.X * 0.25})}, t.Orbs)
	if info.Collided && !g.Flying && !t.Damaged {
		if time.Since(g.DamageCoolDown) > time.Millisecond*1000 {
			g.HP -= t.DoDamage
			t.IncreaseLength()
			g.DamageCoolDown = time.Now()
		}
	}
	return nil
}

func (g *Guy) Collider() math.Shape {
	if g.Flying {
		return g.LastKnownPosition
	}
	return g.Position
}

func (g *Guy) PreDraw(c canvas.Canvas) {
	c.SetCameraCenter(g.Position.Corner)
}

func (g *Guy) Draw(c canvas.Canvas) {
	if g.Attacking {
		c.DrawShape(g.SwordTextureID, g.SwordPosition, g.SwordTextureShape)
	}

	groundY := g.Position.Corner.Y + g.Position.Size.Y
	if g.Flying {
		groundY = g.LastKnownPosition.Corner.Y + g.LastKnownPosition.Size.Y
	}
	rad := 20.0
	if g.Flying {
		rad = math.ClampF(math2.Abs(g.JumpDirection.Y)/25, 0, rad)
	}
	height := 10.0
	if g.Flying {
		height = math.ClampF(math2.Abs(g.JumpDirection.Y)/50, 0, height)
	}
	s := math.Ellipse{Radius: math.Vector{X: rad, Y: height}, Center: math.Vector{X: g.Position.Center().X - 5, Y: groundY}}
	c.DrawColor(color.RGBA{A: 10}, s, s)

	c.DrawShape(g.TextureID, g.Position, g.TextureShape)
	c.DrawText(fmt.Sprintf("HP = %.0f", g.HP), g.Position.Corner, "36px serif")
}

func (g *Guy) Move(duration time.Duration, processor elements.EventProcessor) error {
	if g.HP < 0 {
		g.HP = 0
		return processor.ProcessEvent(event.Event{Type: "lose", From: g.ID})
	}
	if g.HP == 0 {
		return nil
	}
	switch g.AnimState {
	case GuyIdle:
		if time.Since(g.T) > guyAnimDurationIdle {
			if g.TextureShape.Corner.X == 0 {
				g.TextureShape.Corner.X = g.TextureShape.Size.X * 8
			} else {
				g.TextureShape.Corner.X = 0
			}
			g.T = time.Now()
		}
	case GuyMoveRight, GuyMoveDown, GuyMoveLeft, GuyMoveUp:
		if time.Since(g.T) > guyAnimDurationGo {
			if g.TextureShape.Corner.X >= g.TextureShape.Size.X*7 {
				g.TextureShape.Corner.X = 0
			} else {
				g.TextureShape.Corner.X += g.TextureShape.Size.X
			}
			g.T = time.Now()
		}
	case GuyJump:
		if time.Since(g.T) > guyAnimDurationJump {
			g.TextureShape.Corner.X = 0
			g.JumpDirection = math.Vector{Y: -jumpSpeed}
			g.Flying = true
			g.AnimState = GuyIdle
			g.LastKnownPosition = g.Position
			g.T = time.Now()
		}
	}
	moveSpeed := guyMoveSpeed
	switch {
	case g.AnimState == GuyJump:
		moveSpeed = 0
	case g.I.Jump && !g.Flying:
		g.AnimState = GuyJump
		g.TextureShape.Corner.X = g.TextureShape.Size.X * 8
	case g.I.MoveDir.X > 0:
		g.AnimState = GuyMoveRight
		g.TextureShape.Corner.Y = 0
	case g.I.MoveDir.X < 0:
		g.AnimState = GuyMoveLeft
		g.TextureShape.Corner.Y = g.TextureShape.Size.Y
	case g.I.MoveDir.Y < 0:
		g.AnimState = GuyMoveUp
		g.TextureShape.Corner.Y = g.TextureShape.Size.Y * 3
	case g.I.MoveDir.Y > 0:
		g.AnimState = GuyMoveDown
		g.TextureShape.Corner.Y = g.TextureShape.Size.Y * 2
	default:
		g.AnimState = GuyIdle
	}

	if !g.Attacking && g.I.Attack && time.Since(g.LastAttack) > guyAnimAttackCoolDown {
		g.Attacking = true
		g.LastAttack = time.Now()
	}
	if g.Attacking && time.Since(g.LastAttack) > guyAnimAttackDuration {
		g.Attacking = false
		g.LastAttack = time.Now()
	}

	if g.Flying {
		moveSpeed = guyFlySpeed
		g.I.MoveDir.Y = 0
		delta := duration.Seconds() * gravity
		if g.LastKnownPosition.Corner.Y <= g.Position.Corner.Y && g.JumpDirection.Y > 0 {
			g.JumpDirection.Y = 0
			g.Position.Corner.Y = g.LastKnownPosition.Corner.Y
			g.Flying = false
		} else {
			g.JumpDirection.Y += delta
		}
	}
	g.Position.Corner = g.Position.Corner.Add(g.I.MoveDir.Mul(duration.Seconds() * moveSpeed)).Add(g.JumpDirection.Mul(duration.Seconds()))
	switch g.TextureShape.Corner.Y {
	case 0:
		g.SwordPosition.Corner = g.Position.Corner.Add(math.Vector{X: g.Position.Size.X * 0.75, Y: g.Position.Size.Y * 0.2})
		g.SwordTextureShape.Corner.X = 0
	case g.TextureShape.Size.Y:
		g.SwordPosition.Corner = g.Position.Corner.Add(math.Vector{X: g.Position.Size.X*0.25 - g.SwordPosition.Size.X, Y: g.Position.Size.Y * 0.2})
		g.SwordTextureShape.Corner.X = g.SwordTextureShape.Size.X
	case 2 * g.TextureShape.Size.Y:
		g.SwordPosition.Corner = g.Position.Corner.Add(math.Vector{X: g.Position.Size.X * 0.6, Y: g.Position.Size.Y * 0.65})
		g.SwordTextureShape.Corner.X = g.SwordTextureShape.Size.X * 2
	case 3 * g.TextureShape.Size.Y:
		g.SwordPosition.Corner = g.Position.Corner.Add(math.Vector{X: g.Position.Size.X*0.3 - g.SwordPosition.Size.X, Y: g.Position.Size.Y*0.65 - g.SwordPosition.Size.Y})
		g.SwordTextureShape.Corner.X = g.SwordTextureShape.Size.X * 3
	}

	return nil
}

func (g *Guy) GetID() int {
	return g.ID
}

func (g *Guy) GetType() int {
	return GuyType
}

func (g *Guy) GetState() ([]byte, error) {
	return json.Marshal(g)
}

func (g *Guy) SetState(bytes []byte) error {
	return json.Unmarshal(bytes, &g)
}

func (g *Guy) Input() ([]byte, error) {
	moveV := math.Vector{}
	switch {
	case input.Pressed[input.KEY_D]:
		moveV.X = 1
	case input.Pressed[input.KEY_A]:
		moveV.X = -1
	}
	switch {
	case input.Pressed[input.KEY_W]:
		moveV.Y = -1
	case input.Pressed[input.KEY_S]:
		moveV.Y = 1
	}
	return json.Marshal(GuyInput{
		Attack:  input.MousePressed,
		Jump:    input.Pressed[input.KEY_SPACE],
		MoveDir: moveV,
	})
}

func (g *Guy) SetInput(bytes []byte) error {
	return json.Unmarshal(bytes, &g.I)
}

func init() {
	elements.GenElements[GuyType] = func() elements.Element {
		return &Guy{}
	}
}
