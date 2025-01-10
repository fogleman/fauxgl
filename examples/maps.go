package main

import (
	"fmt"
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

type MapShader struct {
	Mode      int
	Matrix    Matrix
	Z0, Z1    float64
	HeightMap []Color
}

func NewMapShader(mode int, matrix Matrix, z0, z1 float64, heightmap []Color) *MapShader {
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

	if len(shader.HeightMap) > 0 {
		u := shader.HeightMap[v.Y*width+v.X]
		if u.R >= t {
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

func updateHeightMap(heightMap, update []Color) []Color {
	if len(heightMap) == 0 {
		return update
	}
	for i := range heightMap {
		if update[i].R > heightMap[i].R {
			heightMap[i] = update[i]
		}
	}
	return heightMap
}

func computeCurvatureMap(heightMap, normalMap []Color, z0, z1, xyScale float64) []Color {
	const c0 = -0.05
	const c1 = 0.05
	result := make([]Color, len(heightMap))
	for py := 0; py < height; py++ {
		for px := 0; px < width; px++ {
			i := py*width + px
			hm := heightMap[i]
			nm := normalMap[i]

			if hm.A == 0 {
				continue
			}

			x := float64(px) * xyScale
			y := float64(py) * xyScale
			z := z0 + (z1-z0)*hm.R
			p := Vector{x, y, z}

			nx := nm.R*2 - 1
			ny := nm.G*2 - 1
			nz := nm.B*2 - 1
			n := Vector{nx, ny, nz}.Normalize()

			m := RotateTo(Vector{0, 0, -1}, n)

			var sum float64
			var total float64
			for i := 0; i < 32; i++ {
				a := float64(i) / 32 * 2 * math.Pi
				dir := Vector{math.Cos(a), math.Sin(a), 0}
				s := m.MulDirection(dir).MulScalar(5 * xyScale).Add(p)
				sx := int(math.Round(s.X / xyScale))
				sy := int(math.Round(s.Y / xyScale))
				c := heightMap[sy*width+sx]
				if c.A == 0 {
					continue
				}
				sz := z0 + (z1-z0)*c.R
				q := Vector{float64(sx) * xyScale, float64(sy) * xyScale, sz}
				d := n.Dot(q.Sub(p))
				sum += d
				total++
			}
			meanDistance := sum / total
			t := (meanDistance - c0) / (c1 - c0)
			t = Clamp(t, 0, 1)
			result[i] = Color{t, t, t, 1}
		}
	}
	return result
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

	var prevHeightMap []Color

	for i := 0; i < 5; i++ {
		context.ClearDepthBuffer()
		context.ClearColorBufferWith(Color{})
		context.Shader = NewMapShader(modeNormal, matrix, z0, z1, prevHeightMap)
		context.DrawMesh(mesh)
		SavePNG(fmt.Sprintf("normal-%02d.png", i), context.Image())

		normalMap := context.Buffer()

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

		heightMap := context.Buffer()

		curvatureMap := computeCurvatureMap(heightMap, normalMap, z0, z1, xyScale)
		context.ColorBuffer = curvatureMap
		SavePNG(fmt.Sprintf("curvature-%02d.png", i), context.Image())

		prevHeightMap = updateHeightMap(prevHeightMap, heightMap)
	}

	return nil
}

func main() {
	inputPath := os.Args[1]
	run(inputPath)
}
