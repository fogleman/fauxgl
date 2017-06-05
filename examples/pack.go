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
	fovy   = 30   // vertical field of view in degrees
	near   = 1    // near clipping plane
	far    = 10   // far clipping plane
)

var (
	eye        = V(2, 5, 1.25)               // camera position
	center     = V(0, 0, 0)                  // view center position
	up         = V(0, 0, 1)                  // up vector
	light      = V(0.25, 0.5, 1).Normalize() // light direction
	color      = HexColor("#468966")         // object color
	background = HexColor("#FFF8E3")         // background color
)

func timed(name string) func() {
	if len(name) > 0 {
		fmt.Printf("%s... ", name)
	}
	start := time.Now()
	return func() {
		fmt.Println(time.Since(start))
	}
}

func main() {
	var done func()

	done = timed("loading mesh")
	mesh, err := LoadSTL("examples/boat.stl")
	if err != nil {
		panic(err)
	}
	done()

	fmt.Println(len(mesh.Triangles))

	done = timed("transforming mesh")
	mesh.BiUnitCube()
	done()

	done = timed("creating bvh")
	boxes := NewTreeForMesh(mesh)
	done()

	fmt.Println(len(boxes))

	// create a rendering context
	context := NewContext(width*scale, height*scale)
	context.ClearColorBufferWith(background)

	// create transformation matrix and light direction
	aspect := float64(width) / float64(height)
	matrix := LookAt(eye, center, up).Perspective(fovy, aspect, near, far)

	// render
	shader := NewPhongShader(matrix, light, eye)
	shader.ObjectColor = color
	context.Shader = shader
	done = timed("rendering mesh")
	context.DrawMesh(mesh)
	done()

	for _, box := range boxes {
		m := Translate(Vector{0.5, 0.5, 0.5})
		m = m.Scale(box.Size())
		m = m.Translate(box.Min)
		cube := NewCube()
		cube.Transform(m)
		context.DrawMesh(cube)
	}

	// context.Shader = NewSolidColorShader(matrix, Black)
	// context.LineWidth = 8
	// // context.DepthBias = -1e-4
	// for _, box := range boxes {
	// 	context.DrawLines(box.Outline())
	// }

	// downsample image for antialiasing
	done = timed("downsampling image")
	image := context.Image()
	image = resize.Resize(width, height, image, resize.Bilinear)
	done()

	// save image
	done = timed("writing output")
	SavePNG("out.png", image)
	done()
}
