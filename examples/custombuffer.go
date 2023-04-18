package main

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"time"

	. "github.com/fogleman/fauxgl"
)

const (
	scale  = 1
	width  = 256
	height = 256
	fovy   = 40
	near   = 1
	far    = 100
)

var (
	eye    = V(0, 0, 4)
	center = V(0, 0, 0)
	up     = V(0, 1, 0)
	light  = V(0.5, 0.5, 2).Normalize()
)

func RandomColor() Color {
	r := rand.Float64()
	g := rand.Float64()
	b := rand.Float64()
	return Color{r, g, b, 1}
}

var Directions = []Cell{
	{-1, 0, 0},
	{1, 0, 0},
	{0, -1, 0},
	{0, 1, 0},
	{0, 0, -1},
	{0, 0, 1},
}

type Cell struct {
	X, Y, Z int
}

func (c Cell) Add(d Cell) Cell {
	return Cell{c.X + d.X, c.Y + d.Y, c.Z + d.Z}
}

func (c Cell) Sub(d Cell) Cell {
	return Cell{c.X - d.X, c.Y - d.Y, c.Z - d.Z}
}

func (c Cell) Vector() Vector {
	return V(float64(c.X), float64(c.Y), float64(c.Z))
}

func (c Cell) Mesh() *Mesh {
	const s = 0.125
	mesh := NewLatLngSphere(15, 15)
	mesh.Transform(Scale(V(s, s, s)).Translate(c.Vector()))
	return mesh
}

type Grid struct {
	W, H, D int
	Cells   map[Cell]bool
}

func NewGrid(w, h, d int) *Grid {
	cells := make(map[Cell]bool)
	return &Grid{w, h, d, cells}
}

func (g *Grid) Load() float64 {
	capacity := g.W * g.H * g.D
	size := len(g.Cells)
	return float64(size) / float64(capacity)
}

func (g *Grid) RandomEmptyCell() Cell {
	for {
		c := Cell{rand.Intn(g.W), rand.Intn(g.H), rand.Intn(g.D)}
		if !g.Get(c) {
			g.Set(c)
			return c
		}
	}
}

func (g *Grid) Get(c Cell) bool {
	if c.X < 0 || c.Y < 0 || c.Z < 0 {
		return true
	}
	if c.X >= g.W || c.Y >= g.H || c.Z >= g.D {
		return true
	}
	return g.Cells[c]
}

func (g *Grid) Set(c Cell) {
	g.Cells[c] = true
}

func MakeSegment(p0, p1 Vector, r float64, c Color) *Mesh {
	p := p0.Add(p1).MulScalar(0.5)
	h := p0.Distance(p1)
	up := p1.Sub(p0).Normalize()
	mesh := NewCylinder(15, false)
	mesh.Transform(Orient(p, V(r, r, h), up, 0))
	return mesh
}

type Pipe struct {
	Cell      Cell
	Direction Cell
	Color     Color
	Done      bool
	Mesh      *Mesh
}

func NewPipe(cell Cell) *Pipe {
	direction := Cell{}
	color := RandomColor()
	return &Pipe{cell, direction, color, false, NewEmptyMesh()}
}

func (pipe *Pipe) Update(grid *Grid) {
	if pipe.Done {
		return
	}
	cells := make([]Cell, 0, 6)
	for _, d := range Directions {
		c := pipe.Cell.Add(d)
		if !grid.Get(c) {
			cells = append(cells, c)
		}
	}
	if len(cells) == 0 {
		pipe.Done = true
		return
	}
	c := cells[rand.Intn(len(cells))]
	d := c.Sub(pipe.Cell)
	if d != pipe.Direction {
		pipe.Mesh.Add(pipe.Cell.Mesh())
	}
	p0 := pipe.Cell.Vector()
	pipe.Cell = c
	p1 := pipe.Cell.Vector()
	pipe.Mesh.Add(MakeSegment(p0, p1, 0.125, pipe.Color))
	grid.Set(pipe.Cell)
	pipe.Direction = d
}

func (pipe *Pipe) GetMesh() *Mesh {
	mesh := pipe.Mesh.Copy()
	mesh.Add(pipe.Cell.Mesh())
	for _, t := range mesh.Triangles {
		t.V1.Color = pipe.Color
		t.V2.Color = pipe.Color
		t.V3.Color = pipe.Color
	}
	return mesh
}

type CustomPixel struct {
	Normal Vector
	Color  Color
}

type CustomBuffer struct {
	Height int
	Width  int
	Buffer []CustomPixel
}

