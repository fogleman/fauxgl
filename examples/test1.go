package main

import (
	"fmt"
	"time"

	. "github.com/fogleman/fauxgl"
)

const (
	width  = 1920 * 1
	height = 1080 * 1
	fovy   = 50
	near   = 1
	far    = 50
)

var (
	eye    = V(0, -7, 2)
	center = V(0, 0, -1.8)
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

		for x := -2; x <= 2; x++ {
			for y := -2; y <= 2; y++ {
				// create transformation matrix and light direction
				t := V(float64(x)*2, float64(y)*2, 0)
				matrix := Rotate(up, angle).Translate(t).LookAt(eye, center, up).Perspective(fovy, aspect, near, far)
				light := Rotate(up, -angle).MulDirection(V(-0.25, -0.25, 1).Normalize())
				color := HexColor(0xFFFAD5)
				camera := Translate(t.Negate()).MulPosition(eye)

				// render
				shader := NewDefaultShader(matrix, light, camera, color)
				context.DrawMesh(mesh, shader)
			}
		}

		elapsed := time.Since(start)
		fmt.Println(frame, elapsed)

		// save image
		SavePNG(fmt.Sprintf("out%03d.png", frame), context.Image())
	}
}
