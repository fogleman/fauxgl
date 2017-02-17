package main

import (
	"fmt"
	"os"
	"time"

	. "github.com/fogleman/fauxgl"
	"github.com/nfnt/resize"
)

const (
	scale  = 2    // optional supersampling
	width  = 2048 // output width in pixels
	height = 2048 // output height in pixels
	fovy   = 10   // vertical field of view in degrees
	near   = 1    // near clipping plane
	far    = 100  // far clipping plane
)

var (
	eye        = V(3, 3, 3)                 // camera position
	center     = V(0.15, 0, 0.3)            // view center position
	up         = V(0, 0, 1)                 // up vector
	light      = V(0.5, 0.3, 1).Normalize() // light direction
	color      = HexColor("#468966")        // object color
	background = HexColor("#FFF8E3")        // background color
)

func main() {
	// create a rendering context
	context := NewContext(width*scale, height*scale)
	context.ClearColorBufferWith(background)

	mesh, err := LoadSTL(os.Args[1])
	if err != nil {
		panic(err)
	}
	MeshEdges(mesh)
	mesh.BiUnitCube()
	mesh.Transform(Rotate(up, Radians(0)))
	// mesh.SmoothNormals()
	fmt.Println(len(mesh.Triangles))

	// create transformation matrix and light direction
	aspect := float64(width) / float64(height)
	matrix := LookAt(eye, center, up).Perspective(fovy, aspect, near, far)

	// render
	shader := NewPhongShader(matrix, light, eye)
	// shader.ObjectColor = color
	context.Shader = shader
	start := time.Now()
	context.DrawTriangles(mesh.Triangles)
	fmt.Println(time.Since(start))

	context.Shader = NewSolidColorShader(matrix, Black)
	context.Wireframe = true
	context.DepthBias = -1e-4
	context.LineWidth = scale / 2.0
	context.DrawTriangles(mesh.Triangles)
	context.LineWidth = scale * 3
	context.DrawLines(mesh.Lines)

	// downsample image for antialiasing
	image := context.Image()
	image = resize.Resize(width, height, image, resize.Bilinear)

	// save image
	SavePNG("out.png", image)
}
