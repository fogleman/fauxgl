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

	. "github.com/fogleman/fauxgl"
)

var palette = colormap.New(colormap.ParseColors("67001fb2182bd6604df4a582fddbc7f7f7f7d1e5f092c5de4393c32166ac053061"))

// var palette = colormap.New(colormap.ParseColors("000000ffffff"))

const (
	pixelsPerMillimeter        = 20
	padding_mm                 = 1
	curvatureSamplingRadius_mm = 0.5
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

var Invalid = Vector{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64}

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
		a := math.Abs(v.Normal.Dot(V(0, 0, 1)))
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

type Segment struct {
	A, B Vector
}

func MarchingSquares(w, h int, data []float64, z float64, segments []Segment) []Segment {
	fraction := func(z0, z1, z float64) float64 {
		var f float64
		if z0 != z1 {
			f = (z - z0) / (z1 - z0)
		}
		f = math.Max(f, 0)
		f = math.Min(f, 1)
		return f
	}
	for y := 0; y < h-1; y++ {
		for x := 0; x < w-1; x++ {
			ul := data[x+y*w]
			ur := data[x+1+y*w]
			ll := data[x+y*w+w]
			lr := data[x+1+y*w+w]

			var squareCase int
			if ul > z {
				squareCase |= 1
			}
			if ur > z {
				squareCase |= 2
			}
			if ll > z {
				squareCase |= 4
			}
			if lr > z {
				squareCase |= 8
			}

			if squareCase == 0 || squareCase == 15 {
				continue
			}

			fx := float64(x)
			fy := float64(y)

			t := Vector{fx + fraction(ul, ur, z), fy, 0}
			b := Vector{fx + fraction(ll, lr, z), fy + 1, 0}
			l := Vector{fx, fy + fraction(ul, ll, z), 0}
			r := Vector{fx + 1, fy + fraction(ur, lr, z), 0}

			const connectHigh = false
			switch squareCase {
			case 1:
				segments = append(segments, Segment{t, l})
			case 2:
				segments = append(segments, Segment{r, t})
			case 3:
				segments = append(segments, Segment{r, l})
			case 4:
				segments = append(segments, Segment{l, b})
			case 5:
				segments = append(segments, Segment{t, b})
			case 6:
				if connectHigh {
					segments = append(segments, Segment{l, t})
					segments = append(segments, Segment{r, b})
				} else {
					segments = append(segments, Segment{r, t})
					segments = append(segments, Segment{l, b})
				}
			case 7:
				segments = append(segments, Segment{r, b})
			case 8:
				segments = append(segments, Segment{b, r})
			case 9:
				if connectHigh {
					segments = append(segments, Segment{t, r})
					segments = append(segments, Segment{b, l})
				} else {
					segments = append(segments, Segment{t, l})
					segments = append(segments, Segment{b, r})
				}
			case 10:
				segments = append(segments, Segment{b, t})
			case 11:
				segments = append(segments, Segment{b, l})
			case 12:
				segments = append(segments, Segment{l, r})
			case 13:
				segments = append(segments, Segment{t, r})
			case 14:
				segments = append(segments, Segment{l, t})
			}
		}
	}
	return segments
}

type CurvatureSamplerBuffers struct {
	Grid     []float64
	Segments []Segment
}

type CurvatureSampler struct {
	Width     int
	Height    int
	Radius    float64
	Matrix    Matrix
	HeightMap []Color
	NormalMap []Color

	Inverse   Matrix
	HalfWidth int
	Points    []Vector
	Normals   []Vector
}

func NewCurvatureSampler(width, height int, radius float64, matrix Matrix, heightMap, normalMap []Color) *CurvatureSampler {
	cs := &CurvatureSampler{}
	cs.Width = width
	cs.Height = height
	cs.Radius = radius
	cs.Matrix = matrix
	cs.HeightMap = heightMap
	cs.NormalMap = normalMap
	cs.Initialize()
	return cs
}

func (cs *CurvatureSampler) Initialize() {
	w, h := cs.Width, cs.Height

	cs.Inverse = cs.Matrix.Inverse()

	p0 := cs.Matrix.MulPosition(Vector{0, 0, 0})
	p1 := cs.Matrix.MulPosition(Vector{1, 0, 0})
	px_per_mm := p0.Distance(p1) * float64(w) / 2
	cs.HalfWidth = int(math.Ceil(cs.Radius * px_per_mm))

	cs.Points = make([]Vector, 0, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			cs.Points = append(cs.Points, cs.ComputePointAt(x, y))
		}
	}

	cs.Normals = make([]Vector, 0, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			cs.Normals = append(cs.Normals, cs.ComputeNormalAt(x, y))
		}
	}
}

