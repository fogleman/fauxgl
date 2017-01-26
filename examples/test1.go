package main

import (
	"fmt"
	"time"

	. "github.com/fogleman/soft/soft"
)

const (
	width  = 1024 * 1
	height = 1024 * 1
	fovy   = 50
	near   = 1
	far    = 10
)

var (
	eye    = V(-1, -3, 1)
	center = V(0, 0, -0.25)
	up     = V(0, 0, 1)
)

func main() {
	// load a mesh
	mesh, err := LoadSTL("examples/bunny.stl")
	if err != nil {
		panic(err)
	}

	// fit mesh in a bi-unit cube centered at the origin
	mesh.BiUnitCube()

	// smooth the normals
	mesh.SmoothNormalsThreshold(Radians(30))

	// create a rendering context
	context := NewContext(width, height)

	for frame := 0; frame < 72; frame++ {
		start := time.Now()

		// clear depth and color buffers (black)
		context.ClearDepthBuffer()
		context.ClearColorBuffer(V(0, 0, 0))

		angle := Radians(float64(frame * 5))
		aspect := float64(width) / float64(height)

		// create transformation matrix and light direction
		matrix := Rotate(up, angle).LookAt(eye, center, up).Perspective(fovy, aspect, near, far)
		light := Rotate(up, -angle).MulDirection(V(1, -1, 0.5).Normalize())
		color := V(0.275, 0.537, 0.4)

		// render
		shader := NewDefaultShader(matrix, light, color)
		context.DrawMesh(mesh, shader)

		elapsed := time.Since(start)
		fmt.Println(frame, elapsed)

		// save image
		SavePNG(fmt.Sprintf("out%03d.png", frame), context.Image())
	}
}
