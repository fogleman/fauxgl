package main

import (
	"fmt"
	"math/rand"

	. "github.com/fogleman/fauxgl"
	"github.com/nfnt/resize"
)

const (
	scale  = 4
	width  = 1024
	height = 1024
	fovy   = 40
	near   = 1
	far    = 100
)

var (
	eye    = V(4, 0, 0)
	center = V(0, 0, 0)
	up     = V(0, 0, 1)
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
	mesh := NewSphere(15, 15)
	mesh.Transform(Scale(V(s, s, s)).Translate(c.Vector()))
	return mesh
}

type Grid struct {
	Size  int
	Cells map[Cell]bool
}

func NewGrid(size int) *Grid {
	cells := make(map[Cell]bool)
	return &Grid{size, cells}
}

func (g *Grid) RandomEmptyCell() Cell {
	s := g.Size
	for {
		c := Cell{rand.Intn(s), rand.Intn(s), rand.Intn(s)}
		if !g.Get(c) {
			g.Set(c)
			return c
		}
	}
}

func (g *Grid) Get(c Cell) bool {
	s := g.Size
	if c.X < 0 || c.Y < 0 || c.Z < 0 {
		return true
	}
	if c.X > s || c.Y > s || c.Z > s {
		return true
	}
	return g.Cells[c]
}

func (g *Grid) Set(c Cell) {
	g.Cells[c] = true
}

func MakeSegment(p0, p1 Vector, r float64, c Color) *Mesh {
	p := p0.Add(p1).MulScalar(0.5)
	h := p0.Distance(p1) * 2
	up := p1.Sub(p0).Normalize()
	mesh := NewCylinder(15, false)
	mesh.Transform(Orient(p, V(r, r, h), up, 0))
	return mesh
}

type Pipe struct {
	Cell      Cell
	Direction Cell
	Color     Color
	Mesh      *Mesh
}

func NewPipe(cell Cell) *Pipe {
	direction := Cell{}
	color := RandomColor()
	return &Pipe{cell, direction, color, NewEmptyMesh()}
}

func (pipe *Pipe) Update(grid *Grid) bool {
	cells := make([]Cell, 0, 6)
	for _, d := range Directions {
		c := pipe.Cell.Add(d)
		if !grid.Get(c) {
			cells = append(cells, c)
		}
	}
	if len(cells) == 0 {
		return false
	}
	c := cells[rand.Intn(len(cells))]
	d := c.Sub(pipe.Cell)
	if d != pipe.Direction {
		pipe.Mesh.Add(pipe.Cell.Mesh())
	}
	p0 := pipe.Cell.Vector()
	pipe.Cell = c
	p1 := pipe.Cell.Vector()
	pipe.Mesh.Add(MakeSegment(p0, p1, 0.25, pipe.Color))
	grid.Set(pipe.Cell)
	pipe.Direction = d
	return true
}

func (pipe *Pipe) Terminate() {
	pipe.Mesh.Add(pipe.Cell.Mesh())
	for _, t := range pipe.Mesh.Triangles {
		t.V1.Color = pipe.Color
		t.V2.Color = pipe.Color
		t.V3.Color = pipe.Color
	}
}

func main() {
	grid := NewGrid(10)
	pipes := make([]*Pipe, 8)
	for i := range pipes {
		pipes[i] = NewPipe(grid.RandomEmptyCell())
	}
	mesh := NewEmptyMesh()
	for i := 0; i < 100; i++ {
		for j, pipe := range pipes {
			if !pipe.Update(grid) {
				pipe.Terminate()
				mesh.Add(pipe.Mesh)
				pipes[j] = NewPipe(grid.RandomEmptyCell())
			}
		}
	}
	for _, pipe := range pipes {
		pipe.Terminate()
		mesh.Add(pipe.Mesh)
	}

	fmt.Println(len(mesh.Triangles))
	mesh.BiUnitCube()
	mesh.SmoothNormals()

	context := NewContext(width*scale, height*scale)
	context.ClearColorBufferWith(Gray(0))

	aspect := float64(width) / float64(height)
	matrix := LookAt(eye, center, up).Perspective(fovy, aspect, near, far)
	light := V(2, 1, 1).Normalize()

	shader := NewPhongShader(matrix, light, eye)
	context.Shader = shader
	context.DrawMesh(mesh)

	image := context.Image()
	image = resize.Resize(width, height, image, resize.Bilinear)

	SavePNG("out.png", image)
}
