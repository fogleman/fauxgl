package fauxgl

import (
	"image"
	"image/draw"
	"math"
	"runtime"
	"sync"
)

type Context struct {
	Width        int
	Height       int
	ColorBuffer  *image.NRGBA
	DepthBuffer  []float64
	ClearColor   Color
	Shader       Shader
	ReadDepth    bool
	WriteDepth   bool
	WriteColor   bool
	Wireframe    bool
	screenMatrix Matrix
	locks        []sync.Mutex
}

func NewContext(width, height int) *Context {
	dc := &Context{}
	dc.Width = width
	dc.Height = height
	dc.ColorBuffer = image.NewNRGBA(image.Rect(0, 0, width, height))
	dc.DepthBuffer = make([]float64, width*height)
	dc.ClearColor = Transparent
	dc.Shader = nil // TODO: default
	dc.ReadDepth = true
	dc.WriteDepth = true
	dc.WriteColor = true
	dc.Wireframe = false
	dc.screenMatrix = Screen(width, height)
	dc.locks = make([]sync.Mutex, 256)
	dc.ClearDepthBuffer()
	return dc
}

func (dc *Context) Image() image.Image {
	return dc.ColorBuffer
}

func (dc *Context) ClearColorBufferWith(color Color) {
	im := dc.ColorBuffer
	src := image.NewUniform(color.NRGBA())
	draw.Draw(im, im.Bounds(), src, image.ZP, draw.Src)
}

func (dc *Context) ClearColorBuffer() {
	dc.ClearColorBufferWith(dc.ClearColor)
}

func (dc *Context) ClearDepthBufferWith(value float64) {
	for i := range dc.DepthBuffer {
		dc.DepthBuffer[i] = value
	}
}

func (dc *Context) ClearDepthBuffer() {
	dc.ClearDepthBufferWith(math.MaxFloat64)
}

func (dc *Context) line(s0, s1 Vector, color Color) {
	c := color.NRGBA()
	x0 := Round(s0.X)
	y0 := Round(s0.Y)
	x1 := Round(s1.X)
	y1 := Round(s1.Y)
	steep := false
	if AbsInt(x0-x1) < AbsInt(y0-y1) {
		steep = true
		x0, y0 = y0, x0
		x1, y1 = y1, x1
	}
	if x0 > x1 {
		x0, x1 = x1, x0
		y0, y1 = y1, y0
	}
	dx := x1 - x0
	dy := y1 - y0
	de2 := AbsInt(dy) * 2
	e2 := 0
	y := y0
	for x := x0; x <= x1; x++ {
		if steep {
			dc.ColorBuffer.SetNRGBA(y, x, c)
		} else {
			dc.ColorBuffer.SetNRGBA(x, y, c)
		}
		e2 += de2
		if e2 > dx {
			if y1 > y0 {
				y++
			} else {
				y--
			}
			e2 -= dx * 2
		}
	}
}

func edge(a, b, c Vector) float64 {
	return (b.X-c.X)*(a.Y-c.Y) - (b.Y-c.Y)*(a.X-c.X)
}

func (dc *Context) rasterize(v0, v1, v2 Vertex, s0, s1, s2 Vector) {
	min := s0.Min(s1.Min(s2)).Floor()
	max := s0.Max(s1.Max(s2)).Ceil()
	x0 := int(min.X)
	x1 := int(max.X)
	y0 := int(min.Y)
	y1 := int(max.Y)
	p := Vector{float64(x0) + 0.5, float64(y0) + 0.5, 0}
	w00 := edge(s1, s2, p)
	w01 := edge(s2, s0, p)
	w02 := edge(s0, s1, p)
	a01 := s1.Y - s0.Y
	b01 := s0.X - s1.X
	a12 := s2.Y - s1.Y
	b12 := s1.X - s2.X
	a20 := s0.Y - s2.Y
	b20 := s2.X - s0.X
	ra := 1 / edge(s0, s1, s2)
	r0 := 1 / v0.Output.W
	r1 := 1 / v1.Output.W
	r2 := 1 / v2.Output.W
	for y := y0; y <= y1; y++ {
		w0 := w00
		w1 := w01
		w2 := w02
		for x := x0; x <= x1; x++ {
			b0 := w0 * ra
			b1 := w1 * ra
			b2 := w2 * ra
			w0 += a12
			w1 += a20
			w2 += a01
			if b0 < 0 || b1 < 0 || b2 < 0 {
				continue
			}
			z := b0*s0.Z + b1*s1.Z + b2*s2.Z
			i := y*dc.Width + x
			if z >= dc.DepthBuffer[i] { // completely safe?
				continue
			}
			b := VectorW{b0 * r0, b1 * r1, b2 * r2, 0}
			b.W = 1 / (b.X + b.Y + b.Z)
			v := InterpolateVertexes(v0, v1, v2, b)
			color := dc.Shader.Fragment(v)
			if color == Discard {
				continue
			}
			c := color.NRGBA()
			lock := &dc.locks[(x+y)&255]
			lock.Lock()
			if z < dc.DepthBuffer[i] {
				dc.DepthBuffer[i] = z
				dc.ColorBuffer.SetNRGBA(x, y, c)
			}
			lock.Unlock()
		}
		w00 += b12
		w01 += b20
		w02 += b01
	}
}

func (dc *Context) drawClipped(v0, v1, v2 Vertex) {
	ndc0 := v0.Output.DivScalar(v0.Output.W).Vector()
	ndc1 := v1.Output.DivScalar(v1.Output.W).Vector()
	ndc2 := v2.Output.DivScalar(v2.Output.W).Vector()

	// back face culling
	if !dc.Wireframe {
		a := (ndc1.X-ndc0.X)*(ndc2.Y-ndc0.Y) - (ndc2.X-ndc0.X)*(ndc1.Y-ndc0.Y)
		if a <= 0 {
			// cw
			return
		}
	}

	s0 := dc.screenMatrix.MulPosition(ndc0)
	s1 := dc.screenMatrix.MulPosition(ndc1)
	s2 := dc.screenMatrix.MulPosition(ndc2)
	if dc.Wireframe {
		color := Black
		dc.line(s0, s1, color)
		dc.line(s1, s2, color)
		dc.line(s2, s0, color)
	} else {
		dc.rasterize(v0, v1, v2, s0, s1, s2)
	}
}

func (dc *Context) DrawTriangle(t *Triangle) {
	v1 := dc.Shader.Vertex(t.V1)
	v2 := dc.Shader.Vertex(t.V2)
	v3 := dc.Shader.Vertex(t.V3)
	if v1.Outside() || v2.Outside() || v3.Outside() {
		triangles := ClipTriangle(NewTriangle(v1, v2, v3))
		for _, t := range triangles {
			dc.drawClipped(t.V1, t.V2, t.V3)
		}
	} else {
		dc.drawClipped(v1, v2, v3)
	}
}

func (dc *Context) DrawTriangles(triangles []*Triangle) {
	wn := runtime.NumCPU()
	done := make(chan bool, wn)
	for wi := 0; wi < wn; wi++ {
		go func(wi int) {
			for i, t := range triangles {
				if i%wn == wi {
					dc.DrawTriangle(t)
				}
			}
			done <- true
		}(wi)
	}
	for wi := 0; wi < wn; wi++ {
		<-done
	}
}

func (dc *Context) DrawMesh(mesh *Mesh) {
	dc.DrawTriangles(mesh.Triangles)
}
