package fauxgl

import (
	"fmt"
	"math/rand"
)

type edge struct {
	A, B Vector
}

func makeEdge(a, b Vector) edge {
	if b.Less(a) {
		a, b = b, a
	}
	return edge{a, b}
}

func MeshEdges(mesh *Mesh) {
	// map edges to faces
	edgeTriangles := make(map[edge][]*Triangle)
	for _, t := range mesh.Triangles {
		v1 := t.V1.Position
		v2 := t.V2.Position
		v3 := t.V3.Position
		fmt.Println(v1)
		e1 := makeEdge(v1, v2)
		e2 := makeEdge(v2, v3)
		e3 := makeEdge(v3, v1)
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
		for i := 0; i < len(group); i++ {
			t := group[i]
			e1 := makeEdge(t.V1.Position, t.V2.Position)
			e2 := makeEdge(t.V2.Position, t.V3.Position)
			e3 := makeEdge(t.V3.Position, t.V1.Position)
			for _, e := range []edge{e1, e2, e3} {
				for _, u := range edgeTriangles[e] {
					if t == u || seen[u] {
						continue
					}
					if !n.NearEqual(u.Normal(), 1e-9) {
						mesh.Lines = append(mesh.Lines, NewLineForPoints(e.A, e.B))
						continue
					}
					group = append(group, u)
					seen[u] = true
				}
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
