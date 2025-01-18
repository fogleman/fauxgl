package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"

	"github.com/fogleman/colormap"
	. "github.com/fogleman/fauxgl"
	"github.com/nfnt/resize"
)

const (
	scale  = 1    // optional supersampling
	width  = 1600 // output width in pixels
	height = 1600 // output height in pixels
)

var (
	eye    = V(0, 0, 0)
	center = V(0, 0, 1)
	up     = V(0, -1, 0)
)

var palette = colormap.New(colormap.ParseColors("67001fb2182bd6604df4a582fddbc7f7f7f7d1e5f092c5de4393c32166ac053061"))

type ColorShader struct {
	Matrix Matrix
}

func NewColorShader(matrix Matrix) *ColorShader {
	return &ColorShader{matrix}
}

func (shader *ColorShader) Vertex(v Vertex) Vertex {
	v.Output = shader.Matrix.MulPositionW(v.Position)
	return v
}

func (shader *ColorShader) Fragment(v Vertex) Color {
	return v.Color
}

func LoadVoxels(path string) ([]Voxel, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	var voxels []Voxel
	for {
		row, err := reader.Read()
		if row == nil {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		x, _ := strconv.ParseInt(row[0], 0, 0)
		y, _ := strconv.ParseInt(row[1], 0, 0)
		z, _ := strconv.ParseInt(row[2], 0, 0)
		t, _ := strconv.ParseFloat(row[3], 64)
		c := Color{t, t, t, 1}

		v := Voxel{int(x), int(y), int(z), c}
		voxels = append(voxels, v)
	}
	return voxels, nil
}

func FilterVoxels(gridVoxels, curvatureVoxels []Voxel, voxelSize float64) []Voxel {
	var voxels []Voxel
	lookup := make(map[Voxel]float64)
	for _, voxel := range gridVoxels {
		key := voxel
		key.Color = Color{}
		lookup[key] = voxel.Color.R
	}
	for _, voxel := range curvatureVoxels {
		key := voxel
		key.Color = Color{}
		dist, ok := lookup[key]
		if !ok {
			continue
		}
		if math.Abs(dist) > voxelSize/2*math.Sqrt2 {
			continue
		}
		r, g, b, _ := palette.At((-voxel.Color.R + 1) / 2).RGBA()
		c := Color{float64(r) / 0xffff, float64(g) / 0xffff, float64(b) / 0xffff, 1}
		voxel.Color = c
		voxels = append(voxels, voxel)
	}
	return voxels
}

func main() {
	args := os.Args[1:]
	if len(args) != 3 {
		log.Fatal("Usage: vdbcurvature grid.csv curvature.csv voxel_size")
		return
	}

	voxelSize, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		log.Fatal(err)
	}

	gridVoxels, err := LoadVoxels(args[0])
	if err != nil {
		log.Fatal(err)
	}

	curvatureVoxels, err := LoadVoxels(args[1])
	if err != nil {
		log.Fatal(err)
	}

	voxels := FilterVoxels(gridVoxels, curvatureVoxels, voxelSize)

	mesh := NewVoxelMesh(voxels)
	mesh.Transform(Rotate(Vector{1, 0, 0}, -math.Pi/2))

	fmt.Println(len(voxels), "voxels")
	fmt.Println(len(mesh.Triangles), "triangles")

	mesh.BiUnitCube()

	context := NewContext(width*scale, height*scale)
	context.ClearColorBufferWith(HexColor("323"))

	const s = 1.5
	matrix := LookAt(eye, center, up).Orthographic(-s, s, -s, s, -20, 20)

	// render
	context.Shader = NewColorShader(matrix)
	context.DrawTriangles(mesh.Triangles)

	// downsample image for antialiasing
	image := context.Image()
	image = resize.Resize(width, height, image, resize.Bilinear)

	// save image
	SavePNG("out.png", image)
}
