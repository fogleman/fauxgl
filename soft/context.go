package soft

import (
	"image"
	"image/draw"
	"math"
)

var clipBox = Box{V(-1, -1, -1), V(1, 1, 1)}

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

func (dc *Context) drawTriangle(t *Triangle, shader Shader, buf []Fragment) []Fragment {
	outside := 0
	w1 := shader.Vertex(t.V1)
	ndc1 := w1.Output.DivScalar(w1.Output.W)
	if !clipBox.Contains(ndc1) {
		outside++
	}
	w2 := shader.Vertex(t.V2)
	ndc2 := w2.Output.DivScalar(w2.Output.W)
	if !clipBox.Contains(ndc2) {
		outside++
	}
	w3 := shader.Vertex(t.V3)
	ndc3 := w3.Output.DivScalar(w3.Output.W)
	if !clipBox.Contains(ndc3) {
		outside++
	}
	if outside == 3 {
		return buf
	}
	s1 := dc.screenMatrix.MulPosition(ndc1)
	s2 := dc.screenMatrix.MulPosition(ndc2)
	s3 := dc.screenMatrix.MulPosition(ndc3)
	buf = Rasterize(dc.Width, dc.Height, s1, s2, s3, buf)
	for _, f := range buf {
		if outside > 0 {
			if f.X < 0 || f.X >= dc.Width || f.Y < 0 || f.Y >= dc.Height {
				continue
			}
		}
		i := f.Y*dc.Width + f.X
		if f.Depth > dc.DepthBuffer[i] {
			continue
		}
		v := InterpolateVertexes(t.V1, t.V2, t.V3, f.Barycentric)
		color := shader.Fragment(v)
		if color == Discard {
			continue
		}
		dc.DepthBuffer[i] = f.Depth
		dc.ColorBuffer.SetNRGBA(f.X, f.Y, color.NRGBA())
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
