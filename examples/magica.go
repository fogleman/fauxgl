package main

import (
	"fmt"
	"os"
	"time"

	. "github.com/fogleman/fauxgl"
	"github.com/nfnt/resize"
)

const (
	scale  = 4    // optional supersampling
	width  = 1600 // output width in pixels
	height = 1600 // output height in pixels
	fovy   = 10   // vertical field of view in degrees
	near   = 1    // near clipping plane
	far    = 100  // far clipping plane
)

var (
	eye    = V(-10, -10, 10)                // camera position
	center = V(0, 0, -0.2)                  // view center position
	up     = V(0, 0, 1)                     // up vector
	light  = V(-0.25, -0.75, 1).Normalize() // light direction
)

func main() {
	// load a mesh
	mesh, err := LoadVOX(os.Args[1])
	if err != nil {
		panic(err)
	}

	fmt.Println(len(mesh.Triangles))
	// mesh.SaveSTL("out.stl")

	// fit mesh in a bi-unit cube centered at the origin
	mesh.BiUnitCube()

	// create a rendering context
	context := NewContext(width*scale, height*scale)

	// create transformation matrix and light direction
	// aspect := float64(width) / float64(height)
	// matrix := LookAt(eye, center, up).Perspective(fovy, aspect, near, far)
	matrix := LookAt(eye, center, up).Orthographic(-1.45, 1.45, -1.45, 1.45, -20, 20)

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
