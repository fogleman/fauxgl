package fauxgl

import (
	"math"
)

type Mesh struct {
	Triangles []*Triangle
	box       *Box
}

func NewMesh(triangles []*Triangle) *Mesh {
	return &Mesh{triangles, nil}
}

func (m *Mesh) dirty() {
	m.box = nil
}

func (m *Mesh) Copy() *Mesh {
	triangles := make([]*Triangle, len(m.Triangles))
	for i, t := range m.Triangles {
		a := *t
		triangles[i] = &a
	}
	return NewMesh(triangles)
}

func (a *Mesh) Add(b *Mesh) {
	a.Triangles = append(a.Triangles, b.Triangles...)
	a.dirty()
}

func (m *Mesh) BoundingBox() Box {
	if m.box == nil {
		box := BoxForTriangles(m.Triangles)
		m.box = &box
	}
	return *m.box
}

func smoothNormalsThreshold(normal Vector, normals []Vector, threshold float64) Vector {
	result := Vector{}
	for _, x := range normals {
		if x.Dot(normal) >= threshold {
			result = result.Add(x)
		}
	}
	return result.Normalize()
}

func (m *Mesh) SmoothNormalsThreshold(radians float64) {
	threshold := math.Cos(radians)
	lookup := make(map[Vector][]Vector)
	for _, t := range m.Triangles {
		lookup[t.V1.Position] = append(lookup[t.V1.Position], t.V1.Normal)
		lookup[t.V2.Position] = append(lookup[t.V2.Position], t.V2.Normal)
		lookup[t.V3.Position] = append(lookup[t.V3.Position], t.V3.Normal)
	}
	for _, t := range m.Triangles {
		t.V1.Normal = smoothNormalsThreshold(t.V1.Normal, lookup[t.V1.Position], threshold)
		t.V2.Normal = smoothNormalsThreshold(t.V2.Normal, lookup[t.V2.Position], threshold)
		t.V3.Normal = smoothNormalsThreshold(t.V3.Normal, lookup[t.V3.Position], threshold)
	}
}

func (m *Mesh) SmoothNormals() {
	lookup := make(map[Vector]Vector)
	for _, t := range m.Triangles {
		lookup[t.V1.Position] = lookup[t.V1.Position].Add(t.V1.Normal)
		lookup[t.V2.Position] = lookup[t.V2.Position].Add(t.V2.Normal)
		lookup[t.V3.Position] = lookup[t.V3.Position].Add(t.V3.Normal)
	}
	for k, v := range lookup {
		lookup[k] = v.Normalize()
	}
	for _, t := range m.Triangles {
		t.V1.Normal = lookup[t.V1.Position]
		t.V2.Normal = lookup[t.V2.Position]
		t.V3.Normal = lookup[t.V3.Position]
	}
}

func (m *Mesh) UnitCube() {
	m.FitInside(Box{Vector{}, Vector{1, 1, 1}}, Vector{})
	m.MoveTo(Vector{}, Vector{0.5, 0.5, 0.5})
}

func (m *Mesh) BiUnitCube() {
	m.FitInside(Box{Vector{-1, -1, -1}, Vector{1, 1, 1}}, Vector{0.5, 0.5, 0.5})
}

func (m *Mesh) MoveTo(position, anchor Vector) {
	matrix := Translate(position.Sub(m.BoundingBox().Anchor(anchor)))
	m.Transform(matrix)
}

func (m *Mesh) FitInside(box Box, anchor Vector) {
	scale := box.Size().Div(m.BoundingBox().Size()).MinComponent()
	extra := box.Size().Sub(m.BoundingBox().Size().MulScalar(scale))
	matrix := Identity()
	matrix = matrix.Translate(m.BoundingBox().Min.Negate())
	matrix = matrix.Scale(Vector{scale, scale, scale})
	matrix = matrix.Translate(box.Min.Add(extra.Mul(anchor)))
	m.Transform(matrix)
}

func (m *Mesh) Transform(matrix Matrix) {
	for _, t := range m.Triangles {
		t.V1.Position = matrix.MulPosition(t.V1.Position)
		t.V2.Position = matrix.MulPosition(t.V2.Position)
		t.V3.Position = matrix.MulPosition(t.V3.Position)
		t.V1.Normal = matrix.MulDirection(t.V1.Normal)
		t.V2.Normal = matrix.MulDirection(t.V2.Normal)
		t.V3.Normal = matrix.MulDirection(t.V3.Normal)
	}
	m.dirty()
}

func (m *Mesh) SaveSTL(path string) error {
	return SaveSTL(path, m)
}
