package fauxgl

func silhouette(mesh *Mesh, eye Vector, offset float64) *Mesh {
	var lines []*Line

	lerp := func(p1, p2 Vector, w1, w2 float64) Vector {
		t := -w1 / (w2 - w1)
		return p1.Lerp(p2, t)
	}

	mesh = mesh.Copy()
	mesh.SmoothNormals()
	for _, t := range mesh.Triangles {
		p1 := t.V1.Position
		p2 := t.V2.Position
		p3 := t.V3.Position
		n1 := t.V1.Normal
		n2 := t.V2.Normal
		n3 := t.V3.Normal
		g1 := n1.Dot(p1.Sub(eye))
		g2 := n2.Dot(p2.Sub(eye))
		g3 := n3.Dot(p3.Sub(eye))
		b1 := g1 > 0
		b2 := g2 > 0
		b3 := g3 > 0
		if b1 == b2 && b2 == b3 {
			continue
		}
		var v1, v2, vn1, vn2 Vector
		if b1 == b2 {
			v1 = lerp(p1, p3, g1, g3)
			vn1 = lerp(n1, n3, g1, g3)
			v2 = lerp(p2, p3, g2, g3)
			vn2 = lerp(n2, n3, g2, g3)
		} else if b1 == b3 {
			v1 = lerp(p1, p2, g1, g2)
			vn1 = lerp(n1, n2, g1, g2)
			v2 = lerp(p3, p2, g3, g2)
			vn2 = lerp(n3, n2, g3, g2)
		} else {
			v1 = lerp(p2, p1, g2, g1)
			vn1 = lerp(n2, n1, g2, g1)
			v2 = lerp(p3, p1, g3, g1)
			vn2 = lerp(n3, n1, g3, g1)
		}
		v1 = v1.Add(vn1.MulScalar(offset))
		v2 = v2.Add(vn2.MulScalar(offset))
		line := NewLineForPoints(v1, v2)
		lines = append(lines, line)
	}

	return NewLineMesh(lines)
}
