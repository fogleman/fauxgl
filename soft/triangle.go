package soft

type Triangle struct {
	V1, V2, V3 Vector
	N1, N2, N3 Vector
}

func NewTriangle(v1, v2, v3 Vector) *Triangle {
	t := Triangle{}
	t.V1 = v1
	t.V2 = v2
	t.V3 = v3
	t.FixNormals()
	return &t
}

func (t *Triangle) BoundingBox() Box {
	min := t.V1.Min(t.V2).Min(t.V3)
	max := t.V1.Max(t.V2).Max(t.V3)
	return Box{min, max}
}

func (t *Triangle) Normal() Vector {
	e1 := t.V2.Sub(t.V1)
	e2 := t.V3.Sub(t.V1)
	return e1.Cross(e2).Normalize()
}

func (t *Triangle) BarycentricNormal(b Vector) Vector {
	n := Vector{}
	n = n.Add(t.N1.MulScalar(b.X))
	n = n.Add(t.N2.MulScalar(b.Y))
	n = n.Add(t.N3.MulScalar(b.Z))
	n = n.Normalize()
	return n
}

func (t *Triangle) Barycentric(p Vector) Vector {
	v0 := t.V2.Sub(t.V1)
	v1 := t.V3.Sub(t.V1)
	v2 := p.Sub(t.V1)
	d00 := v0.Dot(v0)
	d01 := v0.Dot(v1)
	d11 := v1.Dot(v1)
	d20 := v2.Dot(v0)
	d21 := v2.Dot(v1)
	d := d00*d11 - d01*d01
	v := (d11*d20 - d01*d21) / d
	w := (d00*d21 - d01*d20) / d
	u := 1 - v - w
	return Vector{u, v, w}
}

func (t *Triangle) FixNormals() {
	n := t.Normal()
	zero := Vector{}
	if t.N1 == zero {
		t.N1 = n
	}
	if t.N2 == zero {
		t.N2 = n
	}
	if t.N3 == zero {
		t.N3 = n
	}
}

func (t *Triangle) Rasterize() []Fragment {
	box := t.BoundingBox()
	min := box.Min.Floor()
	max := box.Max.Ceil()
	x1 := int(min.X)
	x2 := int(max.X)
	y1 := int(min.Y)
	y2 := int(max.Y)
	var fragments []Fragment
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			p := Vector{float64(x) + 0.5, float64(y) + 0.5, 0}
			b := t.Barycentric(p)
			if b.X < 0 || b.Y < 0 || b.Z < 0 {
				continue
			}
			z := b.X*t.V1.Z + b.Y*t.V2.Z + b.Z*t.V3.Z
			f := Fragment{Vector{float64(x), float64(y), z}, b}
			fragments = append(fragments, f)
		}
	}
	return fragments
}
