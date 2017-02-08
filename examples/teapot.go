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
	light := V(0, 1, 1).Normalize()

	// render
	start := time.Now()
	shader := NewPhongShader(matrix, light, eye)
	shader.ObjectColor = HexColor("#B9121B")
	shader.SpecularColor = Gray(0.25)
	shader.SpecularPower = 64
	context.Shader = shader
	context.DrawMesh(mesh)
	fmt.Println(time.Since(start))

	// save image
	image := context.Image()
	image = resize.Resize(width, height, image, resize.Bilinear)
	SavePNG("out.png", image)
}
