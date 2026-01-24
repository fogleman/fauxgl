package main

import (
	"fmt"
	"image/png"
	"os"

	"github.com/fogleman/fauxgl"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("usage: %s <mesh file> <output file>\n", os.Args[0])
		return
	}

	mesh, err := fauxgl.LoadMesh(os.Args[1])
	if err != nil {
		fmt.Println("failed loading mesh:", err.Error())
		return
	}

	mesh.Transform(fauxgl.Rotate(fauxgl.V(0, 1, 0), fauxgl.Radians(45)))
	mesh.Transform(fauxgl.Rotate(fauxgl.V(0, 0, 1), fauxgl.Radians(30)))
	orientFrontFacing(mesh)
	mesh.BiUnitCube()

	backside := mesh.Copy()
	backside.ReverseWinding()

	mat := fauxgl.LookAt(fauxgl.V(3, 0, 0), fauxgl.V(0.0, 0.0, 0.0), fauxgl.V(0, 0, 1)).Orthographic(-1.1, 1.1, -1.1, 1.1, 1, 50)
	context := fauxgl.NewContext(1920, 1920)
	context.Cull = fauxgl.CullBack
	context.ClearColorBufferWith(fauxgl.HexColor("#00000000"))
	shader := fauxgl.NewPhongShader(mat, fauxgl.V(1, 0, 0), fauxgl.V(17, 0, 0))
	shader.ObjectColor = fauxgl.Gray(0.5).Opaque()
	shader.AmbientColor = fauxgl.Gray(0.25).Opaque()
	shader.DiffuseColor = fauxgl.Gray(1).Opaque()
	shader.SpecularColor = fauxgl.Gray(0.15).Opaque()
	shader.SpecularPower = 32
	context.Shader = shader
	context.DrawMesh(mesh)
	context.DrawMesh(backside)

	edges := mesh.SharpEdges(fauxgl.Radians(70))
	context.Shader = fauxgl.NewSolidColorShader(mat, fauxgl.Black)
	context.DrawMesh(edges)

	img := context.Image()
	out, err := os.Create(os.Args[2])
	if err != nil {
		fmt.Println("failed creating output file:", err.Error())
		return
	}
	defer out.Close()

	if err := png.Encode(out, img); err != nil {
		fmt.Println("failed writing output file:", err.Error())
		return
	}
}

func orientFrontFacing(mesh *fauxgl.Mesh) {
	mesh.Transform(fauxgl.Matrix{
		1, 0, 0, 0,
		0, 0, -1, 0,
		0, 1, 0, 0,
		0, 0, 0, 1,
	})
}
