package fauxgl

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func parseIndex(value string, length int) int {
	parsed, _ := strconv.ParseInt(value, 0, 0)
	n := int(parsed)
	if n < 0 {
		n += length
	}
	return n
}

func LoadOBJ(path string) (*Mesh, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return LoadOBJReader(file)
}

func LoadOBJReader(r io.Reader) (*Mesh, error) {
	vs := make([]Vector, 1, 1024)  // 1-based indexing
	vts := make([]Vector, 1, 1024) // 1-based indexing
	vns := make([]Vector, 1, 1024) // 1-based indexing
	var triangles []*Triangle
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		keyword := fields[0]
		args := fields[1:]
		switch keyword {
		case "v":
			f := ParseFloats(args)
			v := Vector{f[0], f[1], f[2]}
			vs = append(vs, v)
		case "vt":
			f := ParseFloats(args)
			v := Vector{f[0], f[1], 0}
			vts = append(vts, v)
		case "vn":
			f := ParseFloats(args)
			v := Vector{f[0], f[1], f[2]}
			vns = append(vns, v)
		case "f":
			fvs := make([]int, len(args))
			fvts := make([]int, len(args))
			fvns := make([]int, len(args))
			for i, arg := range args {
				vertex := strings.Split(arg+"//", "/")
				fvs[i] = parseIndex(vertex[0], len(vs))
				fvts[i] = parseIndex(vertex[1], len(vts))
				fvns[i] = parseIndex(vertex[2], len(vns))
			}
			for i := 1; i < len(fvs)-1; i++ {
				i1, i2, i3 := 0, i, i+1
				t := Triangle{}
				if err := setvecs(
					vs,
					fvs,
					i1, i2, i3,
					&t.V1.Position, &t.V2.Position, &t.V3.Position,
					"triangle position",
				); err != nil {
					return nil, err
				}

				if err := setvecs(
					vs,
					fvns,
					i1, i2, i3,
					&t.V1.Normal, &t.V2.Normal, &t.V3.Normal,
					"triangle normal",
				); err != nil {
					return nil, err
				}

				if err := setvecs(
					vs,
					fvts,
					i1, i2, i3,
					&t.V1.Texture, &t.V2.Texture, &t.V3.Texture,
					"triangle texture",
				); err != nil {
					return nil, err
				}

				t.FixNormals()
				triangles = append(triangles, &t)
			}
		}
	}
	return NewTriangleMesh(triangles), scanner.Err()
}

func setvecs(
	s []Vector,
	indexArr []int,
	idx1, idx2, idx3 int,
	v1, v2, v3 *Vector,
	desc string,
) (error) {
	var err error
	*v1, err = safeindexinner(s, indexArr, idx1, desc)
	if err != nil {
		return err
	}

	*v2, err = safeindexinner(s, indexArr, idx2, desc)
	if err != nil {
		return err
	}

	*v3, err = safeindexinner(s, indexArr, idx3, desc)
	return err
}

func safeindexinner(
	s []Vector,
	indexArr []int,
	idx int,
	desc string,
) (Vector, error) {
	if idx < 0 || idx >= len(indexArr) {
		return Vector{}, fmt.Errorf("found out of bounds %s index: %d (len=%d)", desc, idx, len(indexArr))
	}

	vIdx := indexArr[idx]
	if vIdx < 0 || vIdx >= len(s) {
		return Vector{}, fmt.Errorf("found out of bounds %s vector index: %d (len=%d)", desc, vIdx, len(s))
	}

	return s[vIdx], nil
}
