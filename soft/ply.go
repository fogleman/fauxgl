package soft

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type plyFormat int

const (
	_ plyFormat = iota
	plyAscii
	plyBinaryLittleEndian
	plyBinaryBigEndian
)

var plyFormatMapping = map[string]plyFormat{
	"ascii":                plyAscii,
	"binary_little_endian": plyBinaryLittleEndian,
	"binary_big_endian":    plyBinaryBigEndian,
}

type plyDataType int

const (
	plyNone plyDataType = iota
	plyInt8
	plyUint8
	plyInt16
	plyUint16
	plyInt32
	plyUint32
	plyFloat32
	plyFloat64
)

var plyDataTypeMapping = map[string]plyDataType{
	"char":    plyInt8,
	"uchar":   plyUint8,
	"short":   plyInt16,
	"ushort":  plyUint16,
	"int":     plyInt32,
	"uint":    plyUint32,
	"float":   plyFloat32,
	"double":  plyFloat64,
	"int8":    plyInt8,
	"uint8":   plyUint8,
	"int16":   plyInt16,
	"uint16":  plyUint16,
	"int32":   plyInt32,
	"uint32":  plyUint32,
	"float32": plyFloat32,
	"float64": plyFloat64,
}

type plyProperty struct {
	name      string
	countType plyDataType
	dataType  plyDataType
}

type plyElement struct {
	name       string
	count      int
	properties []plyProperty
}

func LoadPLY(path string) (*Mesh, error) {
	// open file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// read lines
	scanner := bufio.NewScanner(file)
	var element plyElement
	var elements []plyElement
	format := plyAscii
	bytes := 0
	for scanner.Scan() {
		line := scanner.Text()
		bytes += len(line) + 1
		f := strings.Fields(line)
		if len(f) == 0 {
			continue
		}
		if f[0] == "format" {
			format = plyFormatMapping[f[1]]
		}
		if f[0] == "element" {
			if element.count > 0 {
				elements = append(elements, element)
			}
			name := f[1]
			count, _ := strconv.ParseInt(f[2], 0, 0)
			element = plyElement{name, int(count), nil}
		}
		if f[0] == "property" {
			if f[1] == "list" {
				countType := plyDataTypeMapping[f[2]]
				dataType := plyDataTypeMapping[f[3]]
				name := f[4]
				property := plyProperty{name, countType, dataType}
				element.properties = append(element.properties, property)
			} else {
				countType := plyNone
				dataType := plyDataTypeMapping[f[1]]
				name := f[2]
				property := plyProperty{name, countType, dataType}
				element.properties = append(element.properties, property)
			}
		}
		if f[0] == "end_header" {
			if element.count > 0 {
				elements = append(elements, element)
			}
			break
		}
	}

	// check for errors
	if err = scanner.Err(); err != nil {
		return nil, err
	}

	file.Seek(int64(bytes), 0)

	if format != plyAscii {
		panic("only ascii ply files are supported")
	}

	return readFormatAscii(file, elements)
}

func readFormatAscii(file *os.File, elements []plyElement) (*Mesh, error) {
	scanner := bufio.NewScanner(file)
	var vertexes []Vector
	var triangles []*Triangle
	for _, element := range elements {
		for i := 0; i < element.count; i++ {
			scanner.Scan()
			line := scanner.Text()
			f := strings.Fields(line)
			fi := 0
			vertex := Vector{}
			for _, property := range element.properties {
				if property.name == "x" {
					vertex.X, _ = strconv.ParseFloat(f[fi], 64)
				}
				if property.name == "y" {
					vertex.Y, _ = strconv.ParseFloat(f[fi], 64)
				}
				if property.name == "z" {
					vertex.Z, _ = strconv.ParseFloat(f[fi], 64)
				}
				if property.name == "vertex_indices" {
					i1, _ := strconv.ParseInt(f[fi+1], 0, 0)
					i2, _ := strconv.ParseInt(f[fi+2], 0, 0)
					i3, _ := strconv.ParseInt(f[fi+3], 0, 0)
					v1 := vertexes[i1]
					v2 := vertexes[i2]
					v3 := vertexes[i3]
					t := NewTriangle(v1, v2, v3)
					triangles = append(triangles, t)
					fi += 3
				}
				fi++
			}
			if element.name == "vertex" {
				vertexes = append(vertexes, vertex)
			}
		}
	}
	return NewMesh(triangles), nil
}