func NewCustomBuffer(width, height int) *CustomBuffer {
	b := &CustomBuffer{}

	b.Width = width
	b.Height = height
	b.Buffer = make([]CustomPixel, width*height)

	return b
}

func (b *CustomBuffer) Clear(color CustomPixel) {
	for y := 0; y < b.Height; y++ {
		i := y * b.Width
		for x := 0; x < b.Width; x++ {
			b.Buffer[i] = color
			i += 1
		}
	}
}

func (b *CustomBuffer) Write(x, y int, color CustomPixel) {
	b.Buffer[y*b.Width+x] = color
}

func (b *CustomBuffer) Dimensions() (int, int) {
	return b.Width, b.Height
}

func (b *CustomBuffer) ColorImage() image.Image {
	im := image.NewNRGBA(image.Rect(0, 0, b.Width, b.Height))

	for y := 0; y < b.Height; y++ {
		i := y * b.Width
		for x := 0; x < b.Width; x++ {
			im.Set(x, y, b.Buffer[i].Color.NRGBA())
			i += 1
		}
	}

	return im
}

func (b *CustomBuffer) NormalImage() image.Image {
	im := image.NewNRGBA(image.Rect(0, 0, b.Width, b.Height))

	const d = 0xff
	for y := 0; y < b.Height; y++ {
		i := y * b.Width
		for x := 0; x < b.Width; x++ {
			v := b.Buffer[i].Normal

			r := Clamp(v.X, 0, 1)
			g := Clamp(v.Y, 0, 1)
			b := Clamp(v.Z, 0, 1)
			im.Set(x, y, color.NRGBA{uint8(r * d), uint8(g * d), uint8(b * d), uint8(d)})
			i += 1
		}
	}

	return im
}

type CustomShader struct {
	Matrix Matrix
}

func NewCustomShader(matrix Matrix) *CustomShader {
	return &CustomShader{matrix}
}

func (shader *CustomShader) Vertex(v Vertex) Vertex {
	v.Output = shader.Matrix.MulPositionW(v.Position)
	return v
}

func (shader *CustomShader) Fragment(v Vertex) CustomPixel {
	return CustomPixel{Normal: v.Normal, Color: Color{R: 1, G: 1, B: 0, A: 1}}
}

func main() {
	aspect := float64(width) / float64(height)
	matrix := LookAt(eye, center, up).Perspective(fovy, aspect, near, far)

	buffer := NewCustomBuffer(width*scale, height*scale)
	context := NewContext[*CustomBuffer, CustomPixel](buffer)
	context.ClearColor = CustomPixel{Color: Black}
	context.Shader = NewCustomShader(matrix)

	grid := NewGrid(19, 11, 11)
	pipes := make([]*Pipe, 8)
	for i := range pipes {
		pipes[i] = NewPipe(grid.RandomEmptyCell())
	}

	for grid.Load() < 0.9 {
		// mesh := NewEmptyMesh()
		dead := 0
		for _, pipe := range pipes {
			if pipe.Done {
				// mesh.Add(pipe.GetMesh())
				continue
			}
			pipe.Update(grid)
			// mesh.Add(pipe.GetMesh())
			if pipe.Done {
				dead++
			}
		}
		for j := 0; j < dead; j++ {
			pipes = append(pipes, NewPipe(grid.RandomEmptyCell()))
		}
		// mesh.Transform(Translate(V(-5, -5, -5)).Scale(V(0.2, 0.2, 0.2)))
		// mesh.SmoothNormals()

		// fmt.Println(i, len(pipes), len(mesh.Triangles))

		// context.ClearColorBuffer()
		// context.ClearDepthBuffer()
		// context.DrawMesh(mesh)

		// image := context.Image()
		// image = resize.Resize(width, height, image, resize.Bilinear)

		// SavePNG(fmt.Sprintf("frame%06d.png", i), image)
	}

	mesh := NewEmptyMesh()
	for _, pipe := range pipes {
		mesh.Add(pipe.GetMesh())
	}
	mesh.Transform(Translate(V(-9, -5, -5)).Scale(V(0.2, 0.2, 0.2)))
	mesh.SmoothNormals()

	fmt.Println(len(pipes), len(mesh.Triangles))

	context.ClearColorBuffer()
	context.ClearDepthBuffer()

	start := time.Now()
	context.DrawMesh(mesh)
	fmt.Println(time.Since(start))

	normalImage := context.ColorBuffer.NormalImage()
	colorImage := context.ColorBuffer.ColorImage()
	SavePNG("out_normal.png", normalImage)
	SavePNG("out_color.png", colorImage)
}
