package main

import (
	"fmt"
	"image"

	. "github.com/fogleman/fauxgl"
	"github.com/nfnt/resize"
)

const (
	scale  = 4
	width  = 2048
	height = 2048
	fovy   = 30
	near   = 1
	far    = 10
)

var (
	eye    = V(-2, -4, 2)
	center = V(-0.1, 0, -0.1)
	up     = V(0, 0, 1)
)

func render(mesh *Mesh) image.Image {
	context := NewContext(width*scale, height*scale)
	context.ClearColorBufferWith(White)

	aspect := float64(width) / float64(height)
	matrix := LookAt(eye, center, up).Perspective(fovy, aspect, near, far)
	light := V(-0.75, -0.25, 1).Normalize()

	shader := NewPhongShader(matrix, light, eye)
	shader.ObjectColor = HexColor("FFD34E")
	shader.DiffuseColor = Gray(0.9)
	shader.SpecularColor = Gray(0.25)
	shader.SpecularPower = 100
	context.Shader = shader
	context.DrawMesh(mesh)

	context.Shader = NewSolidColorShader(matrix, Black)
	context.DepthBias = -1e-4
	context.Wireframe = true
	context.LineWidth = 4
	context.DrawMesh(mesh)

	image := context.Image()
	image = resize.Resize(width, height, image, resize.Bilinear)
	return image
}

func main() {
	mesh, err := LoadSTL("examples/bunny.stl")
	if err != nil {
		panic(err)
	}
	mesh.Transform(Matrix{0.023175793856519147, 0, 0, 0, 0, 0.023175793856519147, 0, 0, 0, 0, 0.023175793856519147, -0.9704647076255632, 0, 0, 0, 1})
	fmt.Println(len(mesh.Triangles))
	f := 1.0
	for i := 1; ; i++ {
		m := mesh.Copy()
		m.Simplify(f)
		fmt.Println(i, f, len(m.Triangles))
		image := render(m)
		SavePNG(fmt.Sprintf("bunny/out%02d.png", i), image)
		f *= 0.75
	}
}
