package main

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"os"
	"path/filepath"

	"github.com/fogleman/colormap"

	. "github.com/fogleman/fauxgl"
)

var palette = colormap.New(colormap.ParseColors("67001fb2182bd6604df4a582fddbc7f7f7f7d1e5f092c5de4393c32166ac053061"))

const (
	width  = 4000 / 2
	height = 3000 / 2
)

const (
	modeHeight = iota
	modeAngle
	modeNormal
)

var (
	eye    = V(0, 0, -10)
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
		// n.Z = -n.Z
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

func computeCurvatureMap(heightMap, normalMap []Color, matrix Matrix) []Color {
	inverse := matrix.Inverse()
	const c0 = -0.05
	const c1 = 0.05
	result := make([]Color, len(heightMap))
	for py := 0; py < height; py++ {
		for px := 0; px < width; px++ {
			i := py*width + px
			hm := heightMap[i]
			nm := normalMap[i]

			if nm.A == 0 {
				continue
			}

			p := inverse.MulPosition(Vector{
				float64(px)/(width-1)*2 - 1,
				float64(py)/(height-1)*2 - 1,
				hm.R,
			})

			nx := nm.R*2 - 1
			ny := nm.G*2 - 1
			nz := nm.B*2 - 1
			n := Vector{nx, ny, nz}.Normalize()

			m := RotateTo(Vector{0, 0, -1}, n)

			var sum float64
			var total float64
			for i := 0; i < 64; i++ {
				a := float64(i) / 64 * 2 * math.Pi
				dir := Vector{math.Cos(a), math.Sin(a), 0}
				s := matrix.MulPosition(p.Add(m.MulDirection(dir).MulScalar(0.25)))
				sx := int(math.Round((s.X + 1) / 2 * width))
				sy := int(math.Round((s.Y + 1) / 2 * height))
				if sx < 0 || sy < 0 || sx >= width || sy >= height {
					continue
				}
				c := heightMap[sy*width+sx]
				if c.A == 0 {
					continue
				}
				q := inverse.MulPosition(Vector{
					float64(sx)/(width-1)*2 - 1,
					float64(sy)/(height-1)*2 - 1,
					c.R,
				})
				d := n.Dot(q.Sub(p))
				if math.Abs(q.Z-p.Z) > 0.25 {
					continue
				}
				// if d > 0.1 {
				// 	continue
				// }
				sum += d
				total++
			}
			meanDistance := sum / total
			t := meanDistance
			result[i] = Color{t, t, t, nm.A}
		}
	}

	min := result[0].R
	max := result[0].R
	for _, c := range result {
		if c.R < min {
			min = c.R
		}
		if c.R > max {
			max = c.R
		}
	}
	if math.Abs(min) > math.Abs(max) {
		max = -min
		// min = -max
	} else {
		min = -max
		// max = -min
	}
	fmt.Println(min, max)
	for i := range result {
		c := result[i]
		result[i] = result[i].SubScalar(min).DivScalar(max - min)
		result[i].A = c.A
	}
	return result
}

func makeImage(buf []Color) *image.RGBA64 {
	im := image.NewRGBA64(image.Rect(0, 0, width, height))
	var i int
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := buf[i]
			d := palette.At(buf[i].R)
			r, g, b, _ := d.RGBA()
			e := color.NRGBA64{uint16(r), uint16(g), uint16(b), uint16(c.A * 0xffff)}
			im.Set(x, y, e)
			i++
		}
	}
	return im
}

func run(inputPath string) error {
	mesh, err := LoadMesh(inputPath)
	if err != nil {
		return err
	}

	mesh.MoveTo(Vector{}, Vector{0.5, 0.5, 0})
	// mesh.Transform(Rotate(RandomUnitVector(), rand.Float64()*3))
	mesh.SmoothNormalsThreshold(Radians(20))
	box := mesh.BoundingBox()
	z0 := box.Min.Z
	z1 := box.Max.Z

	const s = 30
	aspect := float64(width) / height
	// xyScale := float64(s*2) / height
	matrix := LookAt(eye, center, up).Orthographic(-s*aspect, s*aspect, -s, s, z0, z1)

	context := NewContext(width, height)

	var prevHeightMap []Color

	for i := 0; i < 3; i++ {
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

		curvatureMap := computeCurvatureMap(heightMap, normalMap, matrix)
		// SavePNG(fmt.Sprintf("curvature-%02d.png", i), makeImage(curvatureMap))
		path := fmt.Sprintf("%s-%02d.png", filepath.Base(inputPath), i)
		SavePNG(path, makeImage(curvatureMap))

		prevHeightMap = updateHeightMap(prevHeightMap, heightMap)
		// break
	}

	return nil
}

func main() {
	for _, path := range os.Args[1:] {
		fmt.Println(path)
		run(path)
	}
}
