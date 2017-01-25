package main

import (
	"fmt"
	"math"

	"github.com/fogleman/gg"
	. "github.com/fogleman/soft/soft"
)

const (
	width  = 1024 * 2
	height = 1024 * 2
	fovy   = 50
	near   = 1
	far    = 10
)

var (
	eye    = V(-1, -3, 1)
	center = V(0, 0, -0.25)
	up     = V(0, 0, 1)
	light  = V(1, -1, 0.5).Normalize()
)

var ClipBox = Box{V(-1, -1, -1), V(1, 1, 1)}

func main() {
	// mesh, err := LoadOBJ("examples/dragon.obj")
	mesh, err := LoadSTL("examples/bunny.stl")
	if err != nil {
		panic(err)
	}
	mesh.BiUnitCube()
	mesh.SmoothNormalsThreshold(Radians(30))

	for frame := 0; frame < 72; frame++ {
		fmt.Println(frame)

		aspect := float64(width) / float64(height)
		matrix := Rotate(up, Radians(float64(frame*5))).LookAt(eye, center, up).Perspective(fovy, aspect, near, far)
		screen := Scale(V(1, -1, 1)).Translate(V(1, 1, 0)).Scale(V(width/2, height/2, 1))

		light = V(1, -1, 0.5).Normalize()
		light = Rotate(up, Radians(-float64(frame*5))).MulDirection(light)

		dc := gg.NewContext(width, height)
		dc.SetRGB(1, 1, 1)
		dc.Clear()

		depth := make([]float64, width*height)
		for i := range depth {
			depth[i] = 1
		}

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
				l := Clamp(t.BarycentricNormal(f.Barycentric).Dot(light), 0, 1)
				if math.IsNaN(l) {
					continue
				}
				depth[i] = z
				dc.SetRGB(l*0.275, l*0.537, l*0.4)
				dc.SetPixel(x, y)
			}
		}

		dc.SavePNG(fmt.Sprintf("out%03d.png", frame))
	}
}
