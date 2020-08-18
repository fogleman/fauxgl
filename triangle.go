package fauxgl

type Triangle struct {
	V1, V2, V3 Vertex
}

func NewTriangle(v1, v2, v3 Vertex) *Triangle {
	t := Triangle{v1, v2, v3}
	t.FixNormals()
	return &t
}

func NewTriangleForPoints(p1, p2, p3 Vector) *Triangle {
	v1 := Vertex{Position: p1}
	v2 := Vertex{Position: p2}
	v3 := Vertex{Position: p3}
	return NewTriangle(v1, v2, v3)
}

func (t *Triangle) IsDegenerate() bool {
	p1 := t.V1.Position
	p2 := t.V2.Position
	p3 := t.V3.Position
	if p1 == p2 || p1 == p3 || p2 == p3 {
		return true
	}
	if p1.IsDegenerate() || p2.IsDegenerate() || p3.IsDegenerate() {
		return true
	}
	return false
}

func (t *Triangle) Normal() Vector {
	e1 := t.V2.Position.Sub(t.V1.Position)
	e2 := t.V3.Position.Sub(t.V1.Position)
	return e1.Cross(e2).Normalize()
}

func (t *Triangle) Area() float64 {
	e1 := t.V2.Position.Sub(t.V1.Position)
	e2 := t.V3.Position.Sub(t.V1.Position)
	n := e1.Cross(e2)
	return n.Length() / 2
}

func (t *Triangle) FixNormals() {
	n := t.Normal()
	zero := Vector{}
	if t.V1.Normal == zero {
		t.V1.Normal = n
	}
	if t.V2.Normal == zero {
		t.V2.Normal = n
	}
	if t.V3.Normal == zero {
		t.V3.Normal = n
	}
}

func (t *Triangle) BoundingBox() Box {
	min := t.V1.Position.Min(t.V2.Position).Min(t.V3.Position)
	max := t.V1.Position.Max(t.V2.Position).Max(t.V3.Position)
	return Box{min, max}
}

func (t *Triangle) Transform(matrix Matrix) {
	t.V1.Position = matrix.MulPosition(t.V1.Position)
	t.V2.Position = matrix.MulPosition(t.V2.Position)
	t.V3.Position = matrix.MulPosition(t.V3.Position)
	t.V1.Normal = matrix.MulDirection(t.V1.Normal)
	t.V2.Normal = matrix.MulDirection(t.V2.Normal)
	t.V3.Normal = matrix.MulDirection(t.V3.Normal)
}

func (t *Triangle) ReverseWinding() {
	t.V1, t.V2, t.V3 = t.V3, t.V2, t.V1
	t.V1.Normal = t.V1.Normal.Negate()
	t.V2.Normal = t.V2.Normal.Negate()
	t.V3.Normal = t.V3.Normal.Negate()
}

func (t *Triangle) SetColor(c Color) {
	t.V1.Color = c
	t.V2.Color = c
	t.V3.Color = c
}

// func (t *Triangle) RandomPoint() Vector {
// 	v1 := t.V1.Position
// 	v2 := t.V2.Position.Sub(v1)
// 	v3 := t.V3.Position.Sub(v1)
// 	for {
// 		a := rand.Float64()
// 		b := rand.Float64()
// 		if a+b <= 1 {
// 			return v1.Add(v2.MulScalar(a)).Add(v3.MulScalar(b))
// 		}
// 	}
// }

func (t *Triangle) IntersectRay(r Ray) Hit {
	e1x := t.V2.Position.X - t.V1.Position.X
	e1y := t.V2.Position.Y - t.V1.Position.Y
	e1z := t.V2.Position.Z - t.V1.Position.Z
	e2x := t.V3.Position.X - t.V1.Position.X
	e2y := t.V3.Position.Y - t.V1.Position.Y
	e2z := t.V3.Position.Z - t.V1.Position.Z
	px := r.Direction.Y*e2z - r.Direction.Z*e2y
	py := r.Direction.Z*e2x - r.Direction.X*e2z
	pz := r.Direction.X*e2y - r.Direction.Y*e2x
	det := e1x*px + e1y*py + e1z*pz
	if det > -eps && det < eps {
		return NoHit
	}
	inv := 1 / det
	tx := r.Origin.X - t.V1.Position.X
	ty := r.Origin.Y - t.V1.Position.Y
	tz := r.Origin.Z - t.V1.Position.Z
	u := (tx*px + ty*py + tz*pz) * inv
	if u < 0 || u > 1 {
		return NoHit
	}
	qx := ty*e1z - tz*e1y
	qy := tz*e1x - tx*e1z
	qz := tx*e1y - ty*e1x
	v := (r.Direction.X*qx + r.Direction.Y*qy + r.Direction.Z*qz) * inv
	if v < 0 || u+v > 1 {
		return NoHit
	}
	d := (e2x*qx + e2y*qy + e2z*qz) * inv
	if d < eps {
		return NoHit
	}
	return Hit{t, d}
}
