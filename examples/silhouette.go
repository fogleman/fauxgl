package main

import (
	"fmt"
	"math"
	"os"

	. "github.com/fogleman/fauxgl"
	"github.com/nfnt/resize"
)

const (
	scale  = 2
	width  = 1600
	height = 1600
	fovy   = 100
	near   = 0.01
	far    = 10
)

var (
	eye    = V(0, -1.1, 0)
	center = V(0, 0, 0)
	up     = V(0, 0, 1)
)

func timed(name string) func() {
	if len(name) > 0 {
		// fmt.Printf("%s... ", name)
	}
	// start := time.Now()
	return func() {
		// fmt.Println(time.Since(start))
	}
}

func fineTriangle(buf []*Triangle, t *Triangle, threshold float64) []*Triangle {
	v1 := t.V1
	v2 := t.V2
	v3 := t.V3
	p1 := v1.Position
	p2 := v2.Position
	p3 := v3.Position
	d12 := p1.Distance(p2)
	d23 := p2.Distance(p3)
	d31 := p3.Distance(p1)
	max := math.Max(d12, math.Max(d23, d31))
	if max <= threshold {
		buf = append(buf, t)
		return buf
	}
	if d12 == max {
		v := InterpolateVertexes(v1, v2, v3, VectorW{0.5, 0.5, 0, 1})
		t1 := NewTriangle(v3, v1, v)
		t2 := NewTriangle(v2, v3, v)
		buf = fineTriangle(buf, t1, threshold)
		buf = fineTriangle(buf, t2, threshold)
	} else if d23 == max {
		v := InterpolateVertexes(v1, v2, v3, VectorW{0, 0.5, 0.5, 1})
		t1 := NewTriangle(v1, v2, v)
		t2 := NewTriangle(v3, v1, v)
		buf = fineTriangle(buf, t1, threshold)
		buf = fineTriangle(buf, t2, threshold)
	} else {
		v := InterpolateVertexes(v1, v2, v3, VectorW{0.5, 0, 0.5, 1})
		t1 := NewTriangle(v2, v3, v)
		t2 := NewTriangle(v1, v2, v)
		buf = fineTriangle(buf, t1, threshold)
		buf = fineTriangle(buf, t2, threshold)
	}
	return buf
}

func fineMesh(mesh *Mesh, threshold float64) *Mesh {
	var triangles []*Triangle
	for _, t := range mesh.Triangles {
		triangles = fineTriangle(triangles, t, threshold)
	}
	return NewTriangleMesh(triangles)
}

type Edge struct {
	A, B Vector
}

func MakeEdge(a, b Vector) Edge {
	if a.Less(b) {
		return Edge{a, b}
	}
	return Edge{b, a}
}

func sharpEdges(mesh *Mesh) *Mesh {
	var lines []*Line
	other := make(map[Edge]*Triangle)
	for _, t := range mesh.Triangles {
		p1 := t.V1.Position //.RoundPlaces(6)
		p2 := t.V2.Position //.RoundPlaces(6)
		p3 := t.V3.Position //.RoundPlaces(6)
		e1 := MakeEdge(p1, p2)
		e2 := MakeEdge(p2, p3)
		e3 := MakeEdge(p3, p1)
		for _, e := range []Edge{e1, e2, e3} {
			if u, ok := other[e]; ok {
				a := math.Acos(t.Normal().Dot(u.Normal()))
				if a > Radians(60) {
					lines = append(lines, NewLineForPoints(e.A, e.B))
				}
			}
		}
		other[e1] = t
		other[e2] = t
		other[e3] = t
	}
	return NewLineMesh(lines)
}

func main() {
	var done func()

	// load a mesh
	done = timed("loading mesh")
	mesh, err := LoadMesh(os.Args[1])
	if err != nil {
		panic(err)
	}
	done()

	// fit mesh in a bi-unit cube centered at the origin
	done = timed("transforming mesh")
	mesh.BiUnitCube()
	done()

	// fmt.Println(len(mesh.Triangles))
	mesh = fineMesh(mesh, 0.01/3)
	// fmt.Println(len(mesh.Triangles))

	// create a rendering context
	context := NewContext(width*scale, height*scale)

	// create transformation matrix and light direction
	aspect := float64(width) / float64(height)
	matrix := LookAt(eye, center, up).Perspective(fovy, aspect, near, far)
	// const s = 1.1
	// matrix := LookAt(eye, center, up).Orthographic(-aspect*s, aspect*s, -s, s, near, far)

	// render
	context.Shader = NewSolidColorShader(matrix, Black)
	done = timed("rendering mesh")
	context.DrawMesh(mesh)
	done()

	context.ClearColorBufferWith(White)
	context.DepthBias = -1e-5

	done = timed("rendering mesh")

	// context.Shader = NewSolidColorShader(matrix, Color{1, 0, 0, 1})
	// context.DrawMesh(sharpEdges(mesh))

	context.Shader = NewSolidColorShader(matrix, Black)
	// context.DrawMesh(mesh.Silhouette(eye, 1e-3))
	for _, line := range mesh.Silhouette(eye, 1e-3).Lines {
		info := context.DrawLine(line)
		ratio := float64(info.UpdatedPixels) / float64(info.TotalPixels)
		if ratio < 0.666 {
			continue
		}
		v1 := matrix.MulPositionW(line.V1.Position)
		v1 = v1.DivScalar(v1.W)
		v2 := matrix.MulPositionW(line.V2.Position)
		v2 = v2.DivScalar(v2.W)
		if math.IsNaN(v1.X) || math.IsNaN(v2.X) {
			continue
		}
		fmt.Printf("%g,%g %g,%g\n", v1.X*aspect, v1.Y, v2.X*aspect, v2.Y)
	}

	done()

	// downsample image for antialiasing
	done = timed("downsampling image")
	image := context.Image()
	image = resize.Resize(width, height, image, resize.Bilinear)
	done()

	// save image
	done = timed("writing output")
	SavePNG("out-fine.png", image)
	done()
}
