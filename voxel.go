package fauxgl

type Voxel struct {
	X, Y, Z int
	Color   Color
}

type voxelAxis int

const (
	_ voxelAxis = iota
	voxelX
	voxelY
	voxelZ
)

type voxelNormal struct {
	Axis voxelAxis
	Sign int
}

var (
	voxelPosX = voxelNormal{voxelX, 1}
	voxelNegX = voxelNormal{voxelX, -1}
	voxelPosY = voxelNormal{voxelY, 1}
	voxelNegY = voxelNormal{voxelY, -1}
	voxelPosZ = voxelNormal{voxelZ, 1}
	voxelNegZ = voxelNormal{voxelZ, -1}
)

type voxelPlane struct {
	Normal   voxelNormal
	Position int
	Color    Color
}

type voxelFace struct {
	I0, J0 int
	I1, J1 int
}

func NewVoxelMesh(voxels []Voxel) *Mesh {
	type key struct {
		X, Y, Z int
	}

	// create lookup table
	lookup := make(map[key]bool)
	for _, v := range voxels {
		lookup[key{v.X, v.Y, v.Z}] = true
	}

	// find exposed faces
	planeFaces := make(map[voxelPlane][]voxelFace)
	for _, v := range voxels {
		if !lookup[key{v.X + 1, v.Y, v.Z}] {
			plane := voxelPlane{voxelPosX, v.X, v.Color}
			face := voxelFace{v.Y, v.Z, v.Y, v.Z}
			planeFaces[plane] = append(planeFaces[plane], face)
		}
		if !lookup[key{v.X - 1, v.Y, v.Z}] {
			plane := voxelPlane{voxelNegX, v.X, v.Color}
			face := voxelFace{v.Y, v.Z, v.Y, v.Z}
			planeFaces[plane] = append(planeFaces[plane], face)
		}
		if !lookup[key{v.X, v.Y + 1, v.Z}] {
			plane := voxelPlane{voxelPosY, v.Y, v.Color}
			face := voxelFace{v.X, v.Z, v.X, v.Z}
			planeFaces[plane] = append(planeFaces[plane], face)
		}
		if !lookup[key{v.X, v.Y - 1, v.Z}] {
			plane := voxelPlane{voxelNegY, v.Y, v.Color}
			face := voxelFace{v.X, v.Z, v.X, v.Z}
			planeFaces[plane] = append(planeFaces[plane], face)
		}
		if !lookup[key{v.X, v.Y, v.Z + 1}] {
			plane := voxelPlane{voxelPosZ, v.Z, v.Color}
			face := voxelFace{v.X, v.Y, v.X, v.Y}
			planeFaces[plane] = append(planeFaces[plane], face)
		}
		if !lookup[key{v.X, v.Y, v.Z - 1}] {
			plane := voxelPlane{voxelNegZ, v.Z, v.Color}
			face := voxelFace{v.X, v.Y, v.X, v.Y}
			planeFaces[plane] = append(planeFaces[plane], face)
		}
	}

	var triangles []*Triangle
	var lines []*Line

	// find large rectangles, triangulate and outline
	for plane, faces := range planeFaces {
		faces = combineVoxelFaces(faces)
		lines = append(lines, outlineVoxelFaces(plane, faces)...)
		triangles = append(triangles, triangulateVoxelFaces(plane, faces)...)
	}

	return NewMesh(triangles, lines)
}

func combineVoxelFaces(faces []voxelFace) []voxelFace {
	// determine bounding box
	i0 := faces[0].I0
	j0 := faces[0].J0
	i1 := faces[0].I1
	j1 := faces[0].J1
	for _, f := range faces {
		if f.I0 < i0 {
			i0 = f.I0
		}
		if f.J0 < j0 {
			j0 = f.J0
		}
		if f.I1 > i1 {
			i1 = f.I1
		}
		if f.J1 > j1 {
			j1 = f.J1
		}
	}
	// create arrays
	nj := j1 - j0 + 1
	ni := i1 - i0 + 1
	a := make([][]int, nj)
	w := make([][]int, nj)
	h := make([][]int, nj)
	for j := range a {
		a[j] = make([]int, ni)
		w[j] = make([]int, ni)
		h[j] = make([]int, ni)
	}
	// populate array
	count := 0
	for _, f := range faces {
		for j := f.J0; j <= f.J1; j++ {
			for i := f.I0; i <= f.I1; i++ {
				a[j-j0][i-i0] = 1
				count++
			}
		}
	}
	// find rectangles
	var result []voxelFace
	for count > 0 {
		var maxArea int
		var maxFace voxelFace
		for j := 0; j < nj; j++ {
			for i := 0; i < ni; i++ {
				if a[j][i] == 0 {
					continue
				}
				if j == 0 {
					h[j][i] = 1
				} else {
					h[j][i] = h[j-1][i] + 1
				}
				if i == 0 {
					w[j][i] = 1
				} else {
					w[j][i] = w[j][i-1] + 1
				}
				minw := w[j][i]
				for dh := 0; dh < h[j][i]; dh++ {
					if w[j-dh][i] < minw {
						minw = w[j-dh][i]
					}
					area := (dh + 1) * minw
					if area > maxArea {
						maxArea = area
						maxFace = voxelFace{
							i0 + i - minw + 1, j0 + j - dh, i0 + i, j0 + j}
					}
				}
			}
		}
		result = append(result, maxFace)
		for j := maxFace.J0; j <= maxFace.J1; j++ {
			for i := maxFace.I0; i <= maxFace.I1; i++ {
				a[j-j0][i-i0] = 0
				count--
			}
		}
		for j := 0; j < nj; j++ {
			for i := 0; i < ni; i++ {
				w[j][i] = 0
				h[j][i] = 0
			}
		}
	}
	return result
}

