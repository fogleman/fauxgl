package soft

type clipPlane struct {
	P, N Vector
}

func (p clipPlane) pointInFront(v Vector) bool {
	return v.Sub(p.P).Dot(p.N) > 0
}

func (p clipPlane) intersectSegment(v0, v1 Vector) Vector {
	u := v1.Sub(v0)
	w := v0.Sub(p.P)
	d := p.N.Dot(u)
	n := -p.N.Dot(w)
	return v0.Add(u.MulScalar(n / d))
}

func sutherlandHodgman(points []Vector, planes []clipPlane) []Vector {
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
	return nil
}
