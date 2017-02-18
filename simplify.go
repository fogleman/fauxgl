package fauxgl

import (
	"fmt"
	"math/rand"
)

func MeshEdges(mesh *Mesh) {
	// map edges to faces
	edgeTriangles := make(map[Edge][]*Triangle)
	for _, t := range mesh.Triangles {
		e1 := EdgeKey(t.V1.Position, t.V2.Position)
		e2 := EdgeKey(t.V2.Position, t.V3.Position)
		e3 := EdgeKey(t.V3.Position, t.V1.Position)
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
			e1 := EdgeKey(t.V1.Position, t.V2.Position)
			e2 := EdgeKey(t.V2.Position, t.V3.Position)
			e3 := EdgeKey(t.V3.Position, t.V1.Position)
			for _, e := range []Edge{e1, e2, e3} {
				for _, u := range edgeTriangles[e] {
					if t == u || !n.NearEqual(u.Normal(), 1e-9) {
						continue
					}
					if seen[u] {
						continue
					}
					group = append(group, u)
					seen[u] = true
				}
			}
		}
		polygon := NewPolygonForTriangles(group)
		r := polygon.Exterior
		// for _, r := range polygon.Interiors {
		for i := range r {
			p1 := r.At(i)
			p2 := r.At(i + 1)
			mesh.Lines = append(mesh.Lines, NewLineForPoints(p1, p2))
		}
		// }
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
