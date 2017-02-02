package fauxgl

import "math"

type Shape interface {
	Mesh() *Mesh
}

type Plane struct {
	P, N     Vector
	HalfSize float64
}

func NewPlane(p, n Vector, s float64) *Plane {
	return &Plane{p, n, s}
}

func (p *Plane) Mesh() *Mesh {
	s := p.HalfSize
	v1 := Vector{-s, -s, 0}
	v2 := Vector{s, -s, 0}
	v3 := Vector{s, s, 0}
	v4 := Vector{-s, s, 0}
	mesh := NewMesh([]*Triangle{
		NewTriangleForPoints(v1, v2, v3),
		NewTriangleForPoints(v1, v3, v4),
	})
	up := Vector{0, 0, 1}
	dot := p.N.Dot(up)
	if dot == 1 {
		// no rotation needed
	} else if dot == -1 {
		// rotate 180
		mesh.Transform(Rotate(Vector{1, 0, 0}, math.Pi))
	} else {
		a := math.Acos(dot)
		v := p.N.Cross(up).Normalize()
		mesh.Transform(Rotate(v, a))
	}
	mesh.Transform(Translate(p.P))
	return mesh
}
