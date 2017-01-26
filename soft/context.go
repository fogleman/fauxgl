package soft

import (
	"image"
	"image/draw"
	"math"
)

var clipBox = Box{V(-1, -1, -1), V(1, 1, 1)}

type Context struct {
	Width          int
	Height         int
	ColorBuffer    *image.NRGBA
	DepthBuffer    []float64
	screenMatrix   Matrix
	fragmentBuffer []Fragment
}

func NewContext(width, height int) *Context {
	dc := &Context{}
	dc.Width = width
	dc.Height = height
	dc.ColorBuffer = image.NewNRGBA(image.Rect(0, 0, width, height))
	dc.DepthBuffer = make([]float64, width*height)
	dc.screenMatrix = Scale(V(1, -1, 1)).Translate(V(1, 1, 0)).Scale(V(float64(width)/2, float64(height)/2, 1))
	dc.fragmentBuffer = nil
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

func (dc *Context) DrawTriangle(t *Triangle, shader Shader) {
	w1 := shader.Vertex(t.V1)
	if !clipBox.Contains(w1.Position) {
		return
	}
	w2 := shader.Vertex(t.V2)
	if !clipBox.Contains(w2.Position) {
		return
	}
	w3 := shader.Vertex(t.V3)
	if !clipBox.Contains(w3.Position) {
		return
	}
	s1 := dc.screenMatrix.MulPosition(w1.Position)
	s2 := dc.screenMatrix.MulPosition(w2.Position)
	s3 := dc.screenMatrix.MulPosition(w3.Position)
	dc.fragmentBuffer = Rasterize(s1, s2, s3, dc.fragmentBuffer)
	for _, f := range dc.fragmentBuffer {
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
}

func (dc *Context) DrawMesh(mesh *Mesh, shader Shader) {
	for _, t := range mesh.Triangles {
		dc.DrawTriangle(t, shader)
	}
}
