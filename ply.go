package fauxgl

import (
	"bufio"
	"encoding/binary"
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

	// read header
	reader := bufio.NewReader(file)
	var element plyElement
	var elements []plyElement
	format := plyAscii
	bytes := 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		bytes += len(line)
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

	file.Seek(int64(bytes), 0)

	switch format {
	case plyBinaryBigEndian:
		return loadPlyBinary(file, elements, binary.BigEndian)
	case plyBinaryLittleEndian:
		return loadPlyBinary(file, elements, binary.LittleEndian)
	default:
		return loadPlyAscii(file, elements)
	}
}

func loadPlyAscii(file *os.File, elements []plyElement) (*Mesh, error) {
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
					t := Triangle{}
					t.V1.Position = vertexes[i1]
					t.V2.Position = vertexes[i2]
					t.V3.Position = vertexes[i3]
					t.FixNormals()
					triangles = append(triangles, &t)
					fi += 3
				}
				fi++
			}
			if element.name == "vertex" {
				vertexes = append(vertexes, vertex)
			}
		}
	}
	return NewTriangleMesh(triangles), nil
}

func loadPlyBinary(file *os.File, elements []plyElement, order binary.ByteOrder) (*Mesh, error) {
	var vertexes []Vector
	var triangles []*Triangle
	for _, element := range elements {
		for i := 0; i < element.count; i++ {
			var vertex Vector
			var points []Vector
			for _, property := range element.properties {
				if property.countType == plyNone {
					value, err := readPlyFloat(file, order, property.dataType)
					if err != nil {
						return nil, err
					}
					if property.name == "x" {
						vertex.X = value
					}
					if property.name == "y" {
						vertex.Y = value
					}
					if property.name == "z" {
						vertex.Z = value
					}
				} else {
					count, err := readPlyInt(file, order, property.countType)
					if err != nil {
						return nil, err
					}
					for j := 0; j < count; j++ {
						value, err := readPlyInt(file, order, property.dataType)
						if err != nil {
							return nil, err
						}
						if property.name == "vertex_indices" {
							points = append(points, vertexes[value])
						}
					}
				}
			}
			if element.name == "vertex" {
				vertexes = append(vertexes, vertex)
			}
			if element.name == "face" {
				t := Triangle{}
				t.V1.Position = points[0]
				t.V2.Position = points[1]
				t.V3.Position = points[2]
				t.FixNormals()
				triangles = append(triangles, &t)
			}
		}
	}
	return NewTriangleMesh(triangles), nil
}

func readPlyInt(file *os.File, order binary.ByteOrder, dataType plyDataType) (int, error) {
	value, err := readPlyFloat(file, order, dataType)
	return int(value), err
}

func readPlyFloat(file *os.File, order binary.ByteOrder, dataType plyDataType) (float64, error) {
	switch dataType {
	case plyInt8:
		var value int8
		err := binary.Read(file, order, &value)
		return float64(value), err
	case plyUint8:
		var value uint8
		err := binary.Read(file, order, &value)
		return float64(value), err
	case plyInt16:
		var value int16
		err := binary.Read(file, order, &value)
		return float64(value), err
	case plyUint16:
		var value uint16
		err := binary.Read(file, order, &value)
		return float64(value), err
	case plyInt32:
		var value int32
		err := binary.Read(file, order, &value)
		return float64(value), err
	case plyUint32:
		var value uint32
		err := binary.Read(file, order, &value)
		return float64(value), err
	case plyFloat32:
		var value float32
		err := binary.Read(file, order, &value)
		return float64(value), err
	case plyFloat64:
		var value float64
		err := binary.Read(file, order, &value)
		return float64(value), err
	default:
		return 0, nil
	}
}
