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
	var faces []*Triangle
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
			f, err := readFaceList(file, vertices)
			if err != nil {
				return nil, err
			}
			faces = f
			triangles = append(triangles, faces...)
		case 0x4150:
			err := readSmoothingGroups(file, faces)
			if err != nil {
				return nil, err
			}
		// case 0x4160:
		// 	matrix, err := readLocalAxis(file)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	for i, v := range vertices {
		// 		vertices[i] = matrix.MulPosition(v)
		// 	}
		default:
			file.Seek(int64(header.Length-6), 1)
		}
	}

	return NewTriangleMesh(triangles), nil
}

func readSmoothingGroups(file *os.File, triangles []*Triangle) error {
	groups := make([]uint32, len(triangles))
	if err := binary.Read(file, binary.LittleEndian, &groups); err != nil {
		return err
	}
	var tables [32]map[Vector][]Vector
	for i := 0; i < 32; i++ {
		tables[i] = make(map[Vector][]Vector)
	}
	for i, g := range groups {
		t := triangles[i]
		n := t.Normal()
		for j := 0; j < 32; j++ {
			if g&1 == 1 {
				tables[j][t.V1.Position] = append(tables[j][t.V1.Position], n)
				tables[j][t.V2.Position] = append(tables[j][t.V2.Position], n)
				tables[j][t.V3.Position] = append(tables[j][t.V3.Position], n)
			}
			g >>= 1
		}
	}
	for i, g := range groups {
		t := triangles[i]
		var n1, n2, n3 Vector
		for j := 0; j < 32; j++ {
			if g&1 == 1 {
				for _, v := range tables[j][t.V1.Position] {
					n1 = n1.Add(v)
				}
				for _, v := range tables[j][t.V2.Position] {
					n2 = n2.Add(v)
				}
				for _, v := range tables[j][t.V3.Position] {
					n3 = n3.Add(v)
				}
			}
			g >>= 1
		}
		t.V1.Normal = n1.Normalize()
		t.V2.Normal = n2.Normalize()
		t.V3.Normal = n3.Normalize()
	}
	return nil
}

func readLocalAxis(file *os.File) (Matrix, error) {
	var m [4][3]float32
	if err := binary.Read(file, binary.LittleEndian, &m); err != nil {
		return Matrix{}, err
	}
	matrix := Matrix{
		float64(m[0][0]), float64(m[0][1]), float64(m[0][2]), float64(m[3][0]),
		float64(m[1][0]), float64(m[1][1]), float64(m[1][2]), float64(m[3][1]),
		float64(m[2][0]), float64(m[2][1]), float64(m[2][2]), float64(m[3][2]),
		0, 0, 0, 1,
	}
	return matrix, nil
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
