package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"

	. "github.com/fogleman/fauxgl"
)

const (
	width  = 1600
	height = 1200
)

const (
	modeHeight = iota
	modeAngle
	modeNormal
)

var (
	eye    = V(0, 0, -1)
	center = V(0, 0, 0)
	up     = V(0, -1, 0)
)

func toGray(im image.Image) *image.Gray {
	dst := image.NewGray(im.Bounds())
	draw.Draw(dst, im.Bounds(), im, image.ZP, draw.Src)
	return dst
}

func toRGBA(im image.Image) *image.RGBA {
	dst := image.NewRGBA(im.Bounds())
	draw.Draw(dst, im.Bounds(), im, image.ZP, draw.Src)
	return dst
}

type MapShader struct {
	Mode      int
	Matrix    Matrix
	Z0, Z1    float64
	HeightMap *image.Gray
}

func NewMapShader(mode int, matrix Matrix, z0, z1 float64, heightmap *image.Gray) *MapShader {
	return &MapShader{mode, matrix, z0, z1, heightmap}
}

func (shader *MapShader) Vertex(v Vertex) Vertex {
	v.Output = shader.Matrix.MulPositionW(v.Position)
	return v
}

func (shader *MapShader) Fragment(v Vertex) Color {
	z := v.Position.Z
	t := (z - shader.Z0) / (shader.Z1 - shader.Z0)
	t = Clamp(t, 0, 1)

	if shader.HeightMap != nil {
		p := shader.HeightMap.GrayAt(v.X, v.Y).Y
		u := float64(p+1) / 255
		if u >= t {
			return Discard
		}
	}

	if shader.Mode == modeNormal {
		n := v.Normal
		n = n.AddScalar(1).DivScalar(2)
		return Color{n.X, n.Y, n.Z, 1}
	}

	if shader.Mode == modeAngle {
		a := 1 - math.Abs(v.Normal.Dot(V(0, 0, 1)))
		return Color{a, a, a, 1}
	}

	return Color{t, t, t, 1}
}

func updateHeightMap(heightMap, update *image.Gray) *image.Gray {
	if heightMap == nil {
		return update
	}
	for i := range heightMap.Pix {
		if update.Pix[i] > heightMap.Pix[i] {
			heightMap.Pix[i] = update.Pix[i]
		}
	}
	return heightMap
}

func computeCurvatureMap(heightMap *image.RGBA, normalMap *image.RGBA, z0, z1, xyScale float64) *image.RGBA {
	size := heightMap.Bounds().Size()
	w := size.X
	h := size.Y
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	const c0 = -0.05
	const c1 = 0.05
	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			hm := heightMap.RGBAAt(px, py)
			nm := normalMap.RGBAAt(px, py)

			if hm.A == 0 {
				continue
			}

			x := float64(px) * xyScale
			y := float64(py) * xyScale
			z := z0 + (z1-z0)*(float64(hm.R)/255)
			p := Vector{x, y, z}

			nx := (float64(nm.R)/255)*2 - 1
			ny := (float64(nm.G)/255)*2 - 1
			nz := (float64(nm.B)/255)*2 - 1
			n := Vector{nx, ny, nz}.Normalize()

			m := RotateTo(Vector{0, 0, -1}, n)

			var sum float64
			for i := 0; i < 32; i++ {
				a := float64(i) / 32 * 2 * math.Pi
				dir := Vector{math.Cos(a), math.Sin(a), 0}
				s := m.MulDirection(dir).MulScalar(5 * xyScale).Add(p)
				sx := int(math.Round(s.X / xyScale))
				sy := int(math.Round(s.Y / xyScale))
				sz := z0 + (z1-z0)*(float64(heightMap.RGBAAt(sx, sy).R)/255)
				q := Vector{float64(sx) * xyScale, float64(sy) * xyScale, sz}
				d := n.Dot(q.Sub(p))
				sum += d

			}
			meanDistance := sum / 32
			t := (meanDistance - c0) / (c1 - c0)
			t = Clamp(t, 0, 1)
			c := uint8(math.Round(t * 255))
			dst.SetRGBA(px, py, color.RGBA{c, c, c, 255})
		}
	}
	return dst
}

func run(inputPath string) error {
	mesh, err := LoadMesh(inputPath)
	if err != nil {
		return err
	}

	mesh.BiUnitCube()
	mesh.SmoothNormalsThreshold(Radians(30))
	box := mesh.BoundingBox()
	z0 := box.Min.Z
	z1 := box.Max.Z

	const s = 1
	aspect := float64(width) / height
	xyScale := float64(s*2) / height
	matrix := LookAt(eye, center, up).Orthographic(-s*aspect, s*aspect, -s, s, z0-1, z1+1)

	context := NewContext(width, height)

	var prevHeightMap *image.Gray

	for i := 0; i < 5; i++ {
		context.ClearDepthBuffer()
		context.ClearColorBufferWith(Color{})
		context.Shader = NewMapShader(modeNormal, matrix, z0, z1, prevHeightMap)
		context.DrawMesh(mesh)
		SavePNG(fmt.Sprintf("normal-%02d.png", i), context.Image())

		normalMap := toRGBA(context.Image())

		context.ClearDepthBuffer()
		context.ClearColorBufferWith(Color{})
		context.Shader = NewMapShader(modeAngle, matrix, z0, z1, prevHeightMap)
		context.DrawMesh(mesh)
		SavePNG(fmt.Sprintf("angle-%02d.png", i), context.Image())

		context.ClearDepthBuffer()
		context.ClearColorBufferWith(Color{})
		context.Shader = NewMapShader(modeHeight, matrix, z0, z1, prevHeightMap)
		context.DrawMesh(mesh)
		SavePNG(fmt.Sprintf("height-%02d.png", i), context.Image())

		heightMap := toRGBA(context.Image())

		curvatureMap := computeCurvatureMap(heightMap, normalMap, z0, z1, xyScale)
		SavePNG(fmt.Sprintf("curvature-%02d.png", i), curvatureMap)

		prevHeightMap = updateHeightMap(prevHeightMap, toGray(heightMap))
	}

	return nil
}

func main() {
	inputPath := os.Args[1]
	run(inputPath)
}
