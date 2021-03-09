package math

import (
	"image"
	"math"
)

type Vector struct {
	X float64
	Y float64
}

func (v Vector) ToPoint() image.Point {
	return image.Point{X: int(v.X), Y: int(v.Y)}
}

func (v Vector) Add	(other Vector) Vector {
	return Vector{v.X + other.X, v.Y + other.Y}
}

func (v Vector) Mul	(c float64) Vector {
	return Vector{v.X * c, v.Y * c}
}

func (v Vector) Sub	(other Vector) Vector {
	return Vector{v.X - other.X, v.Y - other.Y}
}

func (v Vector) Abs() float64 {
	return math.Abs(v.X) + math.Abs(v.Y)
}

type Box struct {
	Corner Vector
	Size   Vector
}

func (b Box) ToImageRect() image.Rectangle {
	second := b.Corner.Add(b.Size)
	return image.Rect(int(b.Corner.X), int(b.Corner.Y), int(second.X), int(second.Y))
}

type Sphere struct {
	Center Vector
	R      float64
}

type BackOff struct {
	Delta      Vector
	TouchLeft  bool
	TouchRight bool
	TouchUp    bool
	TouchDown  bool
}

type Shape interface {}

func BoxCollide(a, b Box) bool {
	aTop, aBottom, aLeft, aRight := a.Corner.Y, a.Corner.Y+a.Size.Y, a.Corner.X, a.Corner.X+a.Size.X
	bTop, bBottom, bLeft, bRight := b.Corner.Y, b.Corner.Y+b.Size.Y, b.Corner.X, b.Corner.X+b.Size.X
	return aBottom >= bTop && aTop <= bBottom && aRight >= bLeft && aLeft <= bRight
}

func Collide(a, b Shape) (r BackOff) {
	// assume every shape is a box for now
	aBox, bBox := a.(Box), b.(Box)
	if BoxCollide(aBox, bBox) {
		actions := []BackOff{
			{TouchDown: true, Delta: Vector{Y: bBox.Corner.Y - aBox.Corner.Y - aBox.Size.Y}},
			{TouchUp: true, Delta: Vector{Y: bBox.Corner.Y + bBox.Size.Y - aBox.Corner.Y}},
			{TouchLeft: true, Delta: Vector{X: bBox.Corner.X + bBox.Size.X - aBox.Corner.X}},
			{TouchRight: true, Delta: Vector{X: bBox.Corner.X - aBox.Corner.X - aBox.Size.X}},
		}
		r = actions[0]
		for _, a := range actions {
			if a.Delta.Abs() < r.Delta.Abs() {
				r = a
			}
		}
	}
	return
}

func (b BackOff) Clamp(v Vector) Vector {
	if (v.X < 0 && b.TouchLeft) || (v.X > 0 && b.TouchRight) {
		v.X = 0
	}
	if (v.Y < 0 && b.TouchUp) || (v.Y > 0 && b.TouchDown) {
		v.Y = 0
	}
	return v
}