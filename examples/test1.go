package main

import (
	"fmt"
	"image"
	"image/color"

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

var ClipBox = Box{V(-1, -1, -1), V(1, 1, 1)}

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
	screen := Scale(V(1, -1, 0)).Translate(V(1, 1, 0)).Scale(V(width/2, height/2, 0))

	dc := gg.NewContext(width, height)
	dc.SetRGB(1, 1, 1)
	dc.Clear()

	for _, t := range mesh.Triangles {
		v1 := t.V1
		v2 := t.V2
		v3 := t.V3
		w1 := matrix.MulPositionW(v1)
		w2 := matrix.MulPositionW(v2)
		w3 := matrix.MulPositionW(v3)
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
		t := NewTriangle(s1, s2, s3)
		lines := t.Rasterize()
		c := color.NRGBA{0, 0, 0, 64}
		DrawScanlines(dc.Image().(*image.RGBA), c, lines)
	}
	dc.SavePNG("out.png")
}
