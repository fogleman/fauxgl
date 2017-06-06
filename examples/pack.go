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
	fovy   = 30   // vertical field of view in degrees
	near   = 1    // near clipping plane
	far    = 10   // far clipping plane
)

var (
	eye        = V(4, 4, 2)                  // camera position
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
	mesh, err := LoadSTL(os.Args[1])
	if err != nil {
		panic(err)
	}
	done()

	fmt.Println(len(mesh.Triangles))

	done = timed("transforming mesh")
	mesh.BiUnitCube()
	mesh.Transform(Scale(V(60, 60, 60)))
	done()

	model := NewPackModel()
	model.Add(mesh, 4)
	fmt.Println(model.Valid())
	fmt.Println(model.Volume())
	fmt.Println(model.Energy())

	model = Anneal(model, 10, 0.001, 1000000).(*PackModel)

	mesh = NewEmptyMesh()
	// cubes := NewEmptyMesh()
	for _, item := range model.Items {
		m := item.Mesh.Copy()
		m.Transform(item.Matrix())
		mesh.Add(m)
		// for _, box := range item.Tree.Leaves(-1) {
		// 	m := Translate(Vector{0.5, 0.5, 0.5})
		// 	m = m.Scale(box.Size())
		// 	m = m.Translate(box.Min)
		// 	cube := NewCube()
		// 	cube.Transform(m)
		// 	cubes.Add(cube)
		// }
	}

	done = timed("transforming mesh")
	mesh.BiUnitCube()
	// cubes.BiUnitCube()
	done()

	context := NewContext(width*scale, height*scale)
	context.ClearColorBufferWith(background)

	aspect := float64(width) / float64(height)
	matrix := LookAt(eye, center, up).Perspective(fovy, aspect, near, far)

	shader := NewPhongShader(matrix, light, eye)
	shader.ObjectColor = color
	context.Shader = shader
	done = timed("rendering mesh")
	context.DrawMesh(mesh)
	done()

	done = timed("downsampling image")
	image := context.Image()
	image = resize.Resize(width, height, image, resize.Bilinear)
	done()

	done = timed("writing output")
	SavePNG("out.png", image)
	done()
}
