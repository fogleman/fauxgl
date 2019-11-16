package main

import (
	"fmt"
	"time"

	. "github.com/fogleman/fauxgl"
	"github.com/nfnt/resize"
)

const (
	scale  = 4
	width  = 1024
	height = 1024
	fovy   = 30
	near   = 1
	far    = 10
)

var (
	eye    = V(3, 1, 0.5)
	center = V(0, -0.1, 0)
	up     = V(0, 0, 1)
)

func main() {
	// load a mesh
	mesh, err := LoadSTL("examples/bowser.stl")
	if err != nil {
		panic(err)
	}

	// fit mesh in a bi-unit cube centered at the origin
	mesh.BiUnitCube()

	// smooth the normals
	mesh.SmoothNormalsThreshold(Radians(30))

	var lines []*Line
	for _, t := range mesh.Triangles {
		p1 := t.V1.Position
		p2 := t.V2.Position
		p3 := t.V3.Position
		if p1.Less(p2) {
			lines = append(lines, NewLineForPoints(p1, p2))
		}
		if p1.Less(p3) {
			lines = append(lines, NewLineForPoints(p1, p3))
		}
		if p2.Less(p3) {
			lines = append(lines, NewLineForPoints(p2, p3))
		}
	}
	lineMesh := NewLineMesh(lines)
	lineMesh = HiddenLineRemoval(mesh, lineMesh, eye)
	lines = lineMesh.Lines

	// create a rendering context
	context := NewContext(width*scale, height*scale)
	context.ClearColorBufferWith(White)

	// create transformation matrix and light direction
	aspect := float64(width) / float64(height)
	matrix := LookAt(eye, center, up).Perspective(fovy, aspect, near, far)

	// render
	context.Shader = NewSolidColorShader(matrix, Black)
	start := time.Now()
	info := context.DrawLines(lines)
	fmt.Println(info)
	fmt.Println(time.Since(start))

	// save image
	image := context.Image()
	image = resize.Resize(width, height, image, resize.Bilinear)
	SavePNG("out.png", image)
}