func (cs *CurvatureSampler) ComputePointAt(px, py int) Vector {
	w, h := cs.Width, cs.Height
	if px < 0 || py < 0 || px >= w || py >= h {
		return Invalid
	}
	c := cs.HeightMap[py*w+px]
	x := float64(px)/float64(w-1)*2 - 1
	y := float64(py)/float64(h-1)*2 - 1
	z := c.R*2 - 1
	if c.A == 0 {
		z = 2
	}
	return cs.Inverse.MulPosition(Vector{x, 1 - y, z})
}

func (cs *CurvatureSampler) ComputeNormalAt(px, py int) Vector {
	w, h := cs.Width, cs.Height
	if px < 0 || py < 0 || px >= w || py >= h {
		return Invalid
	}
	c := cs.NormalMap[py*w+px]
	if c.A == 0 {
		return Invalid
	}
	return Vector{c.R*2 - 1, c.G*2 - 1, c.B*2 - 1}.Normalize()
}

func (cs *CurvatureSampler) PointAt(px, py int) Vector {
	if px < 0 || py < 0 || px >= cs.Width || py >= cs.Height {
		return Invalid
	}
	return cs.Points[py*cs.Width+px]
}

func (cs *CurvatureSampler) NormalAt(px, py int) Vector {
	if px < 0 || py < 0 || px >= cs.Width || py >= cs.Height {
		return Invalid
	}
	return cs.Normals[py*cs.Width+px]
}

func (cs *CurvatureSampler) Bilinear(x, y float64) Vector {
	w, h := cs.Width, cs.Height
	x0, y0 := int(math.Floor(x)), int(math.Floor(y))
	x1, y1 := x0+1, y0+1
	x -= float64(x0)
	y -= float64(y0)
	if x0 < 0 || y0 < 0 || x1 >= w || y1 >= h {
		return Invalid
	}
	p00 := cs.Points[y0*w+x0]
	p10 := cs.Points[y0*w+x1]
	p01 := cs.Points[y1*w+x0]
	p11 := cs.Points[y1*w+x1]
	if p00 == Invalid || p01 == Invalid || p10 == Invalid || p11 == Invalid {
		return Invalid
	}
	var v Vector
	v = v.Add(p00.MulScalar((1 - x) * (1 - y)))
	v = v.Add(p10.MulScalar(x * (1 - y)))
	v = v.Add(p01.MulScalar((1 - x) * y))
	v = v.Add(p11.MulScalar(x * y))
	return v
}

func (cs *CurvatureSampler) Sample(px, py int, buf *CurvatureSamplerBuffers) float64 {
	n := cs.NormalAt(px, py)
	if n == Invalid {
		return math.NaN()
	}
	p := cs.PointAt(px, py)

	dx := px - cs.HalfWidth
	dy := py - cs.HalfWidth
	f := func(x, y int) float64 {
		q := cs.PointAt(x+dx, y+dy)
		// TODO: make this a separate configurable threshold
		if p.Z-q.Z > cs.Radius*2 {
			return math.NaN()
		}
		d := p.Distance(q)
		return d
	}

	N := cs.HalfWidth*2 + 1
	if len(buf.Grid) != N*N {
		buf.Grid = make([]float64, N*N)
	}
	i := 0
	for y := 0; y < N; y++ {
		for x := 0; x < N; x++ {
			buf.Grid[i] = f(x, y)
			i++
		}
	}

	buf.Segments = buf.Segments[:0]
	buf.Segments = MarchingSquares(N, N, buf.Grid, cs.Radius, buf.Segments)

	var sum, total float64
	for _, segment := range buf.Segments {
		a := segment.A
		b := segment.B
		p0 := cs.Bilinear(a.X+float64(dx), a.Y+float64(dy))
		p1 := cs.Bilinear(b.X+float64(dx), b.Y+float64(dy))
		if p0 == Invalid || p1 == Invalid {
			continue
		}
		d0 := n.Dot(p0.Sub(p))
		d1 := n.Dot(p1.Sub(p))
		d := p0.Distance(p1)
		sum += (d0 + d1) / 2 * d
		total += d
	}

	if total == 0 {
		return math.NaN()
	}

	mean := sum / total
	return math.Atan(mean / cs.Radius)
}

