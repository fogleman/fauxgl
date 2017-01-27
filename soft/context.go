package soft

import (
	"image"
	"image/draw"
	"math"
)

type Context struct {
	Width        int
	Height       int
	ColorBuffer  *image.NRGBA
	DepthBuffer  []float64
	screenMatrix Matrix
}

func NewContext(width, height int) *Context {
	dc := &Context{}
	dc.Width = width
	dc.Height = height
	dc.ColorBuffer = image.NewNRGBA(image.Rect(0, 0, width, height))
	dc.DepthBuffer = make([]float64, width*height)
	dc.screenMatrix = Scale(V(1, -1, 1)).Translate(V(1, 1, 0)).Scale(V(float64(width)/2, float64(height)/2, 1))
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

func (dc *Context) drawClipped(v1, v2, v3 Vertex, shader Shader, buf []Fragment) []Fragment {
	ndc1 := v1.Output.DivScalar(v1.Output.W).Vector()
	ndc2 := v2.Output.DivScalar(v2.Output.W).Vector()
	ndc3 := v3.Output.DivScalar(v3.Output.W).Vector()
	s1 := dc.screenMatrix.MulPosition(ndc1)
	s2 := dc.screenMatrix.MulPosition(ndc2)
	s3 := dc.screenMatrix.MulPosition(ndc3)
	buf = Rasterize(s1, s2, s3, buf)
	for _, f := range buf {
		i := f.Y*dc.Width + f.X
		if f.Depth > dc.DepthBuffer[i] {
			continue
		}
		v := InterpolateVertexes(v1, v2, v3, f.Barycentric)
		color := shader.Fragment(v)
		if color == Discard {
			continue
		}
		dc.DepthBuffer[i] = f.Depth
		dc.ColorBuffer.SetNRGBA(f.X, f.Y, color.NRGBA())
	}
	return buf
}

func (dc *Context) drawTriangle(t *Triangle, shader Shader, buf []Fragment) []Fragment {
	v1 := shader.Vertex(t.V1)
	v2 := shader.Vertex(t.V2)
	v3 := shader.Vertex(t.V3)
	if v1.Outside() || v2.Outside() || v3.Outside() {
		triangles := ClipTriangle(NewTriangle(v1, v2, v3))
		for _, t := range triangles {
			buf = dc.drawClipped(t.V1, t.V2, t.V3, shader, buf)
		}
	} else {
		buf = dc.drawClipped(v1, v2, v3, shader, buf)
	}
	return buf
}

func (dc *Context) DrawTriangles(triangles []*Triangle, shader Shader) {
	var buf []Fragment
	for _, t := range triangles {
		buf = dc.drawTriangle(t, shader, buf)
	}
}

func (dc *Context) DrawMesh(mesh *Mesh, shader Shader) {
	dc.DrawTriangles(mesh.Triangles, shader)
}
