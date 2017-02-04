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

func axisAligned(p1, p2 Vector) bool {
	count := 0
	if p1.X == p2.X {
		count++
	}
	if p1.Y == p2.Y {
		count++
	}
	if p1.Z == p2.Z {
		count++
	}
	return count >= 2
}

func main() {
	// load a mesh
	mesh, err := LoadVOX(os.Args[1])
	if err != nil {
		panic(err)
	}
	fmt.Println(len(mesh.Triangles))

	// fit mesh in a bi-unit cube centered at the origin
	mesh.BiUnitCube()

	var lines []*Line
	for _, t := range mesh.Triangles {
		p1 := t.V1.Position
		p2 := t.V2.Position
		p3 := t.V3.Position
		if axisAligned(p1, p2) {
			lines = append(lines, NewLine(t.V1, t.V2))
		}
		if axisAligned(p2, p3) {
			lines = append(lines, NewLine(t.V2, t.V3))
		}
		if axisAligned(p3, p1) {
			lines = append(lines, NewLine(t.V3, t.V1))
		}
	}

	// create a rendering context
	context := NewContext(width*scale, height*scale)

	// create transformation matrix and light direction
	// aspect := float64(width) / float64(height)
	// matrix := LookAt(eye, center, up).Perspective(fovy, aspect, near, far)
	matrix := LookAt(eye, center, up).Orthographic(-1.45, 1.45, -1.45, 1.45, -20, 20)

	// render
	context.ClearColorBufferWith(HexColor("323"))
	ambient := Color{0.4, 0.4, 0.4, 1}
	diffuse := Color{0.9, 0.9, 0.9, 1}
	context.Shader = NewDiffuseShader(matrix, light, Discard, ambient, diffuse)
	start := time.Now()
	context.DrawMesh(mesh)
	fmt.Println(time.Since(start))

	context.Shader = NewSolidColorShader(matrix, HexColor("000"))
	context.Wireframe = true
	context.LineWidth = 8
	context.DepthBias = -1e-4
	start = time.Now()
	context.DrawLines(lines)
	fmt.Println(time.Since(start))

	// downsample image for antialiasing
	image := context.Image()
	image = resize.Resize(width, height, image, resize.Bilinear)

	// save image
	SavePNG("out.png", image)
}
