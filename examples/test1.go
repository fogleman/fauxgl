package main

import (
	"fmt"

	"github.com/fogleman/gg"
	. "github.com/fogleman/soft/soft"
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
	mesh.SmoothNormalsThreshold(Radians(30))

	aspect := float64(width) / float64(height)
	matrix := LookAt(eye, center, up)
	matrix = matrix.Perspective(fovy, aspect, near, far)
	screen := Scale(V(1, -1, 1)).Translate(V(1, 1, 0)).Scale(V(width/2, height/2, 1))

	dc := gg.NewContext(width, height)
	dc.SetRGB(0, 0, 0)
	dc.Clear()

	depth := make([]float64, width*height)
	for i := range depth {
		depth[i] = 1
	}

	light := V(1, -0.5, 0.5).Normalize()

	for _, t := range mesh.Triangles {
		v1 := t.V1
		v2 := t.V2
		v3 := t.V3
		w1 := matrix.MulPositionW(v1)
		w2 := matrix.MulPositionW(v2)
		w3 := matrix.MulPositionW(v3)
		if !ClipBox.Contains(w1) {
			continue
		}
		if !ClipBox.Contains(w2) {
			continue
		}
		if !ClipBox.Contains(w3) {
			continue
		}
		s1 := screen.MulPosition(w1)
		s2 := screen.MulPosition(w2)
		s3 := screen.MulPosition(w3)
		s := NewTriangle(s1, s2, s3)
		fragments := s.Rasterize()
		for _, f := range fragments {
			x := int(f.X)
			y := int(f.Y)
			z := f.Z
			i := y*width + x
			if z > depth[i] {
				continue
			}
			depth[i] = z
			l := Clamp(t.NormalAt(f.Barycentric).Dot(light), 0, 1)
			dc.SetRGB(l, l, l)
			dc.SetPixel(x, y)
		}
	}
	dc.SavePNG("out.png")
}
