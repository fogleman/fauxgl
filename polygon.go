package fauxgl

type Polygon struct {
	Exterior  Ring
	Interiors []Ring
}

func NewPolygonForTriangles(triangles []*Triangle) *Polygon {
	edgeCounts := make(map[Edge]int)
	for _, t := range triangles {
		edgeCounts[EdgeKey(t.V1.Position, t.V2.Position)]++
		edgeCounts[EdgeKey(t.V2.Position, t.V3.Position)]++
		edgeCounts[EdgeKey(t.V3.Position, t.V1.Position)]++
	}
	var edges []Edge
	for _, t := range triangles {
		e1 := Edge{t.V1.Position, t.V2.Position}
		e2 := Edge{t.V2.Position, t.V3.Position}
		e3 := Edge{t.V3.Position, t.V1.Position}
		for _, e := range []Edge{e1, e2, e3} {
			if edgeCounts[e.Key()] == 1 {
				edges = append(edges, e)
			}
		}
	}
	rings := edgesToRings(edges)
	var exterior Ring
	var length float64
	for _, r := range rings {
		l := r.Length()
		if l > length {
			exterior = r
			length = l
		}
	}
	var interiors []Ring
	for _, r := range rings {
		if r.Length() < length {
			interiors = append(interiors, r)
		}
	}
	return &Polygon{exterior, interiors}
}

func edgesToRings(edges []Edge) []Ring {
	pointEdges := make(map[Vector][]Edge)
	used := make(map[Edge]bool)
	for _, e := range edges {
		pointEdges[e.A] = append(pointEdges[e.A], e)
	}
	var rings []Ring
	for len(used) < len(edges) {
		// pick an arbitrary unused edge
		var e Edge
		for _, e = range edges {
			if !used[e] {
				break
			}
		}
		used[e] = true
		start := e.A
		var ring Ring
		for e.B != start {
			ring = append(ring, e.A)
			for _, e = range pointEdges[e.B] {
				if !used[e] {
					break
				}
			}
			used[e] = true
		}
		ring = append(ring, e.A)
		rings = append(rings, ring.Simplify())
	}
	return rings
}
