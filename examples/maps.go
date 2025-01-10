package main

import (
	"fmt"
	"image"
	"image/draw"
	"math"
	"os"

	. "github.com/fogleman/fauxgl"
	"github.com/nfnt/resize"
)

const (
	scale  = 1
	width  = 1600
	height = 1200
)

const (
	modeHeight = iota
	modeAngle
	modeNormal
)

var (
	eye    = V(0, 0, -1)
	center = V(0, 0, 0)
	up     = V(0, -1, 0)
)

func ensureGray(im image.Image) *image.Gray {
	switch im := im.(type) {
	case *image.Gray:
		return im
	default:
		dst := image.NewGray(im.Bounds())
		draw.Draw(dst, im.Bounds(), im, image.ZP, draw.Src)
		return dst
	}
}

type MapShader struct {
	Mode      int
	Matrix    Matrix
	Z0, Z1    float64
	HeightMap *image.Gray
}

func NewMapShader(mode int, matrix Matrix, z0, z1 float64, heightmap *image.Gray) *MapShader {
	return &MapShader{mode, matrix, z0, z1, heightmap}
}

func (shader *MapShader) Vertex(v Vertex) Vertex {
	v.Output = shader.Matrix.MulPositionW(v.Position)
	return v
}

func (shader *MapShader) Fragment(v Vertex) Color {
	z := v.Position.Z
	t := (z - shader.Z0) / (shader.Z1 - shader.Z0)
	t = Clamp(t, 0, 1)

	if shader.HeightMap != nil {
		p := shader.HeightMap.GrayAt(v.X, v.Y).Y
		u := float64(p+1) / 255
		if u >= t {
			return Discard
		}
	}

	if shader.Mode == modeNormal {
		n := v.Normal
		n.Z = -n.Z
		n = n.AddScalar(1).DivScalar(2)
		return Color{n.X, n.Y, n.Z, 1}
	}

	if shader.Mode == modeAngle {
		a := 1 - math.Abs(v.Normal.Dot(V(0, 0, 1)))
		return Color{a, a, a, 1}
	}

	return Color{t, t, t, 1}
}

func updateHeightMap(heightMap, update *image.Gray) *image.Gray {
	if heightMap == nil {
		return update
	}
	for i := range heightMap.Pix {
		if update.Pix[i] > heightMap.Pix[i] {
			heightMap.Pix[i] = update.Pix[i]
		}
	}
	return heightMap
}

func run(inputPath string) error {
	mesh, err := LoadMesh(inputPath)
	if err != nil {
		return err
	}

	mesh.BiUnitCube()
	mesh.SmoothNormalsThreshold(Radians(30))
	box := mesh.BoundingBox()
	z0 := box.Min.Z
	z1 := box.Max.Z

	const s = 1
	aspect := float64(width) / height
	matrix := LookAt(eye, center, up).Orthographic(-s*aspect, s*aspect, -s, s, z0-1, z1+1)

	context := NewContext(width*scale, height*scale)

	var heightMap *image.Gray

	for i := 0; i < 5; i++ {
		context.ClearDepthBuffer()
		context.ClearColorBufferWith(Color{})
		context.Shader = NewMapShader(modeNormal, matrix, z0, z1, heightMap)
		context.DrawMesh(mesh)
		path := fmt.Sprintf("normal-%02d.png", i)
		SavePNG(path, resize.Resize(width, height, context.Image(), resize.Bilinear))

		context.ClearDepthBuffer()
		context.ClearColorBufferWith(Color{})
		context.Shader = NewMapShader(modeAngle, matrix, z0, z1, heightMap)
		context.DrawMesh(mesh)
		path = fmt.Sprintf("angle-%02d.png", i)
		SavePNG(path, resize.Resize(width, height, context.Image(), resize.Bilinear))

		context.ClearDepthBuffer()
		context.ClearColorBufferWith(Color{})
		context.Shader = NewMapShader(modeHeight, matrix, z0, z1, heightMap)
		info := context.DrawMesh(mesh)
		fmt.Println(info.UpdatedPixels)
		path = fmt.Sprintf("height-%02d.png", i)
		SavePNG(path, resize.Resize(width, height, context.Image(), resize.Bilinear))

		heightMap = updateHeightMap(heightMap, ensureGray(context.Image()))
	}

	return nil
}

func main() {
	inputPath := os.Args[1]
	run(inputPath)
}
