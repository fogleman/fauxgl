package fauxgl

import (
	"bufio"
	"encoding/binary"
	"os"
	"strings"
)

type STLHeader struct {
	_     [80]uint8
	Count uint32
}

type STLTriangle struct {
	N, V1, V2, V3 [3]float32
	_             uint16
}

func LoadSTL(path string) (*Mesh, error) {
	// open file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// get file size
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	size := info.Size()

	// read header, get expected binary size
	header := STLHeader{}
	if err := binary.Read(file, binary.LittleEndian, &header); err != nil {
		return nil, err
	}
	expectedSize := int64(header.Count)*50 + 84

	// rewind to start of file
	_, err = file.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	// parse ascii or binary stl
	if size == expectedSize {
		return loadSTLB(file)
	} else {
		return loadSTLA(file)
	}
}

func loadSTLA(file *os.File) (*Mesh, error) {
	var vertexes []Vector
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) == 4 && fields[0] == "vertex" {
			f := ParseFloats(fields[1:])
			vertexes = append(vertexes, Vector{f[0], f[1], f[2]})
		}
	}
	var triangles []*Triangle
	for i := 0; i < len(vertexes); i += 3 {
		t := Triangle{}
		t.V1.Position = vertexes[i+0]
		t.V2.Position = vertexes[i+1]
		t.V3.Position = vertexes[i+2]
		t.FixNormals()
		triangles = append(triangles, &t)
	}
	return NewTriangleMesh(triangles), scanner.Err()
}

func loadSTLB(file *os.File) (*Mesh, error) {
	header := STLHeader{}
	if err := binary.Read(file, binary.LittleEndian, &header); err != nil {
		return nil, err
	}
	count := int(header.Count)
	triangles := make([]*Triangle, count)
	for i := 0; i < count; i++ {
		d := STLTriangle{}
		if err := binary.Read(file, binary.LittleEndian, &d); err != nil {
			return nil, err
		}
		t := Triangle{}
		t.V1.Position = Vector{float64(d.V1[0]), float64(d.V1[1]), float64(d.V1[2])}
		t.V2.Position = Vector{float64(d.V2[0]), float64(d.V2[1]), float64(d.V2[2])}
		t.V3.Position = Vector{float64(d.V3[0]), float64(d.V3[1]), float64(d.V3[2])}
		t.FixNormals()
		triangles[i] = &t
	}
	return NewTriangleMesh(triangles), nil
}

func SaveSTL(path string, mesh *Mesh) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	header := STLHeader{}
	header.Count = uint32(len(mesh.Triangles))
	if err := binary.Write(file, binary.LittleEndian, &header); err != nil {
		return err
	}
	for _, triangle := range mesh.Triangles {
		n := triangle.Normal()
		d := STLTriangle{}
		d.N[0] = float32(n.X)
		d.N[1] = float32(n.Y)
		d.N[2] = float32(n.Z)
		d.V1[0] = float32(triangle.V1.Position.X)
		d.V1[1] = float32(triangle.V1.Position.Y)
		d.V1[2] = float32(triangle.V1.Position.Z)
		d.V2[0] = float32(triangle.V2.Position.X)
		d.V2[1] = float32(triangle.V2.Position.Y)
		d.V2[2] = float32(triangle.V2.Position.Z)
		d.V3[0] = float32(triangle.V3.Position.X)
		d.V3[1] = float32(triangle.V3.Position.Y)
		d.V3[2] = float32(triangle.V3.Position.Z)
		if err := binary.Write(file, binary.LittleEndian, &d); err != nil {
			return err
		}
	}
	return nil
}
