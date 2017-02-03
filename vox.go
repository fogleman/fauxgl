package fauxgl

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
)

type VOXHeader struct {
	Magic   [4]byte
	Version int32
}

type VOXChunk struct {
	ID            [4]byte
	ContentBytes  int32
	ChildrenBytes int32
}

type VOXSize struct {
	X, Y, Z int32
}

type VOXVoxel struct {
	X, Y, Z, I uint8
}

func LoadVOX(path string) (*Mesh, error) {
	// open file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// read and check header
	header := VOXHeader{}
	if err := binary.Read(file, binary.LittleEndian, &header); err != nil {
		return nil, err
	}
	if string(header.Magic[:]) != "VOX " {
		return nil, errors.New("invalid vox header")
	}
	if header.Version != 150 {
		return nil, errors.New("unsupported vox version")
	}

	var voxels []VOXVoxel
	var palette [256]Color

	for {
		// read chunk header
		chunk := VOXChunk{}
		if err := binary.Read(file, binary.LittleEndian, &chunk); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return nil, err
		}

		id := string(chunk.ID[:])
		switch id {
		case "SIZE":
			size := VOXSize{}
			if err := binary.Read(file, binary.LittleEndian, &size); err != nil {
				return nil, err
			}
		case "XYZI":
			var numVoxels uint32
			if err := binary.Read(file, binary.LittleEndian, &numVoxels); err != nil {
				return nil, err
			}
			for i := 0; i < int(numVoxels); i++ {
				voxel := VOXVoxel{}
				if err := binary.Read(file, binary.LittleEndian, &voxel); err != nil {
					return nil, err
				}
				voxels = append(voxels, voxel)
			}
		case "RGBA":
			for i := 0; i <= 254; i++ {
				var color [4]uint8
				if err := binary.Read(file, binary.LittleEndian, &color); err != nil {
					return nil, err
				}
				r := float64(color[0]) / 255
				g := float64(color[1]) / 255
				b := float64(color[2]) / 255
				a := float64(color[3]) / 255
				palette[i+1] = Color{r, g, b, a}
			}
		default:
			file.Seek(int64(chunk.ContentBytes), 1)
		}
	}

	mesh := NewMesh(nil)
	for _, v := range voxels {
		x := float64(v.X) + 0.5
		y := float64(v.Y) + 0.5
		z := float64(v.Z) + 0.5
		c := palette[v.I]
		cube := NewCube()
		cube.Transform(Translate(Vector{x, y, z}))
		for _, t := range cube.Triangles {
			t.V1.Color = c
			t.V2.Color = c
			t.V3.Color = c
		}
		mesh.Add(cube)
	}

	return mesh, nil
}
