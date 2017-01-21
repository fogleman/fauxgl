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

func (t *Triangle) Area() float64 {
	e1 := t.V2.Sub(t.V1)
	e2 := t.V3.Sub(t.V1)
	n := e1.Cross(e2)
	return n.Length() / 2
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

func (t *Triangle) IsClockwise() bool {
	var sum float64
	sum += (t.V2.X - t.V1.X) * (t.V2.Y + t.V1.Y)
	sum += (t.V3.X - t.V2.X) * (t.V3.Y + t.V2.Y)
	sum += (t.V1.X - t.V3.X) * (t.V1.Y + t.V3.Y)
	return sum >= 0
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
		var previous uint32
		for x := x1; x <= x2; x++ {
			p := TrianglePixelCoverage(t, x, y)
			a := uint32(p * 0xffff)
			if a == 0 {
				continue
			}
			if a == previous {
				lines[len(lines)-1].X2 = x
			} else {
				lines = append(lines, Scanline{y, x, x, a})
			}
			previous = a
		}
	}
	return lines
}

func (t *Triangle) RasterizeFast() []Scanline {
	x1 := int(t.V1.X)
	y1 := int(t.V1.Y)
	x2 := int(t.V2.X)
	y2 := int(t.V2.Y)
	x3 := int(t.V3.X)
	y3 := int(t.V3.Y)
	return rasterizeTriangle(x1, y1, x2, y2, x3, y3, nil)
}

func rasterizeTriangle(x1, y1, x2, y2, x3, y3 int, buf []Scanline) []Scanline {
	if y1 > y3 {
		x1, x3 = x3, x1
		y1, y3 = y3, y1
	}
	if y1 > y2 {
		x1, x2 = x2, x1
		y1, y2 = y2, y1
	}
	if y2 > y3 {
		x2, x3 = x3, x2
		y2, y3 = y3, y2
	}
	if y2 == y3 {
		return rasterizeTriangleBottom(x1, y1, x2, y2, x3, y3, buf)
	} else if y1 == y2 {
		return rasterizeTriangleTop(x1, y1, x2, y2, x3, y3, buf)
	} else {
		x4 := x1 + int((float64(y2-y1)/float64(y3-y1))*float64(x3-x1))
		y4 := y2
		buf = rasterizeTriangleBottom(x1, y1, x2, y2, x4, y4, buf)
		buf = rasterizeTriangleTop(x2, y2, x4, y4, x3, y3, buf)
		return buf
	}
}

func rasterizeTriangleBottom(x1, y1, x2, y2, x3, y3 int, buf []Scanline) []Scanline {
	s1 := float64(x2-x1) / float64(y2-y1)
	s2 := float64(x3-x1) / float64(y3-y1)
	ax := float64(x1)
	bx := float64(x1)
	if s1 > s2 {
		ax, bx = bx, ax
		s1, s2 = s2, s1
	}
	for y := y1; y <= y2; y++ {
		a := int(ax)
		b := int(bx)
		ax += s1
		bx += s2
		buf = append(buf, Scanline{y, a, b, 0xffff})
	}
	return buf
}

func rasterizeTriangleTop(x1, y1, x2, y2, x3, y3 int, buf []Scanline) []Scanline {
	s1 := float64(x3-x1) / float64(y3-y1)
	s2 := float64(x3-x2) / float64(y3-y2)
	ax := float64(x3)
	bx := float64(x3)
	if s1 < s2 {
		ax, bx = bx, ax
		s1, s2 = s2, s1
	}
	for y := y3; y > y1; y-- {
		ax -= s1
		bx -= s2
		a := int(ax)
		b := int(bx)
		buf = append(buf, Scanline{y, a, b, 0xffff})
	}
	return buf
}
