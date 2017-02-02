package fauxgl

type Shape interface {
	Mesh() *Mesh
}

type Plane struct {
	Point    Vector
	Normal   Vector
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
	mesh.Transform(RotateTo(Vector{0, 0, 1}, p.Normal).Translate(p.Point))
	return mesh
}

type Cube struct {
	Position Vector
	Size     Vector
	Up       Vector
	Rotation float64
}

func NewCube(position, size, up Vector, rotation float64) *Cube {
	return &Cube{position, size, up, rotation}
}

func (c *Cube) Mesh() *Mesh {
	v := []Vector{
		{-1, -1, -1}, {-1, -1, 1}, {-1, 1, -1}, {-1, 1, 1},
		{1, -1, -1}, {1, -1, 1}, {1, 1, -1}, {1, 1, 1},
	}
	mesh := NewMesh([]*Triangle{
		NewTriangleForPoints(v[3], v[5], v[7]),
		NewTriangleForPoints(v[5], v[3], v[1]),
		NewTriangleForPoints(v[0], v[6], v[4]),
		NewTriangleForPoints(v[6], v[0], v[2]),
		NewTriangleForPoints(v[0], v[5], v[1]),
		NewTriangleForPoints(v[5], v[0], v[4]),
		NewTriangleForPoints(v[5], v[6], v[7]),
		NewTriangleForPoints(v[6], v[5], v[4]),
		NewTriangleForPoints(v[6], v[3], v[7]),
		NewTriangleForPoints(v[3], v[6], v[2]),
		NewTriangleForPoints(v[0], v[3], v[2]),
		NewTriangleForPoints(v[3], v[0], v[1]),
	})
	m := Rotate(Vector{0, 0, 1}, c.Rotation)
	m = m.Scale(c.Size.MulScalar(0.5))
	m = m.RotateTo(Vector{0, 0, 1}, c.Up)
	m = m.Translate(c.Position)
	mesh.Transform(m)
	return mesh
}
