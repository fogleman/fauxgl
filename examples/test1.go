package main

import (
	"github.com/fogleman/gg"
	. "github.com/fogleman/soft/soft"
)

const (
	width  = 1024
	height = 1024
	fovy   = 45
	near   = 1
	far    = 10
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
	mesh.BiUnitCube()
	mesh.SmoothNormalsThreshold(Radians(30))

	aspect := float64(width) / float64(height)
	matrix := LookAt(eye, center, up).Perspective(fovy, aspect, near, far)
	screen := Scale(V(1, -1, 1)).Translate(V(1, 1, 0)).Scale(V(width/2, height/2, 1))

	dc := gg.NewContext(width, height)
	// dc.SetRGB(0, 0, 0)
	// dc.Clear()

	depth := make([]float64, width*height)
	for i := range depth {
		depth[i] = 1
	}

	light := V(1, -0.5, 0.5).Normalize()

	for _, t := range mesh.Triangles {
		w1 := matrix.MulPositionW(t.V1)
		if !ClipBox.Contains(w1) {
			continue
		}
		w2 := matrix.MulPositionW(t.V2)
		if !ClipBox.Contains(w2) {
			continue
		}
		w3 := matrix.MulPositionW(t.V3)
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
			l := Clamp(t.BarycentricNormal(f.Barycentric).Dot(light), 0, 1)
			dc.SetRGB(l, l, l)
			dc.SetPixel(x, y)
		}
	}

	dc.SavePNG("out.png")
}
