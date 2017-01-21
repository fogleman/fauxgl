package soft

import "math"

type Segment struct {
	P0, P1 Vector
}

func (a Segment) Point(t float64) Vector {
	return a.P0.Add(a.P1.Sub(a.P0).MulScalar(t))
}

func (a Segment) PointInFront(p Vector) bool {
	d := a.P1.Sub(a.P0)
	n := Vector{-d.Y, d.X, 0}
	return p.Sub(a.P1).Dot(n) > 0
}

func (a Segment) Intersect(b Segment) Vector {
	p := a.P0
	q := b.P0
	r := a.P1.Sub(a.P0)
	s := b.P1.Sub(b.P0)
	if NearZero(r.Cross2D(s)) {
		return Vector{}
	}
	qp := q.Sub(p)
	rs := r.Cross2D(s)
	t := qp.Cross2D(s) / rs
	return a.Point(t)
}

func SutherlandHodgman(points []Vector, segments []Segment) []Vector {
	output := points
	for _, segment := range segments {
		input := output
		output = nil
		if len(input) == 0 {
			return nil
		}
		s := input[len(input)-1]
		for _, e := range input {
			if segment.PointInFront(e) {
				if !segment.PointInFront(s) {
					x := segment.Intersect(Segment{s, e})
					output = append(output, x)
				}
				output = append(output, e)
			} else if segment.PointInFront(s) {
				x := segment.Intersect(Segment{s, e})
				output = append(output, x)
			}
			s = e
		}
	}
	return output
}

func PolygonArea(p []Vector) float64 {
	var area float64
	j := len(p) - 1
	for i := 0; i < len(p); i++ {
		area += (p[j].X + p[i].X) * (p[j].Y - p[i].Y)
		j = i
	}
	return math.Abs(area / 2)
}

func TrianglePixelCoverage(t *Triangle, x, y int) float64 {
	v1 := t.V1
	v2 := t.V2
	v3 := t.V3
	v1.Z = 0
	v2.Z = 0
	v3.Z = 0
	if t.IsClockwise() {
		v1, v2, v3 = v3, v2, v1
	}
	segments := []Segment{
		{v1, v2},
		{v2, v3},
		{v3, v1},
	}
	fx := float64(x) - 0.5
	fy := float64(y) - 0.5
	points := []Vector{
		{fx, fy, 0},
		{fx + 1, fy, 0},
		{fx + 1, fy + 1, 0},
		{fx, fy + 1, 0},
		{fx, fy, 0},
	}
	points = SutherlandHodgman(points, segments)
	if len(points) < 3 {
		return 0
	}
	return PolygonArea(points)
}
