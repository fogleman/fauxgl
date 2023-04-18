package fauxgl

import (
	"image"
	"image/color"
	"math"
	"runtime"
	"sync"
)

type Face int

const (
	_ Face = iota
	FaceCW
	FaceCCW
)

type Cull int

const (
	_ Cull = iota
	CullNone
	CullFront
	CullBack
)

type RasterizeInfo struct {
	TotalPixels   uint64
	UpdatedPixels uint64
}

func (info RasterizeInfo) Add(other RasterizeInfo) RasterizeInfo {
	return RasterizeInfo{
		info.TotalPixels + other.TotalPixels,
		info.UpdatedPixels + other.UpdatedPixels,
	}
}

type Buffer[T comparable] interface {
	Clear(T)
	Write(x, y int, val T)
	Dimensions() (int, int)
}

type Context[B Buffer[T], T comparable] struct {
	Width        int
	Height       int
	ColorBuffer  B
	DepthBuffer  []float64
	ClearColor   T
	Shader       Shader[T]
	ReadDepth    bool
	WriteDepth   bool
	WriteColor   bool
	AlphaBlend   bool
	Wireframe    bool
	FrontFace    Face
	Cull         Cull
	LineWidth    float64
	DepthBias    float64
	Discard      T
	screenMatrix Matrix
	locks        []sync.Mutex
}

func NewContext[B Buffer[T], T comparable](buffer B) *Context[B, T] {
	var tZero T

	width, height := buffer.Dimensions()

	dc := &Context[B, T]{}
	dc.Width = width
	dc.Height = height
	dc.ColorBuffer = buffer
	dc.DepthBuffer = make([]float64, width*height)
	dc.ClearColor = tZero
	dc.Shader = NewZeroShader[T](Identity())
	dc.ReadDepth = true
	dc.WriteDepth = true
	dc.WriteColor = true
	dc.AlphaBlend = true
	dc.Wireframe = false
	dc.FrontFace = FaceCCW
	dc.Cull = CullBack
	dc.LineWidth = 2
	dc.DepthBias = 0
	dc.Discard = tZero
	dc.screenMatrix = Screen(width, height)
	dc.locks = make([]sync.Mutex, 256)
	dc.ClearDepthBuffer()
	return dc
}

func (dc *Context[B, T]) DepthImage() image.Image {
	lo := math.MaxFloat64
	hi := -math.MaxFloat64
	for _, d := range dc.DepthBuffer {
		if d == math.MaxFloat64 {
			continue
		}
		if d < lo {
			lo = d
		}
		if d > hi {
			hi = d
		}
	}

	im := image.NewGray16(image.Rect(0, 0, dc.Width, dc.Height))
	var i int
	for y := 0; y < dc.Height; y++ {
		for x := 0; x < dc.Width; x++ {
			d := dc.DepthBuffer[i]
			t := (d - lo) / (hi - lo)
			if d == math.MaxFloat64 {
				t = 1
			}
			c := color.Gray16{uint16(t * 0xffff)}
			im.SetGray16(x, y, c)
			i++
		}
	}
	return im
}

func (dc *Context[B, T]) ClearColorBufferWith(color T) {
	dc.ColorBuffer.Clear(color)
}

func (dc *Context[B, T]) ClearColorBuffer() {
	dc.ClearColorBufferWith(dc.ClearColor)
}

func (dc *Context[B, T]) ClearDepthBufferWith(value float64) {
	for i := range dc.DepthBuffer {
		dc.DepthBuffer[i] = value
	}
}

func (dc *Context[B, T]) ClearDepthBuffer() {
	dc.ClearDepthBufferWith(math.MaxFloat64)
}

func edge(a, b, c Vector) float64 {
	return (b.X-c.X)*(a.Y-c.Y) - (b.Y-c.Y)*(a.X-c.X)
}

