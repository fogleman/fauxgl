package main

import (
	"fmt"
	"time"

	. "github.com/fogleman/fauxgl"
	"github.com/nfnt/resize"
)

const (
	scale  = 4
	width  = 1920
	height = 1080
	fovy   = 30
	near   = 1
	far    = 10
)

var (
	eye    = V(0, 2.4, 0)
	center = V(0, 0, 0)
	up     = V(0, 0, 1)
)

func main() {
	// load a mesh
	mesh, err := LoadPLY("examples/ply/teapot.ply")
	if err != nil {
		panic(err)
	}

	// fit mesh in a bi-unit cube centered at the origin
	mesh.BiUnitCube()

	// smooth the normals
	mesh.SmoothNormalsThreshold(Radians(60))

	// create a rendering context
	context := NewContext(width*scale, height*scale)
	context.ClearColorBufferWith(White)

	// create transformation matrix and light direction
	aspect := float64(width) / float64(height)
	matrix := LookAt(eye, center, up).Perspective(fovy, aspect, near, far)
	// h := 0.6
	// w := h * float64(width) / float64(height)
	// matrix := LookAt(eye, center, up).Orthographic(-w, w, -h, h, -10, 10)
	light := V(0, 1, 1).Normalize()
	color := HexColor("#B9121B")

	// render
	start := time.Now()
	context.Shader = NewDefaultShader(matrix, light, eye, color)
	context.DrawMesh(mesh)
	fmt.Println(time.Since(start))

	// save image
	image := context.Image()
	image = resize.Resize(width, height, image, resize.Bilinear)
	SavePNG("out.png", image)
}
