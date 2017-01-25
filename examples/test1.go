package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"time"

	. "github.com/fogleman/soft/soft"
)

const (
	width  = 1024 * 1
	height = 1024 * 1
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

	var fragments []Fragment
	im := image.NewRGBA(image.Rect(0, 0, width, height))
	depth := make([]float64, width*height)

	for frame := 0; frame < 72; frame++ {
		start := time.Now()

		aspect := float64(width) / float64(height)
		matrix := Rotate(up, Radians(float64(frame*5))).LookAt(eye, center, up).Perspective(fovy, aspect, near, far)
		screen := Scale(V(1, -1, 1)).Translate(V(1, 1, 0)).Scale(V(width/2, height/2, 1))

		light = V(1, -1, 0.5).Normalize()
		light = Rotate(up, Radians(-float64(frame*5))).MulDirection(light)

		src := image.NewUniform(color.White)
		draw.Draw(im, im.Bounds(), src, image.ZP, draw.Src)

		for i := range depth {
			depth[i] = 1
		}

		s := Triangle{}
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
			s.V1 = screen.MulPosition(w1)
			s.V2 = screen.MulPosition(w2)
			s.V3 = screen.MulPosition(w3)
			fragments = s.Rasterize(fragments)
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
				c := color.RGBA{uint8(255 * l * 0.275), uint8(255 * l * 0.537), uint8(255 * l * 0.4), 255}
				im.SetRGBA(x, y, c)
			}
		}

		elapsed := time.Since(start)
		fmt.Println(frame, elapsed)

		SavePNG(fmt.Sprintf("out%03d.png", frame), im)
	}
}
