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

func (t *Triangle) NormalAt(p Vector) Vector {
	u, v, w := t.Barycentric(p)
	n := Vector{}
	n = n.Add(t.N1.MulScalar(u))
	n = n.Add(t.N2.MulScalar(v))
	n = n.Add(t.N3.MulScalar(w))
	n = n.Normalize()
	return n
}

func (t *Triangle) Barycentric(p Vector) (u, v, w float64) {
	v0 := t.V2.Sub(t.V1)
	v1 := t.V3.Sub(t.V1)
	v2 := p.Sub(t.V1)
	d00 := v0.Dot(v0)
	d01 := v0.Dot(v1)
	d11 := v1.Dot(v1)
	d20 := v2.Dot(v0)
	d21 := v2.Dot(v1)
	d := d00*d11 - d01*d01
	v = (d11*d20 - d01*d21) / d
	w = (d00*d21 - d01*d20) / d
	u = 1 - v - w
	return
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

func (t *Triangle) Rasterize() []Scanline {
	box := t.BoundingBox()
	min := box.Min.Floor()
	max := box.Max.Ceil()
	x1 := int(min.X)
	x2 := int(max.X)
	y1 := int(min.Y)
	y2 := int(max.Y)
	var lines []Scanline
	for y := y1; y <= y2; y++ {
		var lo, hi int
		var ok bool
		for x := x1; x <= x2; x++ {
			p := Vector{float64(x) + 0.5, float64(y) + 0.5, 0}
			b1 := (p.X-t.V2.X)*(t.V1.Y-t.V2.Y)-(t.V1.X-t.V2.X)*(p.Y-t.V2.Y) < 0
			b2 := (p.X-t.V3.X)*(t.V2.Y-t.V3.Y)-(t.V2.X-t.V3.X)*(p.Y-t.V3.Y) < 0
			b3 := (p.X-t.V1.X)*(t.V3.Y-t.V1.Y)-(t.V3.X-t.V1.X)*(p.Y-t.V1.Y) < 0
			if b1 == b2 && b2 == b3 {
				lo = x
				ok = true
				break
			}
		}
		if !ok {
			continue
		}
		for x := x2; x >= x1; x-- {
			p := Vector{float64(x) + 0.5, float64(y) + 0.5, 0}
			b1 := (p.X-t.V2.X)*(t.V1.Y-t.V2.Y)-(t.V1.X-t.V2.X)*(p.Y-t.V2.Y) < 0
			b2 := (p.X-t.V3.X)*(t.V2.Y-t.V3.Y)-(t.V2.X-t.V3.X)*(p.Y-t.V3.Y) < 0
			b3 := (p.X-t.V1.X)*(t.V3.Y-t.V1.Y)-(t.V3.X-t.V1.X)*(p.Y-t.V1.Y) < 0
			if b1 == b2 && b2 == b3 {
				hi = x
				break
			}
		}
		lines = append(lines, Scanline{y, lo, hi, 0xffff})
	}
	return lines
}
