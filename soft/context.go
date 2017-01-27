package soft

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
	screenMatrix Matrix
	locks        []sync.Mutex
}

func NewContext(width, height int) *Context {
	dc := &Context{}
	dc.Width = width
	dc.Height = height
	dc.ColorBuffer = image.NewNRGBA(image.Rect(0, 0, width, height))
	dc.DepthBuffer = make([]float64, width*height)
	dc.screenMatrix = Scale(V(1, -1, 1)).Translate(V(1, 1, 0)).Scale(V(float64(width)/2, float64(height)/2, 1))
	dc.locks = make([]sync.Mutex, 128)
	dc.ClearDepthBuffer()
	return dc
}

func (dc *Context) Image() image.Image {
	return dc.ColorBuffer
}

func (dc *Context) ClearColorBuffer(color Vector) {
	im := dc.ColorBuffer
	src := image.NewUniform(color.NRGBA())
	draw.Draw(im, im.Bounds(), src, image.ZP, draw.Src)
}

func (dc *Context) ClearDepthBuffer() {
	for i := range dc.DepthBuffer {
		dc.DepthBuffer[i] = math.MaxFloat64
	}
}

func (dc *Context) rasterize(v1, v2, v3 Vertex, s1, s2, s3 Vector, shader Shader) {
	min := s1.Min(s2.Min(s3)).Floor()
	max := s1.Max(s2.Max(s3)).Ceil()
	x1 := int(min.X)
	x2 := int(max.X)
	y1 := int(min.Y)
	y2 := int(max.Y)
	d0 := s2.Sub(s1)
	d1 := s3.Sub(s1)
	d00 := d0.X*d0.X + d0.Y*d0.Y
	d01 := d0.X*d1.X + d0.Y*d1.Y
	d11 := d1.X*d1.X + d1.Y*d1.Y
	lock := &dc.locks[(x1+y1)%len(dc.locks)]
	for y := y1; y <= y2; y++ {
		for x := x1; x <= x2; x++ {
			p := Vector{float64(x) + 0.5, float64(y) + 0.5, 0}
			d2 := p.Sub(s1)
			d20 := d2.X*d0.X + d2.Y*d0.Y
			d21 := d2.X*d1.X + d2.Y*d1.Y
			d := d00*d11 - d01*d01
			if d > -0.001 && d < 0.001 {
				continue
			}
			by := (d11*d20 - d01*d21) / d
			if by < 0 {
				continue
			}
			bz := (d00*d21 - d01*d20) / d
			if bz < 0 {
				continue
			}
			bx := 1 - by - bz
			if bx < 0 {
				continue
			}
			z := bx*s1.Z + by*s2.Z + bz*s3.Z
			i := y*dc.Width + x
			if z >= dc.DepthBuffer[i] { // completely safe?
				continue
			}
			b := Vector{bx, by, bz}
			v := InterpolateVertexes(v1, v2, v3, b)
			color := shader.Fragment(v)
			if color == Discard {
				continue
			}
			c := color.NRGBA()
			lock.Lock()
			if z < dc.DepthBuffer[i] {
				dc.DepthBuffer[i] = z
				dc.ColorBuffer.SetNRGBA(x, y, c)
			}
			lock.Unlock()
		}
	}
}

func (dc *Context) drawClipped(v1, v2, v3 Vertex, shader Shader) {
	ndc1 := v1.Output.DivScalar(v1.Output.W).Vector()
	ndc2 := v2.Output.DivScalar(v2.Output.W).Vector()
	ndc3 := v3.Output.DivScalar(v3.Output.W).Vector()

	// back face culling
	var sum float64
	sum += (ndc2.X - ndc1.X) * (ndc2.Y + ndc1.Y)
	sum += (ndc3.X - ndc2.X) * (ndc3.Y + ndc2.Y)
	sum += (ndc1.X - ndc3.X) * (ndc1.Y + ndc3.Y)
	if sum >= 0 {
		return
	}

	s1 := dc.screenMatrix.MulPosition(ndc1)
	s2 := dc.screenMatrix.MulPosition(ndc2)
	s3 := dc.screenMatrix.MulPosition(ndc3)
	dc.rasterize(v1, v2, v3, s1, s2, s3, shader)
}

func (dc *Context) DrawTriangle(t *Triangle, shader Shader) {
	v1 := shader.Vertex(t.V1)
	v2 := shader.Vertex(t.V2)
	v3 := shader.Vertex(t.V3)
	if v1.Outside() || v2.Outside() || v3.Outside() {
		triangles := ClipTriangle(NewTriangle(v1, v2, v3))
		for _, t := range triangles {
			dc.drawClipped(t.V1, t.V2, t.V3, shader)
		}
	} else {
		dc.drawClipped(v1, v2, v3, shader)
	}
}

func (dc *Context) DrawTriangles(triangles []*Triangle, shader Shader) {
	wn := runtime.NumCPU()
	done := make(chan bool, wn)
	for wi := 0; wi < wn; wi++ {
		go func(wi int) {
			for i, t := range triangles {
				if i%wn == wi {
					dc.DrawTriangle(t, shader)
				}
			}
			done <- true
		}(wi)
	}
	for wi := 0; wi < wn; wi++ {
		<-done
	}
}

func (dc *Context) DrawMesh(mesh *Mesh, shader Shader) {
	dc.DrawTriangles(mesh.Triangles, shader)
}
