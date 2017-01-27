package main

import (
	"fmt"
	"time"

	. "github.com/fogleman/soft/soft"
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

	// create a rendering context
	context := NewContext(width*scale, height*scale)

	// create transformation matrix and light direction
	aspect := float64(width) / float64(height)
	matrix := LookAt(eye, center, up).Perspective(fovy, aspect, near, far)
	light := V(0.75, 0.25, 1).Normalize()
	color := V(1, 1, 1)

	// render
	context.ClearColorBuffer(V(0, 0, 0))
	context.ClearDepthBuffer()
	shader := NewDefaultShader(matrix, light, eye, color)
	start := time.Now()
	context.DrawMesh(mesh, shader)
	fmt.Println(time.Since(start))

	// save image
	image := context.Image()
	image = resize.Resize(width, height, image, resize.Bilinear)
	SavePNG("out.png", image)
}
