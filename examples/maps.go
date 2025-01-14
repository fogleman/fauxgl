package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"runtime"
	"sync"

	"os"
	"path/filepath"

	"github.com/fogleman/colormap"

	. "github.com/fogleman/fauxgl"
)

var palette = colormap.New(colormap.ParseColors("67001fb2182bd6604df4a582fddbc7f7f7f7d1e5f092c5de4393c32166ac053061"))

const (
	pixelsPerMillimeter        = 20
	padding_mm                 = 1
	curvatureSamplingRadius_mm = 0.5
	curvatureGamma             = 0.8
	frames                     = 180
)

const (
	modeHeight = iota
	modeAngle
	modeNormal
)

var (
	eye    = V(0, 0, 0)
	center = V(0, 0, 1)
	up     = V(0, -1, 0)
)

type MapShader struct {
	Width     int
	Height    int
	Mode      int
	Matrix    Matrix
	Z0, Z1    float64
	HeightMap []Color
}

func NewMapShader(width, height, mode int, matrix Matrix, z0, z1 float64, heightmap []Color) *MapShader {
	return &MapShader{width, height, mode, matrix, z0, z1, heightmap}
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
		u := shader.HeightMap[v.Y*shader.Width+v.X]
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

func computeCurvatureMap(width, height int, heightMap, normalMap []Color, matrix Matrix) []Color {
	result := make([]Color, len(heightMap))
	inverse := matrix.Inverse()

	p0 := matrix.MulPosition(Vector{0, 0, 0})
	p1 := matrix.MulPosition(Vector{1, 0, 0})
	px_per_mm := p0.Distance(p1) * float64(width) / 2

	curvatureSampleCount := int(math.Ceil(2 * math.Pi * curvatureSamplingRadius_mm * px_per_mm / 2))

	var wg sync.WaitGroup
	wn := runtime.NumCPU()
	for wi := 0; wi < wn; wi++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for py := 0; py < height; py++ {
				if py%wn != wi {
					continue
				}
				for px := 0; px < width; px++ {
					// if px%20 != 0 || py%20 != 0 {
					// 	continue
					// }

					i := py*width + px
					hm := heightMap[i]
					nm := normalMap[i]

					if hm.A == 0 {
						continue
					}

					p := inverse.MulPosition(Vector{
						float64(px)/float64(width-1)*2 - 1,
						float64(py)/float64(height-1)*2 - 1,
						hm.R*2 - 1,
					})

					n := Vector{nm.R*2 - 1, nm.G*2 - 1, nm.B*2 - 1}.Normalize()

					m := RotateTo(Vector{0, 0, -1}, n)

					var sum float64
					var total float64
					for i := 0; i < curvatureSampleCount; i++ {
						a := float64(i) / float64(curvatureSampleCount) * 2 * math.Pi
						dir := Vector{math.Cos(a), math.Sin(a), 0}
						offset := m.MulDirection(dir).MulScalar(curvatureSamplingRadius_mm * px_per_mm)
						sx := px + int(math.Round(offset.X))
						sy := py + int(math.Round(offset.Y))
						if sx < 0 || sy < 0 || sx >= width || sy >= height {
							continue
						}
						// result[sy*width+sx] = Color{1, 0, 0, 1}
						// continue
						hm := heightMap[sy*width+sx]
						q := inverse.MulPosition(Vector{
							float64(sx)/float64(width-1)*2 - 1,
							float64(sy)/float64(height-1)*2 - 1,
							hm.R*2 - 1,
						})
						if hm.A == 0 {
							q.Z = p.Z + curvatureSamplingRadius_mm
						} else {
							if q.Z > p.Z+curvatureSamplingRadius_mm {
								q.Z = p.Z + curvatureSamplingRadius_mm
							}
						}
						d := n.Dot(q.Sub(p))
						if hm.A != 0 {
							if q.Z < p.Z-curvatureSamplingRadius_mm*1 {
								d = 0
								// total += 1
							}
						}
						t := d / curvatureSamplingRadius_mm
						t = math.Max(t, -1)
						t = math.Min(t, 1)
						sum += t
						total++
					}
					t := sum / total
					result[i] = Color{t, t, t, nm.A}
				}
			}
		}()
	}
	wg.Wait()

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
	// fmt.Println(min, max)
	if math.Abs(min) > math.Abs(max) {
		max = -min
	} else {
		min = -max
	}
	// min, max = min/2, max/2
	min, max = -1, 1
	for i := range result {
		c := result[i]
		t := (c.R-min)/(max-min)*2 - 1
		if t < 0 {
			t = -math.Pow(-t, curvatureGamma)
		} else if t > 0 {
			t = math.Pow(t, curvatureGamma)
		}
		t = (t + 1) / 2
		result[i] = Color{t, t, t, c.A}
	}
	return result
}

