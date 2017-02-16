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
	fovy   = 25   // vertical field of view in degrees
	near   = 1    // near clipping plane
	far    = 100  // far clipping plane
)

var (
	eye        = V(3, 3, 3)                 // camera position
	center     = V(0, 0, -0.1)              // view center position
	up         = V(0, 0, 1)                 // up vector
	light      = V(0.1, 0.3, 1).Normalize() // light direction
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
	context.LineWidth = scale * 1
	context.DepthBias = -4e-5
	// context.DrawLines(mesh.Lines)
	context.DrawTriangles(mesh.Triangles)

	// downsample image for antialiasing
	image := context.Image()
	image = resize.Resize(width, height, image, resize.Bilinear)

	// save image
	SavePNG("out.png", image)
}
