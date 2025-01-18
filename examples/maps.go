package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"runtime"
	"sort"
	"sync"

	"os"
	"path/filepath"

	"github.com/fogleman/colormap"
	"github.com/fogleman/contourmap"

	. "github.com/fogleman/fauxgl"
)

var palette = colormap.New(colormap.ParseColors("67001fb2182bd6604df4a582fddbc7f7f7f7d1e5f092c5de4393c32166ac053061"))

// var palette = colormap.New(colormap.ParseColors("000000ffffff"))

const (
	pixelsPerMillimeter        = 50
	padding_mm                 = 1
	curvatureSamplingRadius_mm = 0.25
	curvatureGamma             = 1
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

	p0 := matrix.MulPosition(Vector{0, 0, 0})
	p1 := matrix.MulPosition(Vector{1, 0, 0})
	px_per_mm := p0.Distance(p1) * float64(width) / 2
	inverse := matrix.Inverse()
	invalid := Vector{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64}
	ws := int(math.Ceil(curvatureSamplingRadius_mm * px_per_mm))

	computePointAt := func(px, py int) Vector {
		if px < 0 || py < 0 || px >= width || py >= height {
			return invalid
		}
		c := heightMap[py*width+px]
		// if c.A == 0 {
		// 	return invalid
		// }
		x := float64(px)/float64(width-1)*2 - 1
		y := float64(py)/float64(height-1)*2 - 1
		z := c.R*2 - 1
		if c.A == 0 {
			z = 2
		}
		return inverse.MulPosition(Vector{x, 1 - y, z})
	}

	allPoints := make([]Vector, 0, width*height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			allPoints = append(allPoints, computePointAt(x, y))
		}
	}

	pointAt := func(px, py int) Vector {
		if px < 0 || py < 0 || px >= width || py >= height {
			return invalid
		}
		return allPoints[py*width+px]
	}

	normalAt := func(px, py int) Vector {
		if px < 0 || py < 0 || px >= width || py >= height {
			return invalid
		}
		c := normalMap[py*width+px]
		if c.A == 0 {
			return invalid
		}
		return Vector{c.R*2 - 1, c.G*2 - 1, c.B*2 - 1}.Normalize()
	}

	bilinear := func(x, y float64) Vector {
		x0, y0 := int(math.Floor(x)), int(math.Floor(y))
		x1, y1 := x0+1, y0+1
		x -= float64(x0)
		y -= float64(y0)
		if x0 < 0 || y0 < 0 || x1 >= width || y1 >= height {
			return invalid
		}
		p00 := allPoints[y0*width+x0]
		p10 := allPoints[y0*width+x1]
		p01 := allPoints[y1*width+x0]
		p11 := allPoints[y1*width+x1]
		if p00 == invalid || p01 == invalid || p10 == invalid || p11 == invalid {
			return invalid
		}
		var v Vector
		v = v.Add(p00.MulScalar((1 - x) * (1 - y)))
		v = v.Add(p10.MulScalar(x * (1 - y)))
		v = v.Add(p01.MulScalar((1 - x) * y))
		v = v.Add(p11.MulScalar(x * y))
		return v
	}

	curvatureAt := func(px, py int) float64 {
		p := pointAt(px, py)
		n := normalAt(px, py)
		if n == invalid {
			return math.NaN()
		}

		dx := px - ws
		dy := py - ws
		f := func(x, y int) float64 {
			q := pointAt(x+dx, y+dy)
			// TODO: make this a separate configurable threshold
			if p.Z-q.Z > curvatureSamplingRadius_mm*2 {
				return math.NaN()
			}
			d := p.Distance(q)
			return d
		}
		m := contourmap.FromFunction(ws*2+1, ws*2+1, f)
		contours := m.Contours(curvatureSamplingRadius_mm)

		var sum, total float64
		for _, contour := range contours {
			for i := 1; i < len(contour); i++ {
				j := i - 1
				p0 := bilinear(contour[j].X+float64(dx), contour[j].Y+float64(dy))
				p1 := bilinear(contour[i].X+float64(dx), contour[i].Y+float64(dy))
				if p0 == invalid || p1 == invalid {
					continue
				}
				d0 := n.Dot(p0.Sub(p)) / curvatureSamplingRadius_mm
				d1 := n.Dot(p1.Sub(p)) / curvatureSamplingRadius_mm
				d := p0.Distance(p1)
				sum += (d0 + d1) / 2 * d
				total += d
			}
		}

		if total == 0 {
			return math.NaN()
		}

		mean := sum / total
		return math.Min(math.Max(mean, -1), 1)
	}

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
					t := curvatureAt(px, py)
					if !math.IsNaN(t) {
						result[py*width+px] = Color{t, t, t, 1}
					}
				}
			}
		}()
	}
	wg.Wait()

	min, max := -1.0, 1.0
	max = 0
	for i := range result {
		max = math.Max(max, math.Abs(result[i].R))
	}
	min, max = -max, max

	var values []float64
	for i := range result {
		if result[i].A != 0 {
			values = append(values, math.Abs(result[i].R))
		}
	}
	sort.Float64s(values)
	max = values[len(values)*999/1000]
	min = -max
	fmt.Println(min, max)

	for i := range result {
		c := result[i]
		if c.A == 0 {
			continue
		}
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
				// e = color.NRGBA64{}
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

	const r = 50
	mesh.FitInside(Box{Vector{-r, -r, -r}, Vector{r, r, r}}, Vector{0.5, 0.5, 0.5})
	mesh.Transform(Rotate(Vector{1, 0, 0}, -math.Pi/2))
	mesh.MoveTo(Vector{0, 0, 0}, Vector{0.5, 0.5, 0.5})
	mesh.Transform(Rotate(Vector{0, 1, 0}, angle))
	// mesh.Transform(Rotate(Vector{0, 1, 0}, math.Pi/4))
	// mesh.Transform(Rotate(RandomUnitVector(), rand.Float64()*3))
	// mesh.SmoothNormalsThreshold(Radians(30))

	box := mesh.BoundingBox()
	size := box.Size()
	// size.X = 70
	// size.Y = 48
	// fmt.Println(size)
	z0 := box.Min.Z
	z1 := box.Max.Z

	width := int(math.Ceil((size.X + padding_mm*2) * pixelsPerMillimeter))
	height := int(math.Ceil((size.Y + padding_mm*2) * pixelsPerMillimeter))
	width += width % 2
	height += height % 2

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
		SavePNG(fmt.Sprintf("%s-normal.png", filepath.Base(inputPath)), context.Image())
		normalMap := context.Buffer()

		context.ClearDepthBuffer()
		context.ClearColorBufferWith(Color{})
		context.Shader = NewMapShader(width, height, modeHeight, matrix, z0, z1, prevHeightMap)
		context.DrawMesh(mesh)
		SavePNG(fmt.Sprintf("%s-height.png", filepath.Base(inputPath)), context.Image())
		heightMap := context.Buffer()

		context.ClearDepthBuffer()
		context.ClearColorBufferWith(Color{})
		context.Shader = NewMapShader(width, height, modeAngle, matrix, z0, z1, prevHeightMap)
		context.DrawMesh(mesh)
		SavePNG(fmt.Sprintf("%s-angle.png", filepath.Base(inputPath)), context.Image())

		curvatureMap := computeCurvatureMap(width, height, heightMap, normalMap, matrix)
		SavePNG(fmt.Sprintf("%s-curvature.png", filepath.Base(inputPath)), makeImage(width, height, curvatureMap))

		prevHeightMap = updateHeightMap(prevHeightMap, heightMap)
	}

	return nil
}

func main() {
	for _, path := range os.Args[1:] {
		// fmt.Println(path)
		for i := 0; i < frames; i++ {
			// fmt.Println(i)
			run(path, i)
			break
		}
	}
}
