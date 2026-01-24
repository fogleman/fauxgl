package fauxgl

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func Radians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func Degrees(radians float64) float64 {
	return radians * 180 / math.Pi
}

func LatLngToXYZ(lat, lng float64) Vector {
	lat, lng = Radians(lat), Radians(lng)
	x := math.Cos(lat) * math.Cos(lng)
	y := math.Cos(lat) * math.Sin(lng)
	z := math.Sin(lat)
	return Vector{x, y, z}
}

type SeekReadCloser interface {
	io.ReaderFrom
}

func LoadMesh(path string) (*Mesh, error) {
	meshType, err := MeshTypeFromPath(path)
	if err != nil {
		return nil, err
	}

	fi, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	stat, err := fi.Stat()
	if err != nil {
		return nil, err
	}

	return LoadMeshSizedReader(meshType, fi, stat.Size())
}

type MeshType string

const (
	MeshTypeUnknown = ""
	MeshTypeSTL = ".stl"
	MeshTypeOBJ = ".obj"
	MeshTypePLY = ".ply"
	MeshType3DS = ".3ds"
)

func MeshTypeFromPath(path string) (MeshType, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case MeshTypeSTL:
		return MeshTypeSTL, nil
	case MeshTypeOBJ:
		return MeshTypeOBJ, nil
	case MeshTypePLY:
		return MeshTypePLY, nil
	case MeshType3DS:
		return MeshType3DS, nil
	}

	return MeshTypeUnknown, fmt.Errorf("unrecognized mesh type: %s", ext)
}

func LoadMeshSizedReader(
	meshType MeshType,
	r io.Reader,
	size int64,
) (*Mesh, error) {
	switch meshType {
	case MeshTypeSTL:
		return LoadSTLReader(r, size)
	case MeshTypeOBJ:
		return LoadOBJReader(r)
	case MeshTypePLY:
		return LoadPLYReader(r)
	case MeshType3DS:
		return Load3DSReader(r)
	}

	return nil, fmt.Errorf("unsupported mesh type: %s", meshType)
}

func LoadMeshReader(
	meshType MeshType,
	r io.Reader,
) (*Mesh, error) {
	switch meshType {
	case MeshTypeSTL:
		return nil, fmt.Errorf("cannot load an stl mesh from an unsized io.Reader")
	case MeshTypeOBJ:
		return LoadOBJReader(r)
	case MeshTypePLY:
		return LoadPLYReader(r)
	case MeshType3DS:
		return Load3DSReader(r)
	}

	return nil, fmt.Errorf("unsupported mesh type: %s", meshType)
}

func LoadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	im, _, err := image.Decode(file)
	return im, err
}

func SavePNG(path string, im image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, im)
}

func ParseFloats(items []string) []float64 {
	result := make([]float64, len(items))
	for i, item := range items {
		f, _ := strconv.ParseFloat(item, 64)
		result[i] = f
	}
	return result
}

func Clamp(x, lo, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

func ClampInt(x, lo, hi int) int {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

func AbsInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func Round(a float64) int {
	if a < 0 {
		return int(math.Ceil(a - 0.5))
	} else {
		return int(math.Floor(a + 0.5))
	}
}

func RoundPlaces(a float64, places int) float64 {
	shift := powersOfTen[places]
	return float64(Round(a*shift)) / shift
}

var powersOfTen = []float64{1e0, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9, 1e10, 1e11, 1e12, 1e13, 1e14, 1e15, 1e16}
