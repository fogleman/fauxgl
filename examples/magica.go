package main

import (
	"fmt"
	"time"

	. "github.com/fogleman/fauxgl"
	"github.com/nfnt/resize"
)

const (
	scale  = 4    // optional supersampling
	width  = 2048 // output width in pixels
	height = 2048 // output height in pixels
	fovy   = 10   // vertical field of view in degrees
	near   = 1    // near clipping plane
	far    = 100  // far clipping plane
)

var (
	eye    = V(-10, -10, 10)                // camera position
	center = V(0, 0, -0.3)                  // view center position
	up     = V(0, 0, 1)                     // up vector
	light  = V(-0.75, -0.25, 1).Normalize() // light direction
)

func main() {
	// load a mesh
	mesh, err := LoadVOX("examples/vox/monument/monu7.vox")
	if err != nil {
		panic(err)
	}

	fmt.Println(len(mesh.Triangles))
	mesh.SaveSTL("out.stl")

	// fit mesh in a bi-unit cube centered at the origin
	mesh.BiUnitCube()

	// create a rendering context
	context := NewContext(width*scale, height*scale)

	// create transformation matrix and light direction
	aspect := float64(width) / float64(height)
	matrix := LookAt(eye, center, up).Perspective(fovy, aspect, near, far)

	// render
	context.ClearColorBufferWith(Black)
	ambient := Color{0.4, 0.4, 0.4, 1}
	diffuse := Color{0.9, 0.9, 0.9, 1}
	context.Shader = NewDiffuseShader(matrix, light, Discard, ambient, diffuse)
	start := time.Now()
	context.DrawMesh(mesh)
	fmt.Println(time.Since(start))

	context.Shader = NewSolidColorShader(matrix, Black)
	context.Wireframe = true
	context.DepthBias = -1e-5
	start = time.Now()
	context.DrawMesh(mesh)
	fmt.Println(time.Since(start))

	// downsample image for antialiasing
	image := context.Image()
	image = resize.Resize(width, height, image, resize.Bilinear)

	// save image
	SavePNG("out.png", image)
}
