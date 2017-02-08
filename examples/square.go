package main

import (
	"fmt"
	"time"

	. "github.com/fogleman/fauxgl"
	"github.com/nfnt/resize"
)

const (
	scale  = 1    // optional supersampling
	width  = 1920 // output width in pixels
	height = 1080 // output height in pixels
	fovy   = 40   // vertical field of view in degrees
	near   = 1    // near clipping plane
	far    = 10   // far clipping plane
)

var (
	eye    = V(0, -2, 1.4)             // camera position
	center = V(0, 0, -0.2)             // view center position
	up     = V(0, 0, 1)                // up vector
	light  = V(0, -0.5, 1).Normalize() // light direction
	color  = HexColor("#468966")       // object color
)

func main() {
	// load the mesh
	mesh, err := LoadOBJ("examples/square.obj")
	if err != nil {
		panic(err)
	}

	// load the texture
	texture, err := LoadTexture("examples/texture.png")
	if err != nil {
		panic(err)
	}

	// fit mesh in a bi-unit cube centered at the origin
	mesh.BiUnitCube()

	// create a rendering context
	context := NewContext(width*scale, height*scale)
	context.ClearColor = White
	context.ClearColorBuffer()

	// create transformation matrix and light direction
	aspect := float64(width) / float64(height)
	matrix := LookAt(eye, center, up).Perspective(fovy, aspect, near, far)

	// render
	shader := NewTextureShader(matrix, texture)
	shader.Texture = texture
	start := time.Now()
	context.Shader = shader
	context.DrawMesh(mesh)
	fmt.Println(time.Since(start))

	// downsample image for antialiasing
	image := context.Image()
	image = resize.Resize(width, height, image, resize.Bilinear)

	// save image
	SavePNG("out.png", image)
}
