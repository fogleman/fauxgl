package main

import (
	"fmt"
	"math/rand"
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
	// mesh, err := LoadOBJ("examples/dragon.obj")
	mesh, err := LoadSTL("examples/bunny.stl")
	if err != nil {
		panic(err)
	}
	mesh.BiUnitCube()
	mesh.SmoothNormalsThreshold(Radians(30))

	colors := make(map[Vector]Vector)
	for _, t := range mesh.Triangles {
		for _, v := range []Vertex{t.V1, t.V2, t.V3} {
			c := rand.Float64()
			colors[v.Position] = V(0.275, 0.537, 0.4).MulScalar(1.8 * c)
		}
	}
	for _, t := range mesh.Triangles {
		t.V1.Color = colors[t.V1.Position]
		t.V2.Color = colors[t.V2.Position]
		t.V3.Color = colors[t.V3.Position]
	}

	context := NewContext(width, height)

	for frame := 0; frame < 72; frame++ {
		start := time.Now()

		context.ClearDepthBuffer()
		context.ClearColorBuffer(V(0, 0, 0))

		angle := Radians(float64(frame * 5))
		aspect := float64(width) / float64(height)

		matrix := Rotate(up, angle).LookAt(eye, center, up).Perspective(fovy, aspect, near, far)
		light := Rotate(up, -angle).MulDirection(V(1, -1, 0.5).Normalize())
		color := V(0.275, 0.537, 0.4)

		shader := NewDefaultShader(matrix, light, color)
		context.DrawMesh(mesh, shader)

		elapsed := time.Since(start)
		fmt.Println(frame, elapsed)

		SavePNG(fmt.Sprintf("out%03d.png", frame), context.Image())
	}
}