func makeImage(width, height int, buf []Color) *image.RGBA64 {
	im := image.NewRGBA64(image.Rect(0, 0, width, height))
	var i int
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := buf[i]
			d := palette.At(buf[i].R)
			r, g, b, _ := d.RGBA()
			e := color.NRGBA64{uint16(r), uint16(g), uint16(b), uint16(c.A * 0xffff)}
			if c.A == 0 {
				e = color.NRGBA64{0xffff, 0xffff, 0xffff, 0xffff}
			}
			im.Set(x, y, e)
			i++
		}
	}
	return im
}

func run(inputPath string, frame int) error {
	mesh, err := LoadMesh(inputPath)
	if err != nil {
		return err
	}

	t := float64(frame) / frames
	angle := -t * 2 * math.Pi

	mesh.Transform(Rotate(Vector{1, 0, 0}, -math.Pi/2))
	mesh.MoveTo(Vector{5, 0, 0}, Vector{0.5, 0.5, 0.5})
	mesh.Transform(Rotate(Vector{0, 1, 0}, angle))

	// mesh.Transform(Rotate(Vector{0, 1, 0}, math.Pi/4))

	// mesh.Transform(Rotate(RandomUnitVector(), rand.Float64()*3))
	// mesh.SmoothNormalsThreshold(Radians(40))
	box := mesh.BoundingBox()
	size := box.Size()
	size.X = 70
	size.Y = 48
	fmt.Println(size)
	z0 := box.Min.Z
	z1 := box.Max.Z

	width := int(math.Ceil((size.X + padding_mm*2) * pixelsPerMillimeter))
	height := int(math.Ceil((size.Y + padding_mm*2) * pixelsPerMillimeter))
	if width%2 != 0 {
		width++
	}
	if height%2 != 0 {
		height++
	}

	aspect := float64(width) / float64(height)
	s := size.Y/2 + padding_mm
	matrix := LookAt(eye, center, up).Orthographic(-s*aspect, s*aspect, -s, s, z0, z1)

	context := NewContext(width, height)

	var prevHeightMap []Color

	for i := 0; i < 1; i++ {
		context.ClearDepthBuffer()
		context.ClearColorBufferWith(Color{})
		context.Shader = NewMapShader(width, height, modeNormal, matrix, z0, z1, prevHeightMap)
		context.DrawMesh(mesh)
		// SavePNG(fmt.Sprintf("%s-normal.png", filepath.Base(inputPath)), context.Image())

		normalMap := context.Buffer()

		// context.ClearDepthBuffer()
		// context.ClearColorBufferWith(Color{})
		// context.Shader = NewMapShader(width, height, modeAngle, matrix, z0, z1, prevHeightMap)
		// context.DrawMesh(mesh)
		// SavePNG(fmt.Sprintf("%s-angle.png", filepath.Base(inputPath)), context.Image())

		context.ClearDepthBuffer()
		context.ClearColorBufferWith(Color{})
		context.Shader = NewMapShader(width, height, modeHeight, matrix, z0, z1, prevHeightMap)
		context.DrawMesh(mesh)
		// SavePNG(fmt.Sprintf("%s-height.png", filepath.Base(inputPath)), context.Image())

		heightMap := context.Buffer()

		curvatureMap := computeCurvatureMap(width, height, heightMap, normalMap, matrix)
		SavePNG(fmt.Sprintf("%s-curvature-%08d.png", filepath.Base(inputPath), frame), makeImage(width, height, curvatureMap))

		prevHeightMap = updateHeightMap(prevHeightMap, heightMap)
		// break
	}

	return nil
}

func main() {
	for _, path := range os.Args[1:] {
		fmt.Println(path)
		for i := 0; i < frames; i++ {
			fmt.Println(i)
			run(path, i)
			// break
		}
	}
}
