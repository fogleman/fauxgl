package main

import (
	"fmt"
	"image"
	"image/draw"
	"os"
	"path/filepath"

	. "github.com/fogleman/fauxgl"
	"github.com/nfnt/resize"
)

const (
	scale  = 2
	width  = 800
	height = 600
	fovy   = 40
	near   = 1
	far    = 10
)

var (
	eye      = V(2, 2, 2)
	center   = V(0, 0, -0.25)
	up       = V(0, 0, 1)
	lightDir = V(0.25, -0.25, 3)
)

func ensureGray16(im image.Image) *image.Gray16 {
	switch im := im.(type) {
	case *image.Gray16:
		return im
	default:
		dst := image.NewGray16(im.Bounds())
		draw.Draw(dst, im.Bounds(), im, image.ZP, draw.Src)
		return dst
	}
}

func heightmapMesh(im image.Image, xScale, yScale, zScale float64) *Mesh {
	gray := ensureGray16(im)
	w := gray.Bounds().Size().X
	h := gray.Bounds().Size().Y
	triangles := make([]*Triangle, 0, (w-1)*(h-1)*4)
	zScale /= 65536
	for j0 := 0; j0 < h-1; j0++ {
		j1 := j0 + 1
		y0 := float64(j0) * yScale
		y1 := float64(j1) * yScale
		for i0 := 0; i0 < w-1; i0++ {
			i1 := i0 + 1
			x0 := float64(i0) * xScale
			x1 := float64(i1) * xScale
			z00 := float64(gray.Gray16At(i0, j0).Y) * zScale
			z01 := float64(gray.Gray16At(i0, j1).Y) * zScale
			z10 := float64(gray.Gray16At(i1, j0).Y) * zScale
			z11 := float64(gray.Gray16At(i1, j1).Y) * zScale
			xm := (x0 + x1) / 2
			ym := (y0 + y1) / 2
			zm := (z00 + z01 + z10 + z11) / 4
			p00 := Vector{x0, y0, z00}
			p01 := Vector{x0, y1, z01}
			p10 := Vector{x1, y0, z10}
			p11 := Vector{x1, y1, z11}
			pm := Vector{xm, ym, zm}
			triangles = append(triangles, NewTriangleForPoints(p10, pm, p00))
			triangles = append(triangles, NewTriangleForPoints(p00, pm, p01))
			triangles = append(triangles, NewTriangleForPoints(p11, pm, p10))
			triangles = append(triangles, NewTriangleForPoints(p01, pm, p11))
		}
	}
	return NewTriangleMesh(triangles)
}

func run(inputPath, outputPath string, xScale, yScale, zScale float64) error {
	// load heightmap image
	im, err := LoadImage(inputPath)
	if err != nil {
		return err
	}

	// convert image to a mesh
	mesh := heightmapMesh(im, xScale, yScale, zScale)

	// fit mesh in a bi-unit cube centered at the origin (ignoring Z)
	mesh.FitInside(Box{Vector{-1, -1, 0}, Vector{1, 1, 1e9}}, Vector{0.5, 0.5, 0})

	// smooth the normals
	mesh.SmoothNormals()

	// create a rendering context
	context := NewContext(width*scale, height*scale)
	context.ClearColorBufferWith(Black)

	// create transformation matrix and light direction
	aspect := float64(width) / float64(height)
	matrix := LookAt(eye, center, up).Perspective(fovy, aspect, near, far)
	light := lightDir.Normalize()

	// render
	shader := NewPhongShader(matrix, light, eye)
	shader.ObjectColor = HexColor("5A8CA2")
	shader.DiffuseColor = Gray(1)
	shader.SpecularColor = Gray(0)
	shader.SpecularPower = 100
	context.Shader = shader
	context.DrawMesh(mesh)

	// save image
	image := context.Image()
	image = resize.Resize(width, height, image, resize.Bilinear)
	return SavePNG(outputPath, image)
}

func main() {
	for _, path := range os.Args[1:] {
		name := filepath.Base(path)
		fmt.Println(name)
		outputPath := fmt.Sprintf("out/%s.png", name)
		run(path, outputPath, 1, 1, 64)
	}
}