func (cs *CurvatureSampler) Run() []Color {
	w, h := cs.Width, cs.Height
	result := make([]Color, w*h)

	var wg sync.WaitGroup
	wn := runtime.NumCPU()
	for wi := 0; wi < wn; wi++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := &CurvatureSamplerBuffers{}
			for py := 0; py < h; py++ {
				if py%wn != wi {
					continue
				}
				for px := 0; px < w; px++ {
					t := cs.Sample(px, py, buf)
					if !math.IsNaN(t) {
						result[py*w+px] = Color{t, t, t, 1}
					}
				}
			}
		}()
	}
	wg.Wait()

	return result
}

func NormalizeCurvatureImage(image []Color, min, max, gamma, percentile float64) {
	if min == 0 && max == 0 {
		max = -math.MaxFloat64
		for i := range image {
			max = math.Max(max, math.Abs(image[i].R))
		}
		min, max = -max, max
	}

	if percentile != 0 {
		var values []float64
		for i := range image {
			if image[i].A != 0 {
				values = append(values, math.Abs(image[i].R))
			}
		}
		sort.Float64s(values)
		max = values[int(math.Round(float64(len(values)-1)*percentile))]
		min = -max
	}

	for i := range image {
		c := image[i]
		if c.A == 0 {
			continue
		}
		t := (c.R-min)/(max-min)*2 - 1
		if gamma != 0 && gamma != 1 {
			if t < 0 {
				t = -math.Pow(-t, gamma)
			} else if t > 0 {
				t = math.Pow(t, gamma)
			}
		}
		t = (t + 1) / 2
		image[i] = Color{t, t, t, c.A}
	}
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

func makeCombinedImage(width, height int, angle, curvature []Color) *image.RGBA64 {
	im := image.NewRGBA64(image.Rect(0, 0, width, height))
	var i int
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// a := angle[i]
			// c := curvature[i]
			var e color.NRGBA64
			if angle[i].A == 0 {
				e = color.NRGBA64{0, 0, 0, 0xffff}
				// e = color.NRGBA64{}
			} else {
				a := angle[i].R
				c := curvature[i].R*2 - 1

				var r, g, b uint16
				g = uint16(math.Round(a * 0xffff))
				if c > 0 {
					c = math.Min(c, 1)
					b = uint16(math.Round(c * 0xffff))
				} else {
					c = math.Max(c, -1)
					r = uint16(math.Round(-c * 0xffff))
				}

				e = color.NRGBA64{r, g, b, 0xffff}
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
	size.X = 100
	size.Y = 100
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
		// SavePNG(fmt.Sprintf("%s-normal.png", filepath.Base(inputPath)), context.Image())
		normalMap := context.Buffer()

		context.ClearDepthBuffer()
		context.ClearColorBufferWith(Color{})
		context.Shader = NewMapShader(width, height, modeHeight, matrix, z0, z1, prevHeightMap)
		context.DrawMesh(mesh)
		// SavePNG(fmt.Sprintf("%s-height.png", filepath.Base(inputPath)), context.Image())
		heightMap := context.Buffer()

		// context.ClearDepthBuffer()
		// context.ClearColorBufferWith(Color{})
		// context.Shader = NewMapShader(width, height, modeAngle, matrix, z0, z1, prevHeightMap)
		// context.DrawMesh(mesh)
		// // SavePNG(fmt.Sprintf("%s-angle.png", filepath.Base(inputPath)), context.Image())
		// angleMap := context.Buffer()

		// curvatureMap := computeCurvatureMap(width, height, heightMap, normalMap, matrix)
		cs := NewCurvatureSampler(width, height, curvatureSamplingRadius_mm, matrix, heightMap, normalMap)
		curvatureMap := cs.Run()
		NormalizeCurvatureImage(curvatureMap, 0, 0, curvatureGamma, 0)
		SavePNG(fmt.Sprintf("%s-curvature.png", filepath.Base(inputPath)), makeImage(width, height, curvatureMap))

		// SavePNG(fmt.Sprintf("%s-combined.png", filepath.Base(inputPath)), makeCombinedImage(width, height, angleMap, curvatureMap))

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
