package main

import (
	"fmt"
	"time"

	. "github.com/fogleman/fauxgl"
	"github.com/nfnt/resize"
)

const (
	scale  = 4    // optional supersampling
	width  = 1024 // output width in pixels
	height = 1024 // output height in pixels
	fovy   = 40   // vertical field of view in degrees
	near   = 1    // near clipping plane
	far    = 10   // far clipping plane
)

var (
	eye    = V(-3, 1, 2)                // camera position
	center = V(0, 0, 0)                 // view center position
	up     = V(0, 0, 1)                 // up vector
	light  = V(-1, 1, 0.25).Normalize() // light direction
)

func main() {
	// load the mesh
	mesh, err := LoadOBJ("examples/capsule.obj")
	if err != nil {
		panic(err)
	}

	// load the texture
	texture, err := LoadTexture("examples/capsule.jpg")
	if err != nil {
		panic(err)
	}

	// fit mesh in a bi-unit cube centered at the origin
	mesh.BiUnitCube()

	// create a rendering context
	context := NewContext(width*scale, height*scale)

	// create transformation matrix and light direction
	aspect := float64(width) / float64(height)
	matrix := LookAt(eye, center, up).Perspective(fovy, aspect, near, far)

	// render
	shader := NewPhongShader(matrix, light, eye)
	shader.Texture = texture
	context.Shader = shader
	start := time.Now()
	context.DrawMesh(mesh)
	fmt.Println(time.Since(start))

	// downsample image for antialiasing
	image := context.Image()
	image = resize.Resize(width, height, image, resize.Bilinear)

	// save image
	SavePNG("out.png", image)
}
