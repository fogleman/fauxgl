package main

import (
	"fmt"
	"time"

	. "github.com/fogleman/fauxgl"
	"github.com/nfnt/resize"
)

// download dragon.obj from here:
// http://graphics.cs.williams.edu/data/meshes/dragon.zip

const (
	scale  = 4    // optional supersampling
	width  = 2048 // output width in pixels
	height = 2048 // output height in pixels
	fovy   = 30   // vertical field of view in degrees
	near   = 1    // near clipping plane
	far    = 10   // far clipping plane
)

var (
	eye    = V(-3, 1.25, -2)               // camera position
	center = V(0, -0.1, -0.1)              // view center position
	up     = V(0, 1, 0)                    // up vector
	light  = V(-0.75, 1, 0.25).Normalize() // light direction
	color  = HexColor(0x468966)            // object color
)

func main() {
	// load a mesh
	mesh, err := LoadOBJ("examples/dragon.obj")
	if err != nil {
		panic(err)
	}

	// fit mesh in a bi-unit cube centered at the origin
	mesh.BiUnitCube()

	// smooth the normals
	mesh.SmoothNormalsThreshold(Radians(30))

	// create a rendering context
	context := NewContext(width*scale, height*scale)
	context.ClearColorBuffer(HexColor(0xFFF8E3))

	// create transformation matrix and light direction
	aspect := float64(width) / float64(height)
	matrix := LookAt(eye, center, up).Perspective(fovy, aspect, near, far)

	// render
	shader := NewDefaultShader(matrix, light, eye, color)
	start := time.Now()
	context.DrawMesh(mesh, shader)
	fmt.Println(time.Since(start))

	// downsample image for antialiasing
	image := context.Image()
	image = resize.Resize(width, height, image, resize.Bilinear)

	// save image
	SavePNG("out.png", image)
}
