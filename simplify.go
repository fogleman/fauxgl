package fauxgl

import (
	"fmt"
	"math"
	"math/rand"
)

type edge struct {
	A, B Vector
}

func sortedEdge(a, b Vector) edge {
	if b.Less(a) {
		a, b = b, a
	}
	return edge{a, b}
}

func edgesToRings(edges []edge) [][]Vector {
	pointToEdge := make(map[Vector]edge)
	pending := make(map[edge]bool)
	for _, e := range edges {
		pointToEdge[e.A] = e
		pending[e] = true
	}
	var rings [][]Vector
	for len(pending) > 0 {
		var e edge
		for e, _ = range pending {
			break
		}
		var ring []Vector
		for pending[e] {
			ring = append(ring, e.A)
			delete(pending, e)
			e = pointToEdge[e.B]
		}
		ring = simplifyRing(ring)
		rings = append(rings, ring)
	}
	return rings
}

func simplifyRing(ring []Vector) []Vector {
	n := len(ring)
	straight := make([]bool, len(ring))
	start := -1
	for i := 0; i < n; i++ {
		p1 := ring[(i+n-1)%n]
		p2 := ring[i]
		p3 := ring[(i+1)%n]
		v1 := p2.Sub(p1).Normalize()
		v2 := p3.Sub(p2).Normalize()
		c := v1.Dot(v2)
		straight[i] = math.Abs(c-1) < 1e-9
		if start < 0 && !straight[i] {
			start = i
		}
	}
	var result []Vector
	for i := start; i < start+n; i++ {
		j := i % n
		if !straight[j] {
			result = append(result, ring[j])
		}
	}
	return result
}

func MeshEdges(mesh *Mesh) {
	// map edges to faces
	edgeTriangles := make(map[edge][]*Triangle)
	for _, t := range mesh.Triangles {
		e1 := sortedEdge(t.V1.Position, t.V2.Position)
		e2 := sortedEdge(t.V2.Position, t.V3.Position)
		e3 := sortedEdge(t.V3.Position, t.V1.Position)
		edgeTriangles[e1] = append(edgeTriangles[e1], t)
		edgeTriangles[e2] = append(edgeTriangles[e2], t)
		edgeTriangles[e3] = append(edgeTriangles[e3], t)
	}

	// group neighbors with same normal
	seen := make(map[*Triangle]bool)
	var groups [][]*Triangle
	for _, t := range mesh.Triangles {
		if seen[t] {
			continue
		}
		n := t.Normal()
		var group []*Triangle
		group = append(group, t)
		seen[t] = true
		interiorEdges := make(map[edge]bool)
		for i := 0; i < len(group); i++ {
			t := group[i]
			e1 := sortedEdge(t.V1.Position, t.V2.Position)
			e2 := sortedEdge(t.V2.Position, t.V3.Position)
			e3 := sortedEdge(t.V3.Position, t.V1.Position)
			for _, e := range []edge{e1, e2, e3} {
				for _, u := range edgeTriangles[e] {
					if t == u || !n.NearEqual(u.Normal(), 1e-9) {
						continue
					}
					interiorEdges[e] = true
					if seen[u] {
						continue
					}
					group = append(group, u)
					seen[u] = true
				}
			}
		}
		var exteriorEdges []edge
		for _, t := range group {
			v1 := t.V1.Position
			v2 := t.V2.Position
			v3 := t.V3.Position
			if !interiorEdges[sortedEdge(v1, v2)] {
				exteriorEdges = append(exteriorEdges, edge{v1, v2})
			}
			if !interiorEdges[sortedEdge(v2, v3)] {
				exteriorEdges = append(exteriorEdges, edge{v2, v3})
			}
			if !interiorEdges[sortedEdge(v3, v1)] {
				exteriorEdges = append(exteriorEdges, edge{v3, v1})
			}
		}
		rings := edgesToRings(exteriorEdges)
		for _, ring := range rings {
			n := len(ring)
			for i := range ring {
				p1 := ring[i]
				p2 := ring[(i+1)%n]
				mesh.Lines = append(mesh.Lines, NewLineForPoints(p1, p2))
			}
		}
		groups = append(groups, group)
		c := Color{rand.Float64(), rand.Float64(), rand.Float64(), 1}
		for _, t := range group {
			t.V1.Color = c
			t.V2.Color = c
			t.V3.Color = c
		}
	}

	fmt.Println(len(mesh.Triangles), len(groups))
}
