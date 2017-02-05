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
	center = V(0, 0, -0.25)                 // view center position
	up     = V(0, 0, 1)                     // up vector
	light  = V(-0.25, -0.75, 1).Normalize() // light direction
)

func main() {
	// load a mesh
	voxels, err := LoadVOX(os.Args[1])
	if err != nil {
		panic(err)
	}

	mesh := NewVoxelMesh(voxels)

	fmt.Println(len(voxels), "voxels")
	fmt.Println(len(mesh.Triangles), "triangles")
	fmt.Println(len(mesh.Lines), "lines")

	// fit mesh in a bi-unit cube centered at the origin
	mesh.BiUnitCube()

	// create a rendering context
	context := NewContext(width*scale, height*scale)

	// create transformation matrix and light direction
	// aspect := float64(width) / float64(height)
	// matrix := LookAt(eye, center, up).Perspective(fovy, aspect, near, far)
	const s = 1.5
	matrix := LookAt(eye, center, up).Orthographic(-s, s, -s, s, -20, 20)

	// render
	context.ClearColorBufferWith(HexColor("323"))
	ambient := Color{0.4, 0.4, 0.4, 1}
	diffuse := Color{0.9, 0.9, 0.9, 1}
	context.Shader = NewDiffuseShader(matrix, light, Discard, ambient, diffuse)
	start := time.Now()
	context.DrawTriangles(mesh.Triangles)
	fmt.Println(time.Since(start))

	context.Shader = NewSolidColorShader(matrix, HexColor("000"))
	context.Wireframe = true
	context.LineWidth = scale * 2
	context.DepthBias = -1e-4
	start = time.Now()
	context.DrawLines(mesh.Lines)
	fmt.Println(time.Since(start))

	// downsample image for antialiasing
	image := context.Image()
	image = resize.Resize(width, height, image, resize.Bilinear)

	// save image
	SavePNG("out.png", image)
}
