package fauxgl

var clipPlanes = []clipPlane{
	{VectorW{1, 0, 0, 1}, VectorW{-1, 0, 0, 1}},
	{VectorW{-1, 0, 0, 1}, VectorW{1, 0, 0, 1}},
	{VectorW{0, 1, 0, 1}, VectorW{0, -1, 0, 1}},
	{VectorW{0, -1, 0, 1}, VectorW{0, 1, 0, 1}},
	{VectorW{0, 0, 1, 1}, VectorW{0, 0, -1, 1}},
	{VectorW{0, 0, -1, 1}, VectorW{0, 0, 1, 1}},
}

type clipPlane struct {
	P, N VectorW
}

func (p clipPlane) pointInFront(v VectorW) bool {
	return v.Sub(p.P).Dot(p.N) > 0
}

func (p clipPlane) intersectSegment(v0, v1 VectorW) VectorW {
	u := v1.Sub(v0)
	w := v0.Sub(p.P)
	d := p.N.Dot(u)
	n := -p.N.Dot(w)
	return v0.Add(u.MulScalar(n / d))
}

func sutherlandHodgman(points []VectorW, planes []clipPlane) []VectorW {
	output := points
	for _, plane := range planes {
		input := output
		output = nil
		if len(input) == 0 {
			return nil
		}
		s := input[len(input)-1]
		for _, e := range input {
			if plane.pointInFront(e) {
				if !plane.pointInFront(s) {
					x := plane.intersectSegment(s, e)
					output = append(output, x)
				}
				output = append(output, e)
			} else if plane.pointInFront(s) {
				x := plane.intersectSegment(s, e)
				output = append(output, x)
			}
			s = e
		}
	}
	return output
}

func ClipTriangle(t *Triangle) []*Triangle {
	w1 := t.V1.Output
	w2 := t.V2.Output
	w3 := t.V3.Output
	p1 := w1.Vector()
	p2 := w2.Vector()
	p3 := w3.Vector()
	points := []VectorW{w1, w2, w3}
	newPoints := sutherlandHodgman(points, clipPlanes)
	var result []*Triangle
	for i := 2; i < len(newPoints); i++ {
		b1 := Barycentric(p1, p2, p3, newPoints[0].Vector())
		b2 := Barycentric(p1, p2, p3, newPoints[i-1].Vector())
		b3 := Barycentric(p1, p2, p3, newPoints[i].Vector())
		v1 := InterpolateVertexes(t.V1, t.V2, t.V3, b1)
		v2 := InterpolateVertexes(t.V1, t.V2, t.V3, b2)
		v3 := InterpolateVertexes(t.V1, t.V2, t.V3, b3)
		result = append(result, NewTriangle(v1, v2, v3))
	}
	return result
}

func ClipLine(l *Line) *Line {
	// TODO: interpolate vertex attributes when clipped
	w1 := l.V1.Output
	w2 := l.V2.Output
	for _, plane := range clipPlanes {
		f1 := plane.pointInFront(w1)
		f2 := plane.pointInFront(w2)
		if f1 && f2 {
			continue
		} else if f1 {
			w2 = plane.intersectSegment(w1, w2)
		} else if f2 {
			w1 = plane.intersectSegment(w2, w1)
		} else {
			return nil
		}
	}
	v1 := l.V1
	v2 := l.V2
	v1.Output = w1
	v2.Output = w2
	return NewLine(v1, v2)
}