func triangulateVoxelFaces(plane voxelPlane, faces []voxelFace) []*Triangle {
	triangles := make([]*Triangle, len(faces)*2)
	k := float64(plane.Position) + float64(plane.Normal.Sign)*0.5
	for i, face := range faces {
		i0 := float64(face.I0) - 0.5
		j0 := float64(face.J0) - 0.5
		i1 := float64(face.I1) + 0.5
		j1 := float64(face.J1) + 0.5
		var p1, p2, p3, p4 Vector
		switch plane.Normal.Axis {
		case voxelX:
			p1 = Vector{k, i0, j0}
			p2 = Vector{k, i1, j0}
			p3 = Vector{k, i1, j1}
			p4 = Vector{k, i0, j1}
		case voxelY:
			p1 = Vector{i0, k, j1}
			p2 = Vector{i1, k, j1}
			p3 = Vector{i1, k, j0}
			p4 = Vector{i0, k, j0}
		case voxelZ:
			p1 = Vector{i0, j0, k}
			p2 = Vector{i1, j0, k}
			p3 = Vector{i1, j1, k}
			p4 = Vector{i0, j1, k}
		}
		if plane.Normal.Sign < 0 {
			p1, p2, p3, p4 = p4, p3, p2, p1
		}
		t1 := NewTriangleForPoints(p1, p2, p3)
		t2 := NewTriangleForPoints(p1, p3, p4)
		t1.V1.Color = plane.Color
		t1.V2.Color = plane.Color
		t1.V3.Color = plane.Color
		t2.V1.Color = plane.Color
		t2.V2.Color = plane.Color
		t2.V3.Color = plane.Color
		triangles[i*2+0] = t1
		triangles[i*2+1] = t2
	}
	return triangles
}

func outlineVoxelFaces(plane voxelPlane, faces []voxelFace) []*Line {
	// determine bounding box
	i0 := faces[0].I0
	j0 := faces[0].J0
	i1 := faces[0].I1
	j1 := faces[0].J1
	for _, f := range faces {
		if f.I0 < i0 {
			i0 = f.I0
		}
		if f.J0 < j0 {
			j0 = f.J0
		}
		if f.I1 > i1 {
			i1 = f.I1
		}
		if f.J1 > j1 {
			j1 = f.J1
		}
	}
	// padding
	i0--
	j0--
	i1++
	j1++
	// create array
	nj := j1 - j0 + 1
	ni := i1 - i0 + 1
	a := make([][]bool, nj)
	for j := range a {
		a[j] = make([]bool, ni)
	}
	// populate array
	for _, f := range faces {
		for j := f.J0; j <= f.J1; j++ {
			for i := f.I0; i <= f.I1; i++ {
				a[j-j0][i-i0] = true
			}
		}
	}
	var lines []*Line
	for sign := -1; sign <= 1; sign += 2 {
		// find "horizontal" lines
		for j := 1; j < nj-1; j++ {
			start := -1
			for i := 0; i < ni; i++ {
				if a[j][i] && !a[j+sign][i] {
					if start < 0 {
						start = i
					}
				} else if start >= 0 {
					end := i - 1
					ai := float64(i0+start) - 0.5
					bi := float64(i0+end) + 0.5
					jj := float64(j0+j) + 0.5*float64(sign)
					line := createVoxelOutline(plane, ai, jj, bi, jj)
					lines = append(lines, line)
					start = -1
				}
			}

		}
		// find "vertical" lines
		for i := 1; i < ni-1; i++ {
			start := -1
			for j := 0; j < nj; j++ {
				if a[j][i] && !a[j][i+sign] {
					if start < 0 {
						start = j
					}
				} else if start >= 0 {
					end := j - 1
					aj := float64(j0+start) - 0.5
					bj := float64(j0+end) + 0.5
					ii := float64(i0+i) + 0.5*float64(sign)
					line := createVoxelOutline(plane, ii, aj, ii, bj)
					lines = append(lines, line)
					start = -1
				}
			}
		}
	}
	return lines
}

func createVoxelOutline(plane voxelPlane, i0, j0, i1, j1 float64) *Line {
	k := float64(plane.Position) + float64(plane.Normal.Sign)*0.5
	var p1, p2 Vector
	switch plane.Normal.Axis {
	case voxelX:
		p1 = Vector{k, i0, j0}
		p2 = Vector{k, i1, j1}
	case voxelY:
		p1 = Vector{i0, k, j0}
		p2 = Vector{i1, k, j1}
	case voxelZ:
		p1 = Vector{i0, j0, k}
		p2 = Vector{i1, j1, k}
	}
	return NewLineForPoints(p1, p2)
}
