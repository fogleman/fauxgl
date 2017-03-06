package fauxgl

import (
	"encoding/binary"
	"io"
	"os"
)

func Load3DS(filename string) (*Mesh, error) {
	type ChunkHeader struct {
		ChunkID uint16
		Length  uint32
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var vertices []Vector
	var triangles []*Triangle
	for {
		header := ChunkHeader{}
		if err := binary.Read(file, binary.LittleEndian, &header); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		switch header.ChunkID {
		case 0x4D4D:
		case 0x3D3D:
		case 0x4000:
			_, err := readNullTerminatedString(file)
			if err != nil {
				return nil, err
			}
		case 0x4100:
		case 0x4110:
			v, err := readVertexList(file)
			if err != nil {
				return nil, err
			}
			vertices = v
		case 0x4120:
			t, err := readFaceList(file, vertices)
			if err != nil {
				return nil, err
			}
			triangles = append(triangles, t...)
		default:
			file.Seek(int64(header.Length-6), 1)
		}
	}

	return NewTriangleMesh(triangles), nil
}

func readVertexList(file *os.File) ([]Vector, error) {
	var count uint16
	if err := binary.Read(file, binary.LittleEndian, &count); err != nil {
		return nil, err
	}
	result := make([]Vector, count)
	for i := range result {
		var v [3]float32
		if err := binary.Read(file, binary.LittleEndian, &v); err != nil {
			return nil, err
		}
		result[i] = Vector{float64(v[0]), float64(v[1]), float64(v[2])}
	}
	return result, nil
}

func readFaceList(file *os.File, vertices []Vector) ([]*Triangle, error) {
	var count uint16
	if err := binary.Read(file, binary.LittleEndian, &count); err != nil {
		return nil, err
	}
	result := make([]*Triangle, count)
	for i := range result {
		var v [4]uint16
		if err := binary.Read(file, binary.LittleEndian, &v); err != nil {
			return nil, err
		}
		result[i] = NewTriangleForPoints(
			vertices[v[0]], vertices[v[1]], vertices[v[2]])
	}
	return result, nil
}

func readNullTerminatedString(file *os.File) (string, error) {
	var bytes []byte
	buf := make([]byte, 1)
	for {
		n, err := file.Read(buf)
		if err != nil {
			return "", err
		} else if n == 1 {
			if buf[0] == 0 {
				break
			}
			bytes = append(bytes, buf[0])
		}
	}
	return string(bytes), nil
}
