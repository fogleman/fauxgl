package fauxgl

import (
	"math"
	"sort"
)

func HiddenLineRemoval(mesh *Mesh, lines *Mesh, eye Vector) *Mesh {
	var result []*Line
	type Edge struct {
		T float64
		D int
	}
	var edges []Edge
	for _, line := range lines.Lines {
		v1 := line.V1.Position
		v2 := line.V2.Position
		edges = edges[:0]
		for _, t := range mesh.Triangles {
			p1 := t.V1.Position
			p2 := t.V2.Position
			p3 := t.V3.Position
			t0, t1, ok := clipSegment(eye, v1, v2, p1, p2, p3)
			if ok {
				edges = append(edges, Edge{t0, 1})
				edges = append(edges, Edge{t1, -1})
			}
		}
		if len(edges) == 0 {
			result = append(result, line)
			continue
		}
		sort.Slice(edges, func(i, j int) bool {
			return edges[i].T < edges[j].T
		})
		var t float64
		var d int
		const eps = 1e-9
		for _, e := range edges {
			if d == 0 && e.D > 0 {
				if e.T-t > eps {
					p1 := v1.Lerp(v2, t)
					p2 := v1.Lerp(v2, e.T)
					result = append(result, NewLineForPoints(p1, p2))
				}
			}
			t = e.T
			d += e.D
		}
		if d == 0 && t < 1 {
			p1 := v1.Lerp(v2, t)
			p2 := v1.Lerp(v2, 1)
			result = append(result, NewLineForPoints(p1, p2))
		}
	}
	return NewLineMesh(result)
}

func planeSegmentIntersection(p, n, p0, p1 Vector) Vector {
	t := n.Dot(p.Sub(p0)) / n.Dot(p1.Sub(p0))
	return p0.Lerp(p1, t)
}

func planeTriangleIntersection(p, n, p0, p1, p2 Vector) (Vector, Vector, bool) {
	b0 := n.Dot(p0.Sub(p)) > 0
	b1 := n.Dot(p1.Sub(p)) > 0
	b2 := n.Dot(p2.Sub(p)) > 0
	if b0 == b1 && b1 == b2 {
		return Vector{}, Vector{}, false
	}
	var i0, i1 Vector
	if b0 == b1 {
		i0 = planeSegmentIntersection(p, n, p2, p0)
		i1 = planeSegmentIntersection(p, n, p2, p1)
	} else if b0 == b2 {
		i0 = planeSegmentIntersection(p, n, p1, p0)
		i1 = planeSegmentIntersection(p, n, p1, p2)
	} else {
		i0 = planeSegmentIntersection(p, n, p0, p1)
		i1 = planeSegmentIntersection(p, n, p0, p2)
	}
	return i0, i1, true
}

func triangleTriangleIntersection(p00, p01, p02, p10, p11, p12 Vector) (Vector, Vector, bool) {
	p0 := p00
	p1 := p10
	n0 := p01.Sub(p00).Cross(p02.Sub(p00)).Normalize()
	n1 := p11.Sub(p10).Cross(p12.Sub(p10)).Normalize()

	x00, x01, ok := planeTriangleIntersection(p0, n0, p10, p11, p12)
	if !ok {
		return Vector{}, Vector{}, false
	}

	x10, x11, ok := planeTriangleIntersection(p1, n1, p00, p01, p02)
	if !ok {
		return Vector{}, Vector{}, false
	}

	p := x00
	v := x01.Sub(x00)
	t00 := x00.Distance(x00)
	t01 := x01.Distance(x00)
	t10 := x10.Distance(x00)
	t11 := x11.Distance(x00)

	if t10 > t11 {
		t10, t11 = t11, t10
	}

	if t00 <= t11 && t10 <= t01 {
		t0 := math.Max(t00, t10)
		t1 := math.Min(t01, t11)
		p0 := p.Add(v.MulScalar(t0))
		p1 := p.Add(v.MulScalar(t1))
		return p0, p1, true
	}

	return Vector{}, Vector{}, false
}

func clipSegment(eye, p0, p1, t0, t1, t2 Vector) (float64, float64, bool) {
	x0, x1, ok := triangleTriangleIntersection(eye, p0, p1, t0, t1, t2)
	if !ok {
		return 0, 0, false
	}

	p := p0
	n := p1.Sub(p0).Normalize().Cross(p0.Sub(eye).Cross(p1.Sub(eye)).Normalize())
	x0 = planeSegmentIntersection(p, n, eye, x0)
	x1 = planeSegmentIntersection(p, n, eye, x1)

	d := p0.Distance(p1)
	u0 := x0.Distance(p0) / d
	u1 := x1.Distance(p0) / d

	if u0 > u1 {
		u0, u1 = u1, u0
	}

	if u0 < 0 && u1 < 0 {
		return 0, 0, false
	}
	if u0 > 1 && u1 > 1 {
		return 0, 0, false
	}

	if u0 < 0 {
		u0 = 0
	}
	if u1 > 1 {
		u1 = 1
	}

	return u0, u1, true
}
