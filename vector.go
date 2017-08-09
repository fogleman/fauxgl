package fauxgl

import (
	"math"
	"math/rand"
)

type Vector struct {
	X, Y, Z float64
}

func V(x, y, z float64) Vector {
	return Vector{x, y, z}
}

func RandomUnitVector() Vector {
	for {
		x := rand.Float64()*2 - 1
		y := rand.Float64()*2 - 1
		z := rand.Float64()*2 - 1
		if x*x+y*y+z*z > 1 {
			continue
		}
		return Vector{x, y, z}.Normalize()
	}
}

func (a Vector) VectorW() VectorW {
	return VectorW{a.X, a.Y, a.Z, 1}
}

func (a Vector) IsDegenerate() bool {
	nan := math.IsNaN(a.X) || math.IsNaN(a.Y) || math.IsNaN(a.Z)
	inf := math.IsInf(a.X, 0) || math.IsInf(a.Y, 0) || math.IsInf(a.Z, 0)
	return nan || inf
}

func (a Vector) Length() float64 {
	return math.Sqrt(a.X*a.X + a.Y*a.Y + a.Z*a.Z)
}

func (a Vector) Less(b Vector) bool {
	if a.X != b.X {
		return a.X < b.X
	}
	if a.Y != b.Y {
		return a.Y < b.Y
	}
	return a.Z < b.Z
}

func (a Vector) Distance(b Vector) float64 {
	return a.Sub(b).Length()
}

func (a Vector) LengthSquared() float64 {
	return a.X*a.X + a.Y*a.Y + a.Z*a.Z
}

func (a Vector) DistanceSquared(b Vector) float64 {
	return a.Sub(b).LengthSquared()
}

func (a Vector) Lerp(b Vector, t float64) Vector {
	return a.Add(b.Sub(a).MulScalar(t))
}

func (a Vector) LerpDistance(b Vector, d float64) Vector {
	return a.Add(b.Sub(a).Normalize().MulScalar(d))
}

func (a Vector) Dot(b Vector) float64 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

func (a Vector) Cross(b Vector) Vector {
	x := a.Y*b.Z - a.Z*b.Y
	y := a.Z*b.X - a.X*b.Z
	z := a.X*b.Y - a.Y*b.X
	return Vector{x, y, z}
}

func (a Vector) Normalize() Vector {
	r := 1 / math.Sqrt(a.X*a.X+a.Y*a.Y+a.Z*a.Z)
	return Vector{a.X * r, a.Y * r, a.Z * r}
}

func (a Vector) Negate() Vector {
	return Vector{-a.X, -a.Y, -a.Z}
}

func (a Vector) Abs() Vector {
	return Vector{math.Abs(a.X), math.Abs(a.Y), math.Abs(a.Z)}
}

func (a Vector) Add(b Vector) Vector {
	return Vector{a.X + b.X, a.Y + b.Y, a.Z + b.Z}
}

func (a Vector) Sub(b Vector) Vector {
	return Vector{a.X - b.X, a.Y - b.Y, a.Z - b.Z}
}

func (a Vector) Mul(b Vector) Vector {
	return Vector{a.X * b.X, a.Y * b.Y, a.Z * b.Z}
}

func (a Vector) Div(b Vector) Vector {
	return Vector{a.X / b.X, a.Y / b.Y, a.Z / b.Z}
}

func (a Vector) Mod(b Vector) Vector {
	// as implemented in GLSL
	x := a.X - b.X*math.Floor(a.X/b.X)
	y := a.Y - b.Y*math.Floor(a.Y/b.Y)
	z := a.Z - b.Z*math.Floor(a.Z/b.Z)
	return Vector{x, y, z}
}

func (a Vector) AddScalar(b float64) Vector {
	return Vector{a.X + b, a.Y + b, a.Z + b}
}

func (a Vector) SubScalar(b float64) Vector {
	return Vector{a.X - b, a.Y - b, a.Z - b}
}

func (a Vector) MulScalar(b float64) Vector {
	return Vector{a.X * b, a.Y * b, a.Z * b}
}

func (a Vector) DivScalar(b float64) Vector {
	return Vector{a.X / b, a.Y / b, a.Z / b}
}

func (a Vector) Min(b Vector) Vector {
	return Vector{math.Min(a.X, b.X), math.Min(a.Y, b.Y), math.Min(a.Z, b.Z)}
}

func (a Vector) Max(b Vector) Vector {
	return Vector{math.Max(a.X, b.X), math.Max(a.Y, b.Y), math.Max(a.Z, b.Z)}
}

func (a Vector) Floor() Vector {
	return Vector{math.Floor(a.X), math.Floor(a.Y), math.Floor(a.Z)}
}

func (a Vector) Ceil() Vector {
	return Vector{math.Ceil(a.X), math.Ceil(a.Y), math.Ceil(a.Z)}
}

func (a Vector) Round() Vector {
	return a.RoundPlaces(0)
}

func (a Vector) RoundPlaces(n int) Vector {
	x := RoundPlaces(a.X, n)
	y := RoundPlaces(a.Y, n)
	z := RoundPlaces(a.Z, n)
	return Vector{x, y, z}
}

func (a Vector) MinComponent() float64 {
	return math.Min(math.Min(a.X, a.Y), a.Z)
}

func (a Vector) MaxComponent() float64 {
	return math.Max(math.Max(a.X, a.Y), a.Z)
}

func (i Vector) Reflect(n Vector) Vector {
	return i.Sub(n.MulScalar(2 * n.Dot(i)))
}

func (a Vector) Perpendicular() Vector {
	if a.X == 0 && a.Y == 0 {
		if a.Z == 0 {
			return Vector{}
		}
		return Vector{0, 1, 0}
	}
	return Vector{-a.Y, a.X, 0}.Normalize()
}

func (p Vector) SegmentDistance(v Vector, w Vector) float64 {
	l2 := v.DistanceSquared(w)
	if l2 == 0 {
		return p.Distance(v)
	}
	t := p.Sub(v).Dot(w.Sub(v)) / l2
	if t < 0 {
		return p.Distance(v)
	}
	if t > 1 {
		return p.Distance(w)
	}
	return v.Add(w.Sub(v).MulScalar(t)).Distance(p)
}

type VectorW struct {
	X, Y, Z, W float64
}

func (a VectorW) Vector() Vector {
	return Vector{a.X, a.Y, a.Z}
}

func (a VectorW) Outside() bool {
	x, y, z, w := a.X, a.Y, a.Z, a.W
	return x < -w || x > w || y < -w || y > w || z < -w || z > w
}

func (a VectorW) Dot(b VectorW) float64 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z + a.W*b.W
}

func (a VectorW) Add(b VectorW) VectorW {
	return VectorW{a.X + b.X, a.Y + b.Y, a.Z + b.Z, a.W + b.W}
}

func (a VectorW) Sub(b VectorW) VectorW {
	return VectorW{a.X - b.X, a.Y - b.Y, a.Z - b.Z, a.W - b.W}
}

func (a VectorW) MulScalar(b float64) VectorW {
	return VectorW{a.X * b, a.Y * b, a.Z * b, a.W * b}
}

func (a VectorW) DivScalar(b float64) VectorW {
	return VectorW{a.X / b, a.Y / b, a.Z / b, a.W / b}
}
