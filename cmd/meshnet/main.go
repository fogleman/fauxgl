package main

import (
	"container/heap"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/fogleman/colormap"
	. "github.com/fogleman/fauxgl"
	"github.com/fogleman/gg"
)

const (
	W = 4096
	H = 4096
)

func circleCircleIntersection(p0 Vector, r0 float64, p1 Vector, r1 float64) Vector {
	x0, y0 := p0.X, p0.Y
	x1, y1 := p1.X, p1.Y
	d := math.Hypot(x1-x0, y1-y0)
	a := (r0*r0 - r1*r1 + d*d) / (2 * d)
	h := math.Sqrt(r0*r0 - a*a)
	x2 := x0 + a*(x1-x0)/d
	y2 := y0 + a*(y1-y0)/d
	x := x2 + h*(y1-y0)/d
	y := y2 - h*(x1-x0)/d
	return Vector{x, y, 0}
}

type Edge struct {
	A, B Vector
}

func (e Edge) Opposite() Edge {
	return Edge{e.B, e.A}
}

func (e Edge) Length() float64 {
	return e.A.Distance(e.B)
}

func triangleEdges(t *Triangle) [3]Edge {
	p0 := t.V1.Position
	p1 := t.V2.Position
	p2 := t.V3.Position
	e0 := Edge{p0, p1}
	e1 := Edge{p1, p2}
	e2 := Edge{p2, p0}
	return [3]Edge{e0, e1, e2}
}

func nextEdges(t *Triangle, e Edge) (Edge, Edge) {
	edges := triangleEdges(t)
	if e == edges[0] {
		return edges[1], edges[2]
	}
	if e == edges[1] {
		return edges[2], edges[0]
	}
	if e == edges[2] {
		return edges[0], edges[1]
	}
	panic("edge not found")
}

type PriorityItem struct {
	Edge
	Score float64
}

type PriorityQueue []PriorityItem

func (pq PriorityQueue) Len() int {
	return len(pq)
}

func (pq PriorityQueue) Less(i, j int) bool {
	return -pq[i].Score < -pq[j].Score
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	// pq[i].QueueIndex = i
	// pq[j].QueueIndex = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(PriorityItem)
	// item.QueueIndex = len(*pq)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	// item.QueueIndex = -1
	*pq = old[:n-1]
	return item
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	// load input mesh
	mesh, err := LoadMesh(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// mesh.SplitTriangles(1)

	// maps a 3D directed edge to its triangle
	edgeTriangle := make(map[Edge]*Triangle)
	for _, t := range mesh.Triangles {
		// if t.Normal().Dot(Vector{0, 0, -1}) > 0.99 {
		// 	continue
		// }
		edges := triangleEdges(t)
		edgeTriangle[edges[0]] = t
		edgeTriangle[edges[1]] = t
		edgeTriangle[edges[2]] = t
	}

	// maps an edge in 3D to its 2D placement
	flatEdge := make(map[Edge]Edge)
	// flatTriangles := make(map[*Triangle]*Triangle)
	var triangles []*Triangle

	// true for processed triangles
	done := make(map[*Triangle]bool)

	scoreForEdge := func(e Edge) float64 {
		a := edgeTriangle[e]
		b := edgeTriangle[e.Opposite()]
		if a == nil || b == nil {
			return 0
		}
		return a.Normal().Dot(b.Normal())
	}

	// queue of remaining edges to process
	var q PriorityQueue

	// first triangle
	{
		i := rand.Intn(len(mesh.Triangles))
		t := mesh.Triangles[i]
		edges := triangleEdges(t)

		// compute flat points
		p0 := Vector{}
		p1 := Vector{edges[0].Length(), 0, 0}
		p2 := circleCircleIntersection(p0, edges[2].Length(), p1, edges[1].Length())

		// update data structures
		// flatTriangles[t] = NewTriangleForPoints(p0, p1, p2)
		triangles = append(triangles, NewTriangleForPoints(p0, p1, p2))
		flatEdge[edges[0]] = Edge{p0, p1}
		flatEdge[edges[1]] = Edge{p1, p2}
		flatEdge[edges[2]] = Edge{p2, p0}
		for _, e := range edges {
			e = e.Opposite()
			score := scoreForEdge(e)
			item := PriorityItem{e, score}
			heap.Push(&q, item)
		}
		// q = append(q, edges[0].Opposite())
		// q = append(q, edges[1].Opposite())
		// q = append(q, edges[2].Opposite())
		done[t] = true
	}

	// while queue is non empty
	for len(q) > 0 {
		// pop an edge from the queue
		e := heap.Pop(&q).(PriorityItem).Edge
		// e := q[0]
		// q = q[1:]

		// lookup edge triangle
		t := edgeTriangle[e]
		if t == nil || done[t] {
			continue
		}

		// compute third point
		e2, e3 := nextEdges(t, e)
		f := flatEdge[e.Opposite()].Opposite()
		p := circleCircleIntersection(f.A, e3.Length(), f.B, e2.Length())

		if p.IsDegenerate() {
			// fmt.Println("p.IsDegenerate()")
			continue
		}

		// update data structures
		// flatTriangles[t] = NewTriangleForPoints(f.A, f.B, p)
		triangles = append(triangles, NewTriangleForPoints(f.A, f.B, p))
		flatEdge[e2] = Edge{f.B, p}
		flatEdge[e3] = Edge{p, f.A}
		// q = append(q, e2.Opposite())
		// q = append(q, e3.Opposite())
		heap.Push(&q, PriorityItem{e2.Opposite(), scoreForEdge(e2.Opposite())})
		heap.Push(&q, PriorityItem{e3.Opposite(), scoreForEdge(e3.Opposite())})
		done[t] = true

		// fmt.Println(len(done))
	}

	// compute flat bounds
	var x0, y0, x1, y1, totalArea float64
	for _, t := range triangles {
		totalArea += t.Area()
		for _, p := range []Vector{t.V1.Position, t.V2.Position, t.V3.Position} {
			x0 = math.Min(x0, p.X)
			y0 = math.Min(y0, p.Y)
			x1 = math.Max(x1, p.X)
			y1 = math.Max(y1, p.Y)
		}
	}
	fmt.Println(x0, y0, x1, y1)
	dx := x1 - x0
	dy := y1 - y0
	s := math.Min(W/dx, H/dy) * 0.95
	fmt.Println(s)

	dc := gg.NewContext(W, H)
	dc.Translate(W/2, H/2)
	dc.Scale(s, s)
	dc.Translate(-(x0 + dx/2), -(y0 + dy/2))
	dc.SetRGB(1, 1, 1)
	dc.SetRGB(0, 0, 0)
	dc.Clear()
	dc.SetRGBA(0, 0, 0, 0.5)
	dc.SetLineWidth(0.5)
	var area float64
	for _, t := range triangles {
		area += t.Area()
		dc.SetColor(colormap.Blues.At(area / totalArea))
		p0 := t.V1.Position
		p1 := t.V2.Position
		p2 := t.V3.Position
		dc.LineTo(p0.X, p0.Y)
		dc.LineTo(p1.X, p1.Y)
		dc.LineTo(p2.X, p2.Y)
		dc.ClosePath()
		dc.FillPreserve()
		dc.Stroke()
	}
	dc.SavePNG("out.png")
}
