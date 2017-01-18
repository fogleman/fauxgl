package main

import (
	"fmt"

	"github.com/fogleman/gg"
	. "github.com/fogleman/sr/sr"
)

const (
	width  = 1024
	height = 1024
	fovy   = 45
	near   = 0.1
	far    = 100
)

var (
	eye    = V(-1, -3, 1)
	center = V(0, 0, -0.25)
	up     = V(0, 0, 1)
)

var ClipBox = Box{Vector{-1, -1, -1}, Vector{1, 1, 1}}

func main() {
	mesh, err := LoadSTL("examples/bunny.stl")
	if err != nil {
		panic(err)
	}
	fmt.Println(len(mesh.Triangles))
	mesh.BiUnitCube()

	aspect := float64(width) / float64(height)
	matrix := LookAt(eye, center, up)
	matrix = matrix.Perspective(fovy, aspect, near, far)
	// matrix = matrix.Orthographic(-2, 2, -2, 2, -2, 2)
	screen := Translate(Vector{1, 1, 0}).Scale(Vector{width / 2, height / 2, 0})

	dc := gg.NewContext(width, height)
	dc.InvertY()
	dc.SetRGB(1, 1, 1)
	dc.Clear()

	type Key struct {
		V1, V2 Vector
	}
	seen := make(map[Key]bool)

	for _, t := range mesh.Triangles {
		v1 := t.V1
		v2 := t.V2
		v3 := t.V3
		w1 := matrix.MulPositionW(v1)
		w2 := matrix.MulPositionW(v2)
		w3 := matrix.MulPositionW(v3)
		// fmt.Println(w1)
		// fmt.Println(w2)
		// fmt.Println(w3)
		// fmt.Println()
		// if !ClipBox.Contains(w1) {
		// 	continue
		// }
		// if !ClipBox.Contains(w2) {
		// 	continue
		// }
		// if !ClipBox.Contains(w3) {
		// 	continue
		// }
		s1 := screen.MulPosition(w1)
		s2 := screen.MulPosition(w2)
		s3 := screen.MulPosition(w3)
		if _, ok := seen[Key{s1, s2}]; !ok {
			dc.DrawLine(s1.X, s1.Y, s2.X, s2.Y)
			seen[Key{s1, s2}] = true
		}
		if _, ok := seen[Key{s1, s3}]; !ok {
			dc.DrawLine(s1.X, s1.Y, s3.X, s3.Y)
			seen[Key{s1, s3}] = true
		}
		if _, ok := seen[Key{s2, s3}]; !ok {
			dc.DrawLine(s2.X, s2.Y, s3.X, s3.Y)
			seen[Key{s2, s3}] = true
		}
	}
	dc.SetLineWidth(0.2)
	dc.SetRGB(0, 0, 0)
	dc.Stroke()
	dc.SavePNG("out.png")
}