func (dc *Context[B, T]) rasterize(v0, v1, v2 Vertex, s0, s1, s2 Vector) RasterizeInfo {
	var info RasterizeInfo

	// integer bounding box
	min := s0.Min(s1.Min(s2)).Floor()
	max := s0.Max(s1.Max(s2)).Ceil()
	x0 := int(min.X)
	x1 := int(max.X)
	y0 := int(min.Y)
	y1 := int(max.Y)

	// forward differencing variables
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

	// reciprocals
	ra := 1 / edge(s0, s1, s2)
	r0 := 1 / v0.Output.W
	r1 := 1 / v1.Output.W
	r2 := 1 / v2.Output.W
	ra12 := 1 / a12
	ra20 := 1 / a20
	ra01 := 1 / a01

	// iterate over all pixels in bounding box
	for y := y0; y <= y1; y++ {
		var d float64
		d0 := -w00 * ra12
		d1 := -w01 * ra20
		d2 := -w02 * ra01
		if w00 < 0 && d0 > d {
			d = d0
		}
		if w01 < 0 && d1 > d {
			d = d1
		}
		if w02 < 0 && d2 > d {
			d = d2
		}
		d = float64(int(d))
		if d < 0 {
			// occurs in pathological cases
			d = 0
		}
		w0 := w00 + a12*d
		w1 := w01 + a20*d
		w2 := w02 + a01*d
		wasInside := false
		for x := x0 + int(d); x <= x1; x++ {
			b0 := w0 * ra
			b1 := w1 * ra
			b2 := w2 * ra
			w0 += a12
			w1 += a20
			w2 += a01
			// check if inside triangle
			if b0 < 0 || b1 < 0 || b2 < 0 {
				if wasInside {
					break
				}
				continue
			}
			wasInside = true
			// check depth buffer for early abort
			i := y*dc.Width + x
			if i < 0 || i >= len(dc.DepthBuffer) {
				// TODO: clipping roundoff error; fix
				// TODO: could also be from fat lines going off screen
				continue
			}
			info.TotalPixels++
			z := b0*s0.Z + b1*s1.Z + b2*s2.Z
			bz := z + dc.DepthBias
			if dc.ReadDepth && bz > dc.DepthBuffer[i] { // safe w/out lock?
				continue
			}
			// perspective-correct interpolation of vertex data
			b := VectorW{b0 * r0, b1 * r1, b2 * r2, 0}
			b.W = 1 / (b.X + b.Y + b.Z)
			v := InterpolateVertexes(v0, v1, v2, b)
			// invoke fragment shader
			color := dc.Shader.Fragment(v)
			if color == dc.Discard {
				continue
			}
			// update buffers atomically
			lock := &dc.locks[(x+y)&255]
			lock.Lock()
			// check depth buffer again
			if bz <= dc.DepthBuffer[i] || !dc.ReadDepth {
				info.UpdatedPixels++
				if dc.WriteDepth {
					// update depth buffer
					dc.DepthBuffer[i] = z
				}
				if dc.WriteColor {
					// update color buffer
					dc.ColorBuffer.Write(x, y, color)
				}
			}
			lock.Unlock()
		}
		w00 += b12
		w01 += b20
		w02 += b01
	}

	return info
}

func (dc *Context[B, T]) line(v0, v1 Vertex, s0, s1 Vector) RasterizeInfo {
	n := s1.Sub(s0).Perpendicular().MulScalar(dc.LineWidth / 2)
	s0 = s0.Add(s0.Sub(s1).Normalize().MulScalar(dc.LineWidth / 2))
	s1 = s1.Add(s1.Sub(s0).Normalize().MulScalar(dc.LineWidth / 2))
	s00 := s0.Add(n)
	s01 := s0.Sub(n)
	s10 := s1.Add(n)
	s11 := s1.Sub(n)
	info1 := dc.rasterize(v1, v0, v0, s11, s01, s00)
	info2 := dc.rasterize(v1, v1, v0, s10, s11, s00)
	return info1.Add(info2)
}

func (dc *Context[B, T]) wireframe(v0, v1, v2 Vertex, s0, s1, s2 Vector) RasterizeInfo {
	info1 := dc.line(v0, v1, s0, s1)
	info2 := dc.line(v1, v2, s1, s2)
	info3 := dc.line(v2, v0, s2, s0)
	return info1.Add(info2).Add(info3)
}

func (dc *Context[B, T]) drawClippedLine(v0, v1 Vertex) RasterizeInfo {
	// normalized device coordinates
	ndc0 := v0.Output.DivScalar(v0.Output.W).Vector()
	ndc1 := v1.Output.DivScalar(v1.Output.W).Vector()

	// screen coordinates
	s0 := dc.screenMatrix.MulPosition(ndc0)
	s1 := dc.screenMatrix.MulPosition(ndc1)

	// rasterize
	return dc.line(v0, v1, s0, s1)
}

