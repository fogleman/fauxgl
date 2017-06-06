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
	mesh, err := LoadSTL(os.Args[1])
	if err != nil {
		panic(err)
	}
	done()

	fmt.Println(len(mesh.Triangles))

	done = timed("transforming mesh")
	mesh.BiUnitCube()
	mesh.Transform(Rotate(up, Radians(180)))
	done()

	const N = 10

	done = timed("creating bvh")
	node := NewTreeForMesh(mesh, N)
	done()

	for i := 0; i <= N; i++ {
		boxes := node.Leaves(i)
		fmt.Println(i, len(boxes))

		// for _, a := range boxes {
		// 	for _, b := range boxes {
		// 		if a != b && a.ContainsBox(b) {
		// 			fmt.Println("!!!", a, b)
		// 		}
		// 	}
		// }

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
		// done = timed("rendering mesh")
		// context.DrawMesh(mesh)
		// done()

		cubes := NewEmptyMesh()
		for _, box := range boxes {
			// fmt.Println(box.Volume(), box.Size())
			m := Translate(Vector{0.5, 0.5, 0.5})
			m = m.Scale(box.Size())
			m = m.Translate(box.Min)
			cube := NewCube()
			cube.Transform(m)
			cubes.Add(cube)
		}
		context.DrawMesh(cubes)
		// cubes.SaveSTL(fmt.Sprintf("out%03d.stl", i))

		// context.Shader = NewSolidColorShader(matrix, Black)
		// context.LineWidth = 8
		// // context.DepthBias = -1e-4
		// for _, box := range boxes {
		// 	context.DrawLines(box.Outline())
		// }

		// downsample image for antialiasing
		// done = timed("downsampling image")
		image := context.Image()
		image = resize.Resize(width, height, image, resize.Bilinear)
		// done()

		// save image
		// done = timed("writing output")
		SavePNG(fmt.Sprintf("out%03da.png", i), image)
		// done()
	}
}
