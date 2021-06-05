package math

import (
	"image"
	"log"
	"math"
	"math/rand"
)

type Vector struct {
	X float64
	Y float64
}

func (v Vector) ToPoint() image.Point {
	return image.Point{X: int(v.X), Y: int(v.Y)}
}

func (v Vector) Add(other Vector) Vector {
	return Vector{v.X + other.X, v.Y + other.Y}
}

func (v Vector) Rotate(angle float64) Vector {
	return Vector{
		X: v.X*math.Cos(angle) - v.Y*math.Sin(angle),
		Y: v.X*math.Sin(angle) + v.Y*math.Cos(angle),
	}
}

func (v Vector) Mul(c float64) Vector {
	return Vector{v.X * c, v.Y * c}
}

func (v Vector) Sub(other Vector) Vector {
	return Vector{v.X - other.X, v.Y - other.Y}
}

func (v Vector) Abs() float64 {
	return math.Abs(v.X) + math.Abs(v.Y)
}

func (v Vector) SquaredL() float64 {
	return v.X*v.X + v.Y*v.Y
}

func (v Vector) NormalizedTimes(c float64) Vector {
	return v.Mul(c / v.Len())
}

func (v Vector) Len() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func SquaredEuclideanDistance(a, b Vector) float64 {
	s := a.Sub(b)
	return s.X*s.X + s.Y*s.Y
}

type Box struct {
	Corner Vector
	Size   Vector
}

func (b Box) ToImageRect() image.Rectangle {
	second := b.Corner.Add(b.Size)
	return image.Rect(int(b.Corner.X), int(b.Corner.Y), int(second.X), int(second.Y))
}

func (b Box) Center() Vector {
	return b.Corner.Add(b.Size.Mul(0.5))
}

func (b Box) IsInside(v Vector) bool {
	return v.X >= b.Corner.X && v.X <= b.Corner.X+b.Size.X && v.Y >= b.Corner.Y && v.Y <= b.Corner.Y+b.Size.Y
}

func ClampF(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func Clamp(val, min, max Vector) Vector {
	return Vector{X: ClampF(val.X, min.X, max.X), Y: ClampF(val.Y, min.Y, max.Y)}
}

type Sphere struct {
	Center Vector
	R      float64
}

type Ellipse struct {
	Center Vector
	Radius Vector
}

func (s Sphere) IsInside(v Vector) bool {
	return v.Sub(s.Center).SquaredL() < s.R*s.R
}

type BackOff struct {
	Delta Vector
	// TODO introduce ContactPoint Vector point or smth instead of this
	TouchLeft  bool
	TouchRight bool
	TouchUp    bool
	TouchDown  bool

	Collided bool
}

func CenterOf(s Shape) Vector {
	switch sVal := s.(type) {
	case Box:
		return sVal.Center()
	case Sphere:
		return sVal.Center
	default:
		log.Panicln("not implemented CenterOf", sVal)
		return Vector{}
	}
}

type Shape interface{}

func BoxCollide(a, b Box) bool {
	aTop, aBottom, aLeft, aRight := a.Corner.Y, a.Corner.Y+a.Size.Y, a.Corner.X, a.Corner.X+a.Size.X
	bTop, bBottom, bLeft, bRight := b.Corner.Y, b.Corner.Y+b.Size.Y, b.Corner.X, b.Corner.X+b.Size.X
	return aBottom >= bTop && aTop <= bBottom && aRight >= bLeft && aLeft <= bRight
}

func BoxSphereCollide(a Box, b Sphere) bool {
	return Clamp(b.Center, a.Corner, a.Corner.Add(a.Size)).Sub(b.Center).SquaredL() < b.R*b.R
}

func SphereCollide(a, b Sphere) bool {
	return SquaredEuclideanDistance(a.Center, b.Center) < (a.R+b.R)*(a.R+b.R)
}

// TODO implement BackOff vector for Sphere - Box and Sphere- Sphere, now it is none
// TODO For Sphere - Sphere backoff is always along the center-center line, so 2 basic variants
func Collide(a, b Shape) (r BackOff) {
	var actions []BackOff
	switch aVal := a.(type) {
	case Box:
		switch bVal := b.(type) {
		case Box:
			if BoxCollide(aVal, bVal) {
				actions = []BackOff{
					{TouchDown: true, Delta: Vector{Y: bVal.Corner.Y - aVal.Corner.Y - aVal.Size.Y}, Collided: true},
					{TouchUp: true, Delta: Vector{Y: bVal.Corner.Y + bVal.Size.Y - aVal.Corner.Y}, Collided: true},
					{TouchLeft: true, Delta: Vector{X: bVal.Corner.X + bVal.Size.X - aVal.Corner.X}, Collided: true},
					{TouchRight: true, Delta: Vector{X: bVal.Corner.X - aVal.Corner.X - aVal.Size.X}, Collided: true},
				}
			}
		case Sphere:
			if BoxSphereCollide(aVal, bVal) {
				return BackOff{Collided: true} // TODO to implement true "Contact" point (more at BackOff)
			}
		case []Sphere:
			for _, b := range bVal {
				if BoxSphereCollide(aVal, b) {
					return BackOff{Collided: true}
				}
			}
		default:
			log.Panicln("unknown second val", bVal)
		}
	case Sphere:
		switch bVal := b.(type) {
		case Box:
			if BoxSphereCollide(bVal, aVal) {
				return BackOff{Collided: true} // TODO to implement true "Contact" point (more at BackOff)
			}
		case Sphere:
			if SphereCollide(aVal, bVal) {
				return BackOff{Collided: true} // TODO to implement true "Contact" point (more at BackOff)
			}
		case []Sphere:
			for _, b := range bVal {
				if SphereCollide(aVal, b) {
					return BackOff{Collided: true}
				}
			}
		default:
			log.Panicln("unknown second val", bVal)
		}
	case []Sphere:
		switch bVal := b.(type) {
		case []Sphere:
			for _, a := range aVal {
				for _, b := range bVal {
					if SphereCollide(a, b) {
						return BackOff{Collided: true}
					}
				}
			}
		case Sphere:
			for _, a := range aVal {
				if SphereCollide(a, bVal) {
					return BackOff{Collided: true}
				}
			}
		default:
			log.Panicln("unknown second val", bVal)
		}
	default:
		log.Panicln("unknown val", aVal)
	}
	if len(actions) == 0 {
		return
	}
	r = actions[0]
	for _, a := range actions {
		if a.Delta.Abs() < r.Delta.Abs() {
			r = a
		}
	}

	// assume every shape is a box for now

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

func AngleOn(from, to Vector) float64 {
	if to.X == from.X {
		return 0
	}
	return math.Atan((to.Y - from.Y) / math.Abs(to.X-from.X))
}

func AngleBetween(from, to Vector) float64 {
	return math.Atan2(from.X*to.Y-from.Y*to.X, from.X*to.X+from.Y*to.Y)
}

func Random(min, max int) int {
	return rand.Intn(max-min) + min
}

func RandomF(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func RandomInBox(b Box) Vector {
	return Vector{
		X: RandomF(b.Corner.X, b.Corner.X+b.Size.X),
		Y: RandomF(b.Corner.Y, b.Corner.Y+b.Size.Y),
	}
}