func (dc *Context[B, T]) drawClippedTriangle(v0, v1, v2 Vertex) RasterizeInfo {
	// normalized device coordinates
	ndc0 := v0.Output.DivScalar(v0.Output.W).Vector()
	ndc1 := v1.Output.DivScalar(v1.Output.W).Vector()
	ndc2 := v2.Output.DivScalar(v2.Output.W).Vector()

	// back face culling
	a := (ndc1.X-ndc0.X)*(ndc2.Y-ndc0.Y) - (ndc2.X-ndc0.X)*(ndc1.Y-ndc0.Y)
	if a < 0 {
		v0, v1, v2 = v2, v1, v0
		ndc0, ndc1, ndc2 = ndc2, ndc1, ndc0
	}
	if dc.Cull == CullFront {
		a = -a
	}
	if dc.FrontFace == FaceCW {
		a = -a
	}
	if dc.Cull != CullNone && a <= 0 {
		return RasterizeInfo{}
	}

	// screen coordinates
	s0 := dc.screenMatrix.MulPosition(ndc0)
	s1 := dc.screenMatrix.MulPosition(ndc1)
	s2 := dc.screenMatrix.MulPosition(ndc2)

	// rasterize
	if dc.Wireframe {
		return dc.wireframe(v0, v1, v2, s0, s1, s2)
	} else {
		return dc.rasterize(v0, v1, v2, s0, s1, s2)
	}
}

func (dc *Context[B, T]) DrawLine(t *Line) RasterizeInfo {
	// invoke vertex shader
	v1 := dc.Shader.Vertex(t.V1)
	v2 := dc.Shader.Vertex(t.V2)

	if v1.Outside() || v2.Outside() {
		// clip to viewing volume
		line := ClipLine(NewLine(v1, v2))
		if line != nil {
			return dc.drawClippedLine(line.V1, line.V2)
		} else {
			return RasterizeInfo{}
		}
	} else {
		// no need to clip
		return dc.drawClippedLine(v1, v2)
	}
}

func (dc *Context[B, T]) DrawTriangle(t *Triangle) RasterizeInfo {
	// invoke vertex shader
	v1 := dc.Shader.Vertex(t.V1)
	v2 := dc.Shader.Vertex(t.V2)
	v3 := dc.Shader.Vertex(t.V3)

	if v1.Outside() || v2.Outside() || v3.Outside() {
		// clip to viewing volume
		triangles := ClipTriangle(NewTriangle(v1, v2, v3))
		var result RasterizeInfo
		for _, t := range triangles {
			info := dc.drawClippedTriangle(t.V1, t.V2, t.V3)
			result = result.Add(info)
		}
		return result
	} else {
		// no need to clip
		return dc.drawClippedTriangle(v1, v2, v3)
	}
}

func (dc *Context[B, T]) DrawLines(lines []*Line) RasterizeInfo {
	wn := runtime.NumCPU()
	ch := make(chan RasterizeInfo, wn)
	for wi := 0; wi < wn; wi++ {
		go func(wi int) {
			var result RasterizeInfo
			for i, l := range lines {
				if i%wn == wi {
					info := dc.DrawLine(l)
					result = result.Add(info)
				}
			}
			ch <- result
		}(wi)
	}
	var result RasterizeInfo
	for wi := 0; wi < wn; wi++ {
		result = result.Add(<-ch)
	}
	return result
}

func (dc *Context[B, T]) DrawTriangles(triangles []*Triangle) RasterizeInfo {
	wn := runtime.NumCPU()
	ch := make(chan RasterizeInfo, wn)
	for wi := 0; wi < wn; wi++ {
		go func(wi int) {
			var result RasterizeInfo
			for i, t := range triangles {
				if i%wn == wi {
					info := dc.DrawTriangle(t)
					result = result.Add(info)
				}
			}
			ch <- result
		}(wi)
	}
	var result RasterizeInfo
	for wi := 0; wi < wn; wi++ {
		result = result.Add(<-ch)
	}
	return result
}

func (dc *Context[B, T]) DrawMesh(mesh *Mesh) RasterizeInfo {
	info1 := dc.DrawTriangles(mesh.Triangles)
	info2 := dc.DrawLines(mesh.Lines)
	return info1.Add(info2)
}
